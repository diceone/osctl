package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// detectOSFamily returns the package-manager family for this system:
// "debian" (apt), "rhel" (dnf/yum), "suse" (zypper) or "" when unknown.
func detectOSFamily() string {
	if data, err := os.ReadFile("/etc/os-release"); err == nil {
		release := strings.ToLower(string(data))
		switch {
		case strings.Contains(release, "ubuntu") || strings.Contains(release, "debian"):
			return "debian"
		case strings.Contains(release, "rhel") || strings.Contains(release, "centos") ||
			strings.Contains(release, "fedora") || strings.Contains(release, "rocky") ||
			strings.Contains(release, "almalinux") || strings.Contains(release, "amzn"):
			return "rhel"
		case strings.Contains(release, "suse") || strings.Contains(release, "opensuse"):
			return "suse"
		}
	}
	// Legacy fallbacks when /etc/os-release is missing.
	if _, err := os.Stat("/etc/debian_version"); err == nil {
		return "debian"
	}
	if _, err := os.Stat("/etc/redhat-release"); err == nil {
		return "rhel"
	}
	return ""
}

// listPackageUpdates lists available package updates without installing them.
func listPackageUpdates() string {
	family := detectOSFamily()
	switch family {
	case "debian":
		// The package index must be current for an accurate list; only root
		// may refresh it, so skip the refresh silently when unprivileged.
		if os.Geteuid() == 0 {
			if msg := runPackageCmd(exec.Command("apt-get", "update", "-qq")); msg != "" {
				return msg
			}
		}
		out, err := exec.Command("apt", "list", "--upgradable").CombinedOutput()
		if err != nil {
			return fmt.Sprintf("Failed to list package updates. Error: %v\n%s", err, string(out))
		}
		return summarizeUpdates("debian", string(out))
	case "rhel":
		checker := "dnf"
		if _, err := exec.LookPath("dnf"); err != nil {
			checker = "yum"
		}
		out, err := exec.Command(checker, "check-update", "--quiet").CombinedOutput()
		if err != nil {
			// dnf/yum exit 100 when updates are available; that is not a failure.
			if code, ok := exitCode(err); ok && code == 100 {
				return summarizeUpdates("rhel", string(out))
			}
			msg := fmt.Sprintf("Failed to list package updates. Error: %v\n%s", err, string(out))
			if hint := permissionHint(err, string(out)); hint != "" {
				msg += "\n" + hint
			}
			return msg
		}
		return summarizeUpdates("rhel", string(out))
	case "suse":
		out, err := exec.Command("zypper", "--non-interactive", "list-updates").CombinedOutput()
		if err != nil {
			msg := fmt.Sprintf("Failed to list package updates. Error: %v\n%s", err, string(out))
			if hint := permissionHint(err, string(out)); hint != "" {
				msg += "\n" + hint
			}
			return msg
		}
		return summarizeUpdates("suse", string(out))
	default:
		return "Unsupported OS for listing package updates"
	}
}

// exitCode returns the process exit code when err wraps an *exec.ExitError.
func exitCode(err error) (int, bool) {
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode(), true
	}
	return 0, false
}

// summarizeUpdates extracts the package rows from raw command output of the
// given package-manager family and formats the read-only update list.
func summarizeUpdates(family, raw string) string {
	var rows []string
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimRight(line, "\r")
		switch family {
		case "debian":
			// "pkg/current version,now newer version arch [upgradable from: ...]"
			if strings.Contains(line, "[upgradable") {
				rows = append(rows, line)
			}
		case "suse":
			// zypper table rows: "v | repo | name | current | available | ..."
			if strings.HasPrefix(line, "v |") || strings.HasPrefix(line, "u |") {
				rows = append(rows, line)
			}
		default: // rhel: quiet output contains only package rows
			if strings.TrimSpace(line) != "" {
				rows = append(rows, strings.TrimSpace(line))
			}
		}
	}

	if len(rows) == 0 {
		return "No package updates available."
	}

	var output strings.Builder
	output.WriteString(fmt.Sprintf("Available package updates: %d\n\n", len(rows)))
	for _, row := range rows {
		output.WriteString("  " + row + "\n")
	}
	return output.String()
}
