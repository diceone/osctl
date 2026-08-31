package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// getDoctorReport is the one-shot diagnostic: health, failed services, disk
// pressure and the newest journal errors, prefixed with a summary verdict.
func getDoctorReport() string {
	var output strings.Builder
	output.WriteString("=== OSPCTL DOCTOR REPORT ===\n\n")

	// 1. Health check
	health := getHealthCheck()
	var hr HealthResponse
	if err := json.Unmarshal([]byte(health), &hr); err == nil {
		output.WriteString(fmt.Sprintf("Overall status: %s\n", hr.Status))
		for _, name := range []string{"memory", "disk", "cpu"} {
			if c, ok := hr.Checks[name]; ok {
				output.WriteString(fmt.Sprintf("  %-7s %s (%s)\n", name+":", c.Status, c.Value))
			}
		}
		output.WriteString(fmt.Sprintf("  uptime  %s\n", hr.Uptime))
	} else {
		output.WriteString("Overall status: unknown (health check failed)\n")
		output.WriteString(health)
	}

	// 2. Failed units
	if failedOut, err := exec.Command("systemctl", "list-units", "--failed", "--plain", "--no-legend").Output(); err == nil {
		failed := strings.TrimSpace(string(failedOut))
		if failed == "" {
			output.WriteString("\nFailed services: none\n")
		} else {
			output.WriteString("\nFailed services:\n" + failed + "\n")
		}
	}

	// 3. Disk pressure on critical mounts
	if out, err := exec.Command("df", "-h", "/").Output(); err == nil {
		output.WriteString("\nDisk usage (root):\n" + string(out))
	}

	// 4. Newest errors
	if errOut, err := exec.Command("journalctl", "-p", "err", "-n", "5", "--no-pager", "-q").Output(); err == nil {
		errText := strings.TrimSpace(string(errOut))
		if errText == "" {
			errText = "(no recent errors)"
		}
		output.WriteString("\nLast 5 journal errors:\n" + errText + "\n")
	}

	// 5. Pending updates
	if _, err := os.Stat("/etc/debian_version"); err == nil {
		if out, err := exec.Command("apt", "list", "--upgradable").Output(); err == nil {
			output.WriteString(fmt.Sprintf("\nPending updates: %d\n", strings.Count(string(out), "[upgradable")))
		}
	} else if _, err := os.Stat("/etc/redhat-release"); err == nil {
		if out, err := exec.Command("yum", "check-update", "--quiet").Output(); err == nil {
			output.WriteString(fmt.Sprintf("\nPending updates: ~%d\n", len(strings.Split(strings.TrimSpace(string(out)), "\n"))))
		}
	}

	overall := StatusHealthy
	var hr2 HealthResponse
	if err := json.Unmarshal([]byte(health), &hr2); err == nil {
		overall = hr2.Status
	}
	if overall == StatusHealthy {
		output.WriteString("\nVerdict: system looks healthy.\n")
	} else {
		output.WriteString(fmt.Sprintf("\nVerdict: attention needed (%s).\n", overall))
	}

	return output.String()
}
