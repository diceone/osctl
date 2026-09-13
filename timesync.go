package main

import (
	"fmt"
	"os/exec"
	"strings"
)

// getTimeSyncStatus reports clock synchronization (timedatectl, with a
// chronyc fallback) and sets the osctl_time_synced gauge.
func getTimeSyncStatus() string {
	out, err := exec.Command("timedatectl").CombinedOutput()
	if err != nil {
		if chrono, cerr := exec.Command("chronyc", "tracking").CombinedOutput(); cerr == nil {
			return renderChronyStatus(string(chrono))
		}
		msg := fmt.Sprintf("Failed to get time synchronization status. Error: %v\n%s", err, strings.TrimSpace(string(out)))
		if hint := permissionHint(err, string(out)); hint != "" {
			msg += "\n" + hint
		}
		return msg
	}

	status := parseTimedatectl(string(out))
	if strings.EqualFold(status["synchronized"], "yes") {
		timeSynced.Set(1)
	} else {
		timeSynced.Set(0)
	}

	var b strings.Builder
	b.WriteString("Time Synchronization:\n\n")
	fmt.Fprintf(&b, "  Local time:              %s\n", status["local"])
	fmt.Fprintf(&b, "  Time zone:               %s\n", status["timezone"])
	fmt.Fprintf(&b, "  System clock synchronized: %s\n", status["synchronized"])
	fmt.Fprintf(&b, "  NTP service:             %s\n", status["ntp"])
	fmt.Fprintf(&b, "  RTC in local TZ:         %s\n", status["rtcLocal"])
	if status["synchronized"] == "yes" {
		b.WriteString("\nClock is synchronized.\n")
	} else {
		b.WriteString("\n⚠️ Clock is NOT synchronized.\n")
	}
	return b.String()
}

// parseTimedatectl extracts the relevant fields from `timedatectl` output.
func parseTimedatectl(raw string) map[string]string {
	status := map[string]string{}
	for _, line := range strings.Split(raw, "\n") {
		idx := strings.Index(line, ":")
		if idx < 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		value := strings.TrimSpace(line[idx+1:])
		switch {
		case strings.HasPrefix(key, "Local time"):
			status["local"] = value
		case strings.HasPrefix(key, "Time zone"):
			status["timezone"] = value
		case strings.HasPrefix(key, "System clock synchronized"):
			status["synchronized"] = value
		case strings.HasPrefix(key, "NTP service"):
			status["ntp"] = value
		case strings.HasPrefix(key, "RTC in local TZ"):
			status["rtcLocal"] = value
		}
	}
	return status
}

// renderChronyStatus falls back to `chronyc tracking` output when
// timedatectl is unavailable.
func renderChronyStatus(raw string) string {
	var b strings.Builder
	b.WriteString("Time Synchronization (chrony):\n\n")
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "Stratum"):
			b.WriteString("  " + line + "\n")
		case strings.HasPrefix(line, "System time"):
			// e.g. "System time     : 0.000000123 seconds slow of NTP time"
			b.WriteString("  " + line + "\n")
		case strings.HasPrefix(line, "Leap status"):
			b.WriteString("  " + line + "\n")
		}
	}
	b.WriteString("\n(timedatectl not available; showing chrony tracking)\n")
	return b.String()
}
