package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// rootHint tells the user how to gain the privileges a command needs.
const rootHint = "Hint: this command requires root privileges. Try running it with sudo."

// permissionHint returns rootHint when the error (or captured command output)
// indicates a permission problem and the process is not running as root.
func permissionHint(err error, output string) string {
	if err == nil || os.Geteuid() == 0 {
		return ""
	}
	msg := strings.ToLower(fmt.Sprint(err) + " " + output)
	if errors.Is(err, os.ErrPermission) ||
		strings.Contains(msg, "permission denied") ||
		strings.Contains(msg, "access denied") ||
		strings.Contains(msg, "operation not permitted") ||
		strings.Contains(msg, "authentication required") {
		return rootHint
	}
	return ""
}

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
	found := false
	incomplete := false
	for _, dir := range criticalDirs {
		cmd := exec.Command("find", dir, "-type", "f", "-perm", "-002", "-ls")
		out, err := cmd.Output() // stderr discarded, permission errors are expected
		if err != nil {
			// find exits non-zero when it cannot read every directory, so
			// results are partial rather than a clean "nothing found".
			incomplete = true
		}
		if listing := strings.TrimSpace(string(out)); listing != "" {
			found = true
			output.WriteString(fmt.Sprintf("\nIn %s:\n%s\n", dir, listing))
		}
	}
	if !found {
		if incomplete {
			output.WriteString("(none found; scan incomplete — permission denied on some directories. Try running with sudo)\n")
		} else {
			output.WriteString("(none found)\n")
		}
	} else if incomplete {
		output.WriteString("\nNote: scan incomplete — permission denied on some directories. Try running with sudo for a full scan.\n")
	}

	// Check for SUID/SGID files
	output.WriteString("\n\nSUID/SGID files (may be security risk):\n")
	cmd := exec.Command("find", "/", "-type", "f", "(", "-perm", "-4000", "-o", "-perm", "-2000", ")", "-ls")
	out, err := cmd.Output() // PermissionError noise on stderr is expected
	scanIncomplete := err != nil
	if listing := strings.TrimSpace(string(out)); listing != "" {
		lines := strings.Split(listing, "\n")
		// Limit output to first 50 lines
		if len(lines) > 50 {
			output.WriteString(strings.Join(lines[:50], "\n"))
			output.WriteString(fmt.Sprintf("\n... (%d more files)", len(lines)-50))
		} else {
			output.WriteString(listing + "\n")
		}
		if scanIncomplete {
			output.WriteString("\nNote: scan incomplete — permission denied on some directories. Try running with sudo for a full scan.\n")
		}
	} else if scanIncomplete {
		output.WriteString("(none found; scan incomplete — permission denied on some directories. Try running with sudo)\n")
	} else {
		output.WriteString("(none found)\n")
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
			line := fmt.Sprintf("❌ %s: Not found or not accessible", file)
			if hint := permissionHint(err, ""); hint != "" {
				line += " — " + hint
			}
			output.WriteString(line + "\n")
			continue
		}

		mode := info.Mode().Perm()
		output.WriteString(fmt.Sprintf("📄 %s: %03o (expected: %s)\n", file, mode, expectedPerm))
	}

	return output.String()
}

// checkUnusedUsers finds users that haven't logged in recently
func checkUnusedUsers() string {
	var output strings.Builder
	output.WriteString("User Account Audit:\n\n")

	// Get list of users with login shells, along with their UIDs
	cmd := exec.Command("sh", "-c", `awk -F: '$7 !~ /nologin|false/ {print $1 ":" $3}' /etc/passwd`)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Sprintf("Failed to get user list. Error: %v", err)
	}

	users := strings.Split(strings.TrimSpace(string(out)), "\n")
	output.WriteString(fmt.Sprintf("Users with login shells: %d\n\n", len(users)))

	for _, entry := range users {
		if entry == "" {
			continue
		}

		name, uid := entry, ""
		if idx := strings.LastIndex(entry, ":"); idx >= 0 {
			name, uid = entry[:idx], entry[idx+1:]
		}

		output.WriteString(fmt.Sprintf("User: %s (uid: %s)\n", name, uid))

		// Check last login
		cmd := exec.Command("lastlog", "-u", name)
		lastOut, err := cmd.CombinedOutput()
		if err == nil {
			output.WriteString(string(lastOut) + "\n")
		}
	}

	return output.String()
}

// getSecurityAuditSummary provides a comprehensive security audit
func getSecurityAuditSummary() string {
	var output strings.Builder
	output.WriteString("=== SECURITY AUDIT SUMMARY ===\n\n")

	// Ports
	cmd := exec.Command("ss", "-tulpn")
	portOut, _ := cmd.Output()
	portCount := strings.Count(string(portOut), "LISTEN")
	output.WriteString("Ports:\n")
	output.WriteString(fmt.Sprintf("  Open listening ports: %d\n", portCount))

	// Users: failed login attempts (Debian/Ubuntu: auth.log, RHEL: secure)
	output.WriteString("\nUsers:\n")
	authLog := ""
	for _, candidate := range []string{"/var/log/auth.log", "/var/log/secure"} {
		if _, err := os.Stat(candidate); err == nil {
			authLog = candidate
			break
		}
	}
	if authLog == "" {
		output.WriteString("  Failed login attempts: auth log not found\n")
	} else if f, err := os.Open(authLog); err != nil {
		// An unreadable log would silently report 0 failed attempts.
		output.WriteString(fmt.Sprintf("  Failed login attempts (%s): unreadable (%v)\n", authLog, err))
		if hint := permissionHint(err, ""); hint != "" {
			output.WriteString("  " + hint + "\n")
		}
	} else {
		f.Close()
		// authLog is chosen from a fixed list above, not user input.
		cmd = exec.Command("sh", "-c", fmt.Sprintf("grep 'Failed password' %s 2>/dev/null | wc -l", authLog))
		failedOut, _ := cmd.Output()
		output.WriteString(fmt.Sprintf("  Failed login attempts (%s): %s", authLog, string(failedOut)))
	}

	// Files: SUID files
	output.WriteString("\nFiles:\n")
	cmd = exec.Command("find", "/", "-type", "f", "-perm", "-4000")
	suidOut, _ := cmd.Output() // PermissionError noise on stderr is expected
	suidCount := 0
	trimmed := strings.TrimSpace(string(suidOut))
	if trimmed != "" {
		suidCount = len(strings.Split(trimmed, "\n"))
	}
	output.WriteString(fmt.Sprintf("  SUID files found: %d\n", suidCount))

	// System: firewall, SELinux, package updates
	output.WriteString("\nSystem:\n")
	cmd = exec.Command("systemctl", "is-active", "firewalld")
	firewallOut, _ := cmd.CombinedOutput()
	firewallStatus := strings.TrimSpace(string(firewallOut))
	if firewallStatus == "active" {
		output.WriteString("  ✅ Firewall: Active\n")
	} else {
		output.WriteString("  ⚠️  Firewall: Inactive or not available\n")
	}

	// Check SELinux status
	cmd = exec.Command("getenforce")
	selinuxOut, _ := cmd.CombinedOutput()
	selinuxStatus := strings.TrimSpace(string(selinuxOut))
	if selinuxStatus == "Enforcing" {
		output.WriteString("  ✅ SELinux: Enforcing\n")
	} else if selinuxStatus == "Permissive" {
		output.WriteString("  ⚠️  SELinux: Permissive\n")
	} else {
		output.WriteString("  ❌ SELinux: Disabled or not available\n")
	}

	// Check for available updates
	if _, err := os.Stat("/etc/redhat-release"); err == nil {
		cmd = exec.Command("yum", "check-update", "--quiet")
		updateOut, _ := cmd.CombinedOutput()
		updateCount := len(strings.Split(strings.TrimSpace(string(updateOut)), "\n"))
		output.WriteString(fmt.Sprintf("  Available package updates: ~%d\n", updateCount))
	} else if _, err := os.Stat("/etc/debian_version"); err == nil {
		cmd = exec.Command("apt", "list", "--upgradable")
		updateOut, _ := cmd.CombinedOutput()
		updateCount := strings.Count(string(updateOut), "[upgradable")
		output.WriteString(fmt.Sprintf("  Available package updates: %d\n", updateCount))
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
		msg := fmt.Sprintf("Failed to read SSH config. Error: %v", err)
		if hint := permissionHint(err, ""); hint != "" {
			msg += "\n" + hint
		}
		return msg
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

	output.WriteString("Legend: ✅ explicitly set to the recommended value, ⚠️ explicitly set to another value, ❓ not explicitly set (the sshd default applies)\n\n")

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
			output.WriteString(fmt.Sprintf("❓ %s: not explicitly set — the sshd default applies (recommended: %s)\n", setting, recommended))
		}
	}

	return output.String()
}
