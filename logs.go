package main

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// getServiceLogs shows recent journal entries for a systemd unit.
func getServiceLogs(unit, linesArg string) string {
	unit = strings.TrimSpace(unit)
	if unit == "" {
		return "Usage: osctl logs <unit> [lines]"
	}
	if strings.ContainsAny(unit, ";|&$`\n\r\t /") || len(unit) > 128 {
		return "Invalid unit name: contains forbidden characters"
	}

	lines := 50
	if linesArg != "" {
		n, err := strconv.Atoi(linesArg)
		if err != nil || n < 1 || n > 10000 {
			return "Invalid line count: expected an integer between 1 and 10000"
		}
		lines = n
	}

	out, err := exec.Command("journalctl", "-u", unit, "-n", strconv.Itoa(lines), "--no-pager", "-q").CombinedOutput()
	if err != nil {
		msg := fmt.Sprintf("Failed to get logs for unit %s. Error: %v\n%s", unit, err, string(out))
		if hint := permissionHint(err, string(out)); hint != "" {
			msg += "\n" + hint
		}
		return msg
	}
	if strings.TrimSpace(string(out)) == "" {
		return fmt.Sprintf("No journal entries found for unit %s.", unit)
	}
	return string(out)
}
