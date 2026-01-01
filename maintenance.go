package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

const maintenanceFlagFile = "/tmp/osctl_maintenance_mode"

// MaintenanceStatus represents the current maintenance mode state
type MaintenanceStatus struct {
	Enabled   bool      `json:"enabled"`
	Message   string    `json:"message,omitempty"`
	EnabledAt time.Time `json:"enabled_at,omitempty"`
	EnabledBy string    `json:"enabled_by,omitempty"`
}

// enableMaintenanceMode activates maintenance mode
func enableMaintenanceMode(message string) string {
	status := MaintenanceStatus{
		Enabled:   true,
		Message:   message,
		EnabledAt: time.Now(),
		EnabledBy: os.Getenv("USER"),
	}

	data, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		return fmt.Sprintf("Failed to create maintenance status: %v", err)
	}

	if err := os.WriteFile(maintenanceFlagFile, data, 0644); err != nil {
		return fmt.Sprintf("Failed to enable maintenance mode: %v", err)
	}

	// Broadcast message to all logged in users
	if message != "" {
		cmd := exec.Command("wall", message)
		cmd.Run() // Ignore errors, not critical
	}

	return "Maintenance mode enabled successfully"
}

// disableMaintenanceMode deactivates maintenance mode
func disableMaintenanceMode() string {
	if _, err := os.Stat(maintenanceFlagFile); os.IsNotExist(err) {
		return "Maintenance mode is not enabled"
	}

	if err := os.Remove(maintenanceFlagFile); err != nil {
		return fmt.Sprintf("Failed to disable maintenance mode: %v", err)
	}

	// Broadcast message to all logged in users
	cmd := exec.Command("wall", "Maintenance mode has been disabled. System is now operational.")
	cmd.Run() // Ignore errors, not critical

	return "Maintenance mode disabled successfully"
}

// getMaintenanceStatus returns the current maintenance mode status
func getMaintenanceStatus() string {
	status := MaintenanceStatus{
		Enabled: false,
	}

	data, err := os.ReadFile(maintenanceFlagFile)
	if err == nil {
		json.Unmarshal(data, &status)
	}

	output, _ := json.MarshalIndent(status, "", "  ")
	return string(output)
}

// getMaintenanceActions performs various maintenance-related actions
func getMaintenanceActions(action string) string {
	switch action {
	case "status":
		return getMaintenanceStatus()

	case "enable":
		return enableMaintenanceMode("System is entering maintenance mode. Services may be temporarily unavailable.")

	case "disable":
		return disableMaintenanceMode()

	case "check-services":
		// Check if critical services are running
		services := []string{"sshd", "systemd-journald", "systemd-logind"}
		var results []string
		results = append(results, "Critical Services Status:\n")

		for _, svc := range services {
			cmd := exec.Command("systemctl", "is-active", svc)
			output, _ := cmd.Output()
			state := strings.TrimSpace(string(output))
			results = append(results, fmt.Sprintf("  %s: %s", svc, state))
		}
		return strings.Join(results, "\n")

	case "restart-failed":
		// Restart all failed systemd services
		cmd := exec.Command("systemctl", "list-units", "--failed", "--plain", "--no-legend")
		output, err := cmd.Output()
		if err != nil {
			return fmt.Sprintf("Failed to list failed services: %v", err)
		}

		failed := strings.TrimSpace(string(output))
		if failed == "" {
			return "No failed services found"
		}

		lines := strings.Split(failed, "\n")
		var results []string
		results = append(results, fmt.Sprintf("Restarting %d failed services:\n", len(lines)))

		for _, line := range lines {
			fields := strings.Fields(line)
			if len(fields) > 0 {
				serviceName := fields[0]
				cmd := exec.Command("systemctl", "restart", serviceName)
				err := cmd.Run()
				if err != nil {
					results = append(results, fmt.Sprintf("  %s: FAILED to restart", serviceName))
				} else {
					results = append(results, fmt.Sprintf("  %s: restarted", serviceName))
				}
			}
		}
		return strings.Join(results, "\n")

	case "sync-time":
		// Sync system time
		cmd := exec.Command("timedatectl", "set-ntp", "true")
		if err := cmd.Run(); err != nil {
			return fmt.Sprintf("Failed to enable NTP: %v", err)
		}

		cmd = exec.Command("systemctl", "restart", "systemd-timesyncd")
		if err := cmd.Run(); err != nil {
			return fmt.Sprintf("Failed to restart time sync: %v", err)
		}

		return "Time synchronization restarted successfully"

	case "clear-cache":
		// Clear system caches
		var results []string

		// Drop caches (requires root)
		cmd := exec.Command("sh", "-c", "sync && echo 3 > /proc/sys/vm/drop_caches")
		if err := cmd.Run(); err != nil {
			results = append(results, fmt.Sprintf("Failed to drop caches: %v", err))
		} else {
			results = append(results, "System caches cleared")
		}

		// Clear systemd journal logs older than 7 days
		cmd = exec.Command("journalctl", "--vacuum-time=7d")
		output, err := cmd.Output()
		if err != nil {
			results = append(results, fmt.Sprintf("Failed to vacuum journal: %v", err))
		} else {
			results = append(results, fmt.Sprintf("Journal vacuumed: %s", strings.TrimSpace(string(output))))
		}

		return strings.Join(results, "\n")

	default:
		return fmt.Sprintf("Unknown maintenance action: %s\nValid actions: status, enable, disable, check-services, restart-failed, sync-time, clear-cache", action)
	}
}
