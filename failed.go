package main

import (
	"fmt"
	"os/exec"
	"strings"
)

// getFailedUnits lists failed systemd units via systemctl.
func getFailedUnits() string {
	out, err := exec.Command("systemctl", "list-units", "--state=failed", "--plain", "--no-legend", "--no-pager").CombinedOutput()
	if err != nil {
		msg := fmt.Sprintf("Failed to list failed units. Error: %v\n%s", err, strings.TrimSpace(string(out)))
		if hint := permissionHint(err, string(out)); hint != "" {
			msg += "\n" + hint
		}
		return msg
	}

	units := parseFailedUnits(string(out))
	failedUnits.Set(float64(len(units)))

	if len(units) == 0 {
		return "No failed systemd units."
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Failed systemd units (%d):\n\n", len(units))
	for _, u := range units {
		b.WriteString("  " + u + "\n")
	}
	return b.String()
}

// parseFailedUnits extracts unit names (first column) from
// systemctl list-units output.
func parseFailedUnits(raw string) []string {
	var units []string
	for _, line := range strings.Split(raw, "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 1 && fields[0] != "" {
			units = append(units, fields[0])
		}
	}
	return units
}
