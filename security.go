package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// getOpenPorts scans for open listening ports
func getOpenPorts() string {
	cmd := exec.Command("ss", "-tulpn")
	out, err := cmd.CombinedOutput()
	if err != nil {
		// Fallback to netstat if ss is not available
		cmd = exec.Command("netstat", "-tulpn")
		out, err = cmd.CombinedOutput()
		if err != nil {
			return fmt.Sprintf("Failed to get open ports. Error: %v", err)
		}
	}

	var output strings.Builder
	output.WriteString("Open Listening Ports:\n\n")
	output.WriteString(string(out))
	return output.String()
}

// checkSuspiciousFiles checks for files with suspicious permissions
func checkSuspiciousFiles() string {
	var output strings.Builder
	output.WriteString("Security Audit - Suspicious File Permissions:\n\n")

	// Check for world-writable files in critical directories
	criticalDirs := []string{"/etc", "/usr/bin", "/usr/local/bin", "/bin", "/sbin"}

	output.WriteString("World-writable files in critical directories:\n")
	for _, dir := range criticalDirs {
		cmd := exec.Command("find", dir, "-type", "f", "-perm", "-002", "-ls")
		out, err := cmd.CombinedOutput()
		if err == nil && len(out) > 0 {
			output.WriteString(fmt.Sprintf("\nIn %s:\n", dir))
			output.WriteString(string(out))
		}
	}

	// Check for SUID/SGID files
	output.WriteString("\n\nSUID/SGID files (may be security risk):\n")
	cmd := exec.Command("find", "/", "-type", "f", "(", "-perm", "-4000", "-o", "-perm", "-2000", ")", "-ls")
	out, err := cmd.CombinedOutput()
	if err == nil {
		lines := strings.Split(string(out), "\n")
		// Limit output to first 50 lines
		if len(lines) > 50 {
			output.WriteString(strings.Join(lines[:50], "\n"))
			output.WriteString(fmt.Sprintf("\n... (%d more files)", len(lines)-50))
		} else {
			output.WriteString(string(out))
		}
	}

	return output.String()
}

// checkFilePermissions checks permissions of critical system files
func checkFilePermissions() string {
	var output strings.Builder
	output.WriteString("Critical File Permissions Check:\n\n")

	criticalFiles := map[string]string{
		"/etc/passwd":          "644",
		"/etc/shadow":          "000 or 400",
		"/etc/group":           "644",
		"/etc/gshadow":         "000 or 400",
		"/etc/ssh/sshd_config": "600",
	}

	for file, expectedPerm := range criticalFiles {
		info, err := os.Stat(file)
		if err != nil {
			output.WriteString(fmt.Sprintf("❌ %s: Not found or not accessible\n", file))
			continue
		}

		mode := info.Mode().Perm()
		output.WriteString(fmt.Sprintf("📄 %s: %04o (expected: %s)\n", file, mode, expectedPerm))
	}

	return output.String()
}

// checkUnusedUsers finds users that haven't logged in recently
func checkUnusedUsers() string {
	var output strings.Builder
	output.WriteString("User Account Audit:\n\n")

	// Get list of users with login shells
	cmd := exec.Command("sh", "-c", "awk -F: '$7 !~ /nologin|false/ {print $1}' /etc/passwd")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Sprintf("Failed to get user list. Error: %v", err)
	}

	users := strings.Split(strings.TrimSpace(string(out)), "\n")
	output.WriteString(fmt.Sprintf("Users with login shells: %d\n\n", len(users)))

	for _, user := range users {
		if user == "" {
			continue
		}

		// Check last login
		cmd := exec.Command("lastlog", "-u", user)
		lastOut, err := cmd.CombinedOutput()
		if err == nil {
			output.WriteString(fmt.Sprintf("User: %s\n%s\n", user, string(lastOut)))
		}
	}

	return output.String()
}

// getSecurityAuditSummary provides a comprehensive security audit
func getSecurityAuditSummary() string {
	var output strings.Builder
	output.WriteString("=== SECURITY AUDIT SUMMARY ===\n\n")

	// Count open ports
	cmd := exec.Command("ss", "-tulpn")
	portOut, _ := cmd.CombinedOutput()
	portCount := strings.Count(string(portOut), "LISTEN")
	output.WriteString(fmt.Sprintf("Open listening ports: %d\n", portCount))

	// Check for failed login attempts
	cmd = exec.Command("sh", "-c", "grep 'Failed password' /var/log/auth.log 2>/dev/null | wc -l")
	failedOut, _ := cmd.CombinedOutput()
	output.WriteString(fmt.Sprintf("Failed login attempts (auth.log): %s", string(failedOut)))

	// Check for SUID files
	cmd = exec.Command("find", "/", "-type", "f", "-perm", "-4000", "2>/dev/null")
	suidOut, _ := cmd.CombinedOutput()
	suidCount := len(strings.Split(strings.TrimSpace(string(suidOut)), "\n"))
	output.WriteString(fmt.Sprintf("SUID files found: %d\n", suidCount))

	// Check firewall status
	cmd = exec.Command("systemctl", "is-active", "firewalld")
	firewallOut, _ := cmd.CombinedOutput()
	firewallStatus := strings.TrimSpace(string(firewallOut))
	if firewallStatus == "active" {
		output.WriteString("✅ Firewall: Active\n")
	} else {
		output.WriteString("⚠️  Firewall: Inactive or not available\n")
	}

	// Check SELinux status
	cmd = exec.Command("getenforce")
	selinuxOut, _ := cmd.CombinedOutput()
	selinuxStatus := strings.TrimSpace(string(selinuxOut))
	if selinuxStatus == "Enforcing" {
		output.WriteString("✅ SELinux: Enforcing\n")
	} else if selinuxStatus == "Permissive" {
		output.WriteString("⚠️  SELinux: Permissive\n")
	} else {
		output.WriteString("❌ SELinux: Disabled or not available\n")
	}

	// Check for available updates
	output.WriteString("\n")
	if _, err := os.Stat("/etc/redhat-release"); err == nil {
		cmd = exec.Command("yum", "check-update", "--quiet")
		updateOut, _ := cmd.CombinedOutput()
		updateCount := len(strings.Split(strings.TrimSpace(string(updateOut)), "\n"))
		output.WriteString(fmt.Sprintf("Available package updates: ~%d\n", updateCount))
	} else if _, err := os.Stat("/etc/debian_version"); err == nil {
		cmd = exec.Command("apt", "list", "--upgradable")
		updateOut, _ := cmd.CombinedOutput()
		updateCount := strings.Count(string(updateOut), "[upgradable")
		output.WriteString(fmt.Sprintf("Available package updates: %d\n", updateCount))
	}

	return output.String()
}

// checkSSHSecurity audits SSH configuration
func checkSSHSecurity() string {
	var output strings.Builder
	output.WriteString("SSH Security Configuration:\n\n")

	sshConfigFile := "/etc/ssh/sshd_config"
	content, err := os.ReadFile(sshConfigFile)
	if err != nil {
		return fmt.Sprintf("Failed to read SSH config. Error: %v", err)
	}

	config := string(content)
	lines := strings.Split(config, "\n")

	checks := map[string]string{
		"PermitRootLogin":        "no",
		"PasswordAuthentication": "no",
		"PubkeyAuthentication":   "yes",
		"PermitEmptyPasswords":   "no",
		"X11Forwarding":          "no",
	}

	for setting, recommended := range checks {
		found := false
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "#") {
				continue
			}
			if strings.HasPrefix(line, setting) {
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					value := parts[1]
					if value == recommended {
						output.WriteString(fmt.Sprintf("✅ %s: %s (secure)\n", setting, value))
					} else {
						output.WriteString(fmt.Sprintf("⚠️  %s: %s (recommended: %s)\n", setting, value, recommended))
					}
					found = true
					break
				}
			}
		}
		if !found {
			output.WriteString(fmt.Sprintf("❓ %s: not explicitly set (recommended: %s)\n", setting, recommended))
		}
	}

	return output.String()
}
