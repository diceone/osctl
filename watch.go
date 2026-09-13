package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// runWatch re-runs a command on an interval, printing each result with a
// timestamp header and notifying OSCTL_WEBHOOK_URL when the output changes.
// Usage: osctl watch [--interval SECONDS] <command> [args...]
func runWatch(args []string) string {
	interval := 60
	if len(args) >= 2 && (args[0] == "--interval" || args[0] == "-i") {
		n, err := strconv.Atoi(args[1])
		if err != nil || n < 1 || n > 86400 {
			return "Invalid watch interval: expected an integer between 1 and 86400 seconds"
		}
		interval = n
		args = args[2:]
	}
	if len(args) < 1 {
		return "Usage: osctl watch [--interval SECONDS] <command> [args...]"
	}

	webhookURL := os.Getenv("OSCTL_WEBHOOK_URL")
	var previous string
	first := true
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	for {
		result := executeCommand(args)
		fmt.Printf("\n=== %s ===\n%s\n", time.Now().Format("2006-01-02 15:04:05"), result)

		if !first && webhookURL != "" && strings.TrimSpace(result) != strings.TrimSpace(previous) {
			notifyWatchChange(webhookURL, args, previous, result)
		}
		previous = result
		first = false

		<-ticker.C
	}
}

// notifyWatchChange posts a change notification for watched command output.
func notifyWatchChange(webhookURL string, command []string, previous, current string) {
	payload := map[string]any{
		"event":    "watch_output_changed",
		"command":  strings.Join(command, " "),
		"host":     hostname(),
		"time":     time.Now().Format(time.RFC3339),
		"previous": strings.TrimSpace(previous),
		"current":  strings.TrimSpace(current),
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return
	}
	resp, err := http.Post(webhookURL, "application/json", bytes.NewReader(body)) //nolint:gosec // URL is operator-configured
	if err != nil {
		return
	}
	resp.Body.Close()
}
