package main

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// getDockerStats reports per-container CPU and memory usage from
// `docker stats --no-stream` and updates Prometheus gauges.
func getDockerStats() string {
	out, err := exec.Command("docker", "stats", "--no-stream", "--format",
		"{{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.MemPerc}}").CombinedOutput()
	if err != nil {
		msg := fmt.Sprintf("Failed to get Docker stats. Error: %v\n%s", err, string(out))
		if hint := permissionHint(err, string(out)); hint != "" {
			msg += "\n" + hint
		}
		return msg
	}

	raw := strings.TrimSpace(string(out))
	if raw == "" {
		return "No running containers."
	}

	var output strings.Builder
	output.WriteString("Docker Container Stats:\n\n")

	for _, line := range strings.Split(raw, "\n") {
		fields := strings.Split(line, "\t")
		if len(fields) != 4 {
			continue
		}
		name, cpuStr, memStr, memPctStr := fields[0], fields[1], fields[2], fields[3]

		cpuPct := parsePercentValue(cpuStr)
		used, limit := parseDockerMemUsage(memStr)
		memPct := parsePercentValue(memPctStr)

		dockerCPUPercent.WithLabelValues(name).Set(cpuPct)
		dockerMemBytes.WithLabelValues(name, "used").Set(used)
		dockerMemBytes.WithLabelValues(name, "limit").Set(limit)

		output.WriteString(fmt.Sprintf("Container: %s\n", name))
		output.WriteString(fmt.Sprintf("  CPU: %.2f%%\n", cpuPct))
		output.WriteString(fmt.Sprintf("  Memory: %s / %s (%.2f%%)\n\n",
			formatBytes(uint64(used)), formatBytes(uint64(limit)), memPct))
	}

	return output.String()
}

// parsePercentValue converts "23.45%" into 23.45 (0 for unparsable input).
func parsePercentValue(s string) float64 {
	v, _ := strconv.ParseFloat(strings.TrimSuffix(strings.TrimSpace(s), "%"), 64)
	return v
}

// parseDockerMemUsage splits "12.5MiB / 1.944GiB" into used and limit bytes.
func parseDockerMemUsage(s string) (float64, float64) {
	parts := strings.SplitN(s, "/", 2)
	used := parseDockerSize(strings.TrimSpace(parts[0]))
	var limit float64
	if len(parts) == 2 {
		limit = parseDockerSize(strings.TrimSpace(parts[1]))
	}
	return used, limit
}

// parseDockerSize converts a Docker size string ("12.5MiB", "318.7MB", "0B")
// into bytes. Both decimal (kB) and binary (KiB) unit families are supported.
func parseDockerSize(s string) float64 {
	if s == "" {
		return 0
	}
	unitIndex := strings.IndexFunc(s, func(r rune) bool {
		return (r < '0' || r > '9') && r != '.' && r != '-'
	})
	value := strings.TrimSpace(s)
	unit := ""
	if unitIndex >= 0 {
		value = strings.TrimSpace(s[:unitIndex])
		unit = strings.TrimSpace(s[unitIndex:])
	}
	num, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0
	}
	multiplier, ok := sizeUnitMultiplier(unit)
	if !ok {
		return 0
	}
	return num * multiplier
}

// sizeUnitMultiplier maps Docker size units to byte multipliers. Decimal
// units (kB, MB, ...) use powers of 1000, binary units (KiB, MiB, ...) 1024.
func sizeUnitMultiplier(unit string) (float64, bool) {
	switch unit {
	case "B":
		return 1, true
	case "kB", "KB":
		return 1e3, true
	case "MB":
		return 1e6, true
	case "GB":
		return 1e9, true
	case "TB":
		return 1e12, true
	case "PB":
		return 1e15, true
	case "KiB":
		return 1 << 10, true
	case "MiB":
		return 1 << 20, true
	case "GiB":
		return 1 << 30, true
	case "TiB":
		return 1 << 40, true
	case "PiB":
		return 1 << 50, true
	default:
		return 0, false
	}
}
