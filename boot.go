package main

import (
	"fmt"
	"os/exec"
	"strings"
)

// getBootAnalysis reports boot time plus the slowest systemd units.
func getBootAnalysis() string {
	summary, err := exec.Command("systemd-analyze").CombinedOutput()
	if err != nil {
		msg := fmt.Sprintf("Failed to analyze boot. Error: %v\n%s", err, strings.TrimSpace(string(summary)))
		if hint := permissionHint(err, string(summary)); hint != "" {
			msg += "\n" + hint
		}
		return msg
	}

	var output strings.Builder
	output.WriteString("Boot Analysis:\n\n")
	output.WriteString("Summary: " + strings.TrimSpace(string(summary)) + "\n")

	if blame, err := exec.Command("systemd-analyze", "blame").Output(); err == nil {
		const maxUnits = 15
		lines := strings.Split(strings.TrimSpace(string(blame)), "\n")
		if len(lines) > 0 && lines[0] != "" {
			if len(lines) > maxUnits {
				lines = append(lines[:maxUnits], fmt.Sprintf("... (%d more units)", len(lines)-maxUnits))
			}
			output.WriteString("\nSlowest units to start:\n")
			for _, l := range lines {
				output.WriteString("  " + l + "\n")
			}
		}
	}

	if chain, err := exec.Command("systemd-analyze", "critical-chain").CombinedOutput(); err == nil && strings.TrimSpace(string(chain)) != "" {
		output.WriteString("\nCritical chain:\n" + strings.TrimSpace(string(chain)) + "\n")
	}

	return output.String()
}
