package main

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// Limits to prevent oversize/injection payloads from reaching crontab.
const (
	maxCronScheduleLength = 100
	maxCronCommandLength  = 1024
)

// isCronFieldChar reports whether r is valid inside a single cron field
// (digits, step/range/list syntax, and month/weekday names).
func isCronFieldChar(r rune) bool {
	if (r >= '0' && r <= '9') || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
		return true
	}
	switch r {
	case '*', '/', '-', ',':
		return true
	}
	return false
}

// validateCronSchedule checks that the schedule has exactly 5 fields with only
// characters cron understands, rejecting anything that could smuggle in
// additional crontab lines.
func validateCronSchedule(schedule string) error {
	if len(schedule) > maxCronScheduleLength {
		return fmt.Errorf("schedule too long (max %d characters)", maxCronScheduleLength)
	}

	parts := strings.Fields(schedule)
	if len(parts) != 5 {
		return fmt.Errorf("invalid cron schedule format. Expected 5 fields: minute hour day month weekday")
	}

	for _, field := range parts {
		for _, r := range field {
			if !isCronFieldChar(r) {
				return fmt.Errorf("invalid character %q in schedule field %q (allowed: digits, *, /, -, ',' and names)", r, field)
			}
		}
	}
	return nil
}

// validateCronCommand rejects control characters (notably newlines) that would
// let a caller inject additional crontab entries, and empty commands.
func validateCronCommand(command string) error {
	if command == "" {
		return fmt.Errorf("command must not be empty")
	}
	if len(command) > maxCronCommandLength {
		return fmt.Errorf("too long (max %d characters)", maxCronCommandLength)
	}
	for _, r := range command {
		if r < 0x20 || r == 0x7f {
			return fmt.Errorf("contains control characters")
		}
	}
	return nil
}

// addCronJob adds a cron job for the current user
func addCronJob(schedule, command string) string {
	if schedule == "" || command == "" {
		return "Usage: osctl cron add \"schedule\" \"command\"\nExample: osctl cron add \"0 2 * * *\" \"/backup.sh\""
	}

	if err := validateCronSchedule(schedule); err != nil {
		return "Invalid cron schedule: " + err.Error()
	}
	if err := validateCronCommand(command); err != nil {
		return "Invalid cron command: " + err.Error()
	}

	// Get current crontab. Exit status 1 with no stdout means the user has
	// no crontab yet; only fail on a real error that produced output.
	cmd := exec.Command("crontab", "-l")
	currentCron, err := cmd.Output()
	if err != nil && len(currentCron) > 0 {
		return fmt.Sprintf("Failed to get current crontab. Error: %v", err)
	}

	// Append new job
	newCron := string(currentCron)
	if !strings.HasSuffix(newCron, "\n") && len(newCron) > 0 {
		newCron += "\n"
	}
	newCron += fmt.Sprintf("%s %s\n", schedule, command)

	// Write new crontab
	cmd = exec.Command("crontab", "-")
	cmd.Stdin = strings.NewReader(newCron)
	err = cmd.Run()
	if err != nil {
		return fmt.Sprintf("Failed to add cron job. Error: %v", err)
	}

	return fmt.Sprintf("Cron job added successfully:\n%s %s", schedule, command)
}

// removeCronJob removes a cron job by line number
func removeCronJob(lineNumber string) string {
	if lineNumber == "" {
		return "Usage: osctl cron remove <line_number>\nUse 'osctl cron list' to see line numbers"
	}

	// Get current crontab
	cmd := exec.Command("crontab", "-l")
	currentCron, err := cmd.Output()
	if err != nil && len(currentCron) == 0 {
		return fmt.Sprintf("Failed to get current crontab. Error: %v", err)
	}

	content := strings.TrimRight(string(currentCron), "\n")
	if content == "" {
		return "Crontab is empty; nothing to remove"
	}
	lines := strings.Split(content, "\n")

	lineNum, parseErr := strconv.Atoi(strings.TrimSpace(lineNumber))
	if parseErr != nil {
		return fmt.Sprintf("Invalid line number %q: must be an integer", lineNumber)
	}

	if lineNum < 1 || lineNum > len(lines) {
		return fmt.Sprintf("Invalid line number. Valid range: 1-%d", len(lines))
	}

	// Remove the line (convert to 0-based index)
	lines = append(lines[:lineNum-1], lines[lineNum:]...)
	newCron := strings.Join(lines, "\n") + "\n"

	// Write new crontab
	cmd = exec.Command("crontab", "-")
	cmd.Stdin = strings.NewReader(newCron)
	err = cmd.Run()
	if err != nil {
		return fmt.Sprintf("Failed to update crontab. Error: %v", err)
	}

	return fmt.Sprintf("Cron job at line %s removed successfully", lineNumber)
}

// listCronJobsFormatted lists cron jobs with line numbers
func listCronJobsFormatted() string {
	var output strings.Builder
	output.WriteString("Current User Cron Jobs:\n\n")

	cmd := exec.Command("crontab", "-l")
	currentCron, err := cmd.Output()
	if err != nil {
		return "No crontab for current user or insufficient permissions"
	}

	lines := strings.Split(string(currentCron), "\n")
	for i, line := range lines {
		if strings.TrimSpace(line) == "" || strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}
		output.WriteString(fmt.Sprintf("%d: %s\n", i+1, line))
	}

	if output.Len() == len("Current User Cron Jobs:\n\n") {
		return "No active cron jobs found"
	}

	return output.String()
}

// getCronNextRun shows when cron jobs will run next
func getCronNextRun() string {
	// This requires additional parsing of cron schedules
	// For simplicity, we'll show systemd timers which are easier to query
	var output strings.Builder
	output.WriteString("Systemd Timers (Next Scheduled Runs):\n\n")

	cmd := exec.Command("systemctl", "list-timers", "--all")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Sprintf("Failed to get timer information. Error: %v\nNote: Traditional cron doesn't provide next-run info easily.", err)
	}

	output.WriteString(string(out))
	return output.String()
}
