package main

import (
	"fmt"
	"os/exec"
	"strings"
)

// listCronJobs lists all cron jobs for all users
func listCronJobs() string {
	var output strings.Builder
	output.WriteString("Cron Jobs:\n\n")

	// List system-wide cron jobs
	output.WriteString("=== System Cron Jobs ===\n")

	// /etc/crontab
	cmd := exec.Command("cat", "/etc/crontab")
	out, err := cmd.CombinedOutput()
	if err == nil {
		output.WriteString("\n/etc/crontab:\n")
		output.WriteString(string(out))
		output.WriteString("\n")
	}

	// /etc/cron.d/
	cmd = exec.Command("sh", "-c", "ls -1 /etc/cron.d/ 2>/dev/null")
	out, err = cmd.CombinedOutput()
	if err == nil && len(out) > 0 {
		output.WriteString("\n/etc/cron.d/:\n")
		files := strings.Split(strings.TrimSpace(string(out)), "\n")
		for _, file := range files {
			if file == "" {
				continue
			}
			cmd = exec.Command("cat", "/etc/cron.d/"+file)
			fileOut, err := cmd.CombinedOutput()
			if err == nil {
				output.WriteString(fmt.Sprintf("\n  File: %s\n", file))
				output.WriteString(string(fileOut))
			}
		}
	}

	// User cron jobs
	output.WriteString("\n\n=== User Cron Jobs ===\n")
	cmd = exec.Command("sh", "-c", "for user in $(cut -f1 -d: /etc/passwd); do crontab -u $user -l 2>/dev/null && echo \"User: $user\"; done")
	out, err = cmd.CombinedOutput()
	if err == nil && len(out) > 0 {
		output.WriteString(string(out))
	} else {
		output.WriteString("No user cron jobs found or insufficient permissions\n")
	}

	return output.String()
}

// addCronJob adds a cron job for the current user
func addCronJob(schedule, command string) string {
	if schedule == "" || command == "" {
		return "Usage: osctl cron add \"schedule\" \"command\"\nExample: osctl cron add \"0 2 * * *\" \"/backup.sh\""
	}

	// Validate cron schedule format (basic validation)
	parts := strings.Fields(schedule)
	if len(parts) != 5 {
		return "Invalid cron schedule format. Expected 5 fields: minute hour day month weekday"
	}

	// Get current crontab
	cmd := exec.Command("crontab", "-l")
	currentCron, _ := cmd.CombinedOutput()

	// Append new job
	newCron := string(currentCron)
	if !strings.HasSuffix(newCron, "\n") && len(newCron) > 0 {
		newCron += "\n"
	}
	newCron += fmt.Sprintf("%s %s\n", schedule, command)

	// Write new crontab
	cmd = exec.Command("crontab", "-")
	cmd.Stdin = strings.NewReader(newCron)
	err := cmd.Run()
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
	currentCron, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Sprintf("Failed to get current crontab. Error: %v", err)
	}

	lines := strings.Split(string(currentCron), "\n")
	var lineNum int
	fmt.Sscanf(lineNumber, "%d", &lineNum)

	if lineNum < 1 || lineNum > len(lines) {
		return fmt.Sprintf("Invalid line number. Valid range: 1-%d", len(lines))
	}

	// Remove the line (convert to 0-based index)
	lines = append(lines[:lineNum-1], lines[lineNum:]...)
	newCron := strings.Join(lines, "\n")

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
	currentCron, err := cmd.CombinedOutput()
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
