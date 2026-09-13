package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// sysctlCheck describes one kernel hardening setting.
type sysctlCheck struct {
	name     string // full sysctl name, e.g. kernel.randomize_va_space
	check    func(v int64) bool
	okDesc   string
	warnDesc string
	infoOnly bool // downgrade failures to informational notes
}

// sysctlChecks covers commonly recommended kernel hardening settings.
var sysctlChecks = []sysctlCheck{
	{"kernel.randomize_va_space", func(v int64) bool { return v == 2 }, "ASLR fully enabled", "ASLR not fully enabled (recommended: 2)", false},
	{"kernel.kptr_restrict", func(v int64) bool { return v >= 1 }, "kernel pointers hidden from unprivileged users", "kernel pointers visible to unprivileged users (recommended: >= 1)", false},
	{"kernel.dmesg_restrict", func(v int64) bool { return v == 1 }, "dmesg restricted to privileged users", "dmesg readable by unprivileged users (recommended: 1)", false},
	{"fs.protected_hardlinks", func(v int64) bool { return v == 1 }, "hardlink protection enabled", "hardlink protection disabled (recommended: 1)", false},
	{"fs.protected_symlinks", func(v int64) bool { return v == 1 }, "symlink protection enabled", "symlink protection disabled (recommended: 1)", false},
	{"fs.suid_dumpable", func(v int64) bool { return v == 0 }, "suid dumps disabled", "core dumps of suid binaries allowed (recommended: 0)", false},
	{"net.ipv4.conf.all.rp_filter", func(v int64) bool { return v == 1 }, "reverse path filtering enabled", "reverse path filtering disabled (recommended: 1)", false},
	{"net.ipv4.conf.all.accept_redirects", func(v int64) bool { return v == 0 }, "ICMP redirects ignored", "ICMP redirects accepted (recommended: 0)", false},
	{"net.ipv4.conf.all.send_redirects", func(v int64) bool { return v == 0 }, "ICMP redirects not sent", "ICMP redirects sent (recommended: 0)", false},
	{"net.ipv4.ip_forward", func(v int64) bool { return v == 0 }, "IP forwarding disabled", "IP forwarding enabled (expected on routers, gateways and Docker hosts)",
		true},
}

// checkSysctlHardening audits kernel hardening settings via /proc/sys.
func checkSysctlHardening() string {
	var output strings.Builder
	output.WriteString("Security Audit - Kernel Hardening (sysctl):\n\n")

	passed, warned := 0, 0
	for _, c := range sysctlChecks {
		raw, err := os.ReadFile("/proc/sys/" + strings.ReplaceAll(c.name, ".", "/")) //nolint:gosec // kernel sysfs paths
		if err != nil {
			output.WriteString(fmt.Sprintf("  SKIP %s (not readable on this system)\n", c.name))
			continue
		}
		value := strings.TrimSpace(string(raw))
		n, _ := strconv.ParseInt(value, 10, 64)

		pass := c.check(n)
		sysctlCompliance.WithLabelValues(c.name).Set(boolValue(pass))
		switch {
		case pass:
			passed++
			output.WriteString(fmt.Sprintf("  PASS %s = %s (%s)\n", c.name, value, c.okDesc))
		case c.infoOnly:
			output.WriteString(fmt.Sprintf("  INFO %s = %s (%s)\n", c.name, value, c.warnDesc))
		default:
			warned++
			output.WriteString(fmt.Sprintf("  WARN %s = %s (%s)\n", c.name, value, c.warnDesc))
		}
	}

	output.WriteString(fmt.Sprintf("\nSummary: %d passed, %d warnings.\n", passed, warned))
	return output.String()
}

// boolValue maps a boolean to the Prometheus gauge convention 1/0.
func boolValue(b bool) float64 {
	if b {
		return 1
	}
	return 0
}

// checkMACFramework reports the status of SELinux/AppArmor (mandatory access
// control) frameworks.
func checkMACFramework() string {
	var output strings.Builder
	output.WriteString("Security Audit - Mandatory Access Control:\n\n")
	output.WriteString("SELinux:\n  " + selinuxStatus() + "\n")
	output.WriteString("AppArmor:\n  " + apparmorStatus() + "\n")
	return output.String()
}

// selinuxStatus returns a one-line SELinux status description.
func selinuxStatus() string {
	if out, err := exec.Command("getenforce").CombinedOutput(); err == nil {
		return "Status: " + strings.TrimSpace(string(out))
	}
	if readTrimmed("/sys/fs/selinux/enforce") == "1" {
		return "Status: Enforcing"
	}
	if _, err := os.Stat("/sys/fs/selinux"); err == nil {
		return "Status: present but not enforcing (enforce file unreadable; try with sudo)"
	}
	return "Status: Not installed"
}

// apparmorStatus returns a one-line AppArmor status description.
func apparmorStatus() string {
	enabled := readTrimmed("/sys/module/apparmor/parameters/enabled")
	switch enabled {
	case "N":
		return "Status: Disabled"
	case "Y":
		if out, err := exec.Command("aa-status").CombinedOutput(); err == nil {
			return "Status: Enabled (" + apparmorLoadedProfiles(string(out)) + " profiles loaded)"
		}
		return "Status: Enabled (profile count requires root: run with sudo)"
	default:
		return "Status: Not installed"
	}
}

// apparmorLoadedProfiles extracts the number of loaded profiles from
// `aa-status` output ("N profiles are loaded.").
func apparmorLoadedProfiles(output string) string {
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "profiles are loaded") {
			if n, err := strconv.Atoi(strings.Fields(line)[0]); err == nil {
				return strconv.Itoa(n)
			}
		}
	}
	return "unknown"
}
