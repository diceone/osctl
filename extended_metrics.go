package main

import (
	"fmt"
	"strings"

	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/net"
	"github.com/shirou/gopsutil/process"
)

// getNetworkIO returns network I/O statistics with rates
func getNetworkIO() string {
	stats, err := net.IOCounters(true)
	if err != nil {
		return fmt.Sprintf("Error getting network I/O: %v", err)
	}

	var output strings.Builder
	output.WriteString("Network I/O Statistics:\n\n")

	for _, stat := range stats {
		// Update Prometheus metrics
		networkIOBytes.WithLabelValues(stat.Name, "sent").Set(float64(stat.BytesSent))
		networkIOBytes.WithLabelValues(stat.Name, "recv").Set(float64(stat.BytesRecv))

		output.WriteString(fmt.Sprintf("Interface: %s\n", stat.Name))
		output.WriteString(fmt.Sprintf("  Bytes Sent: %s\n", formatBytes(stat.BytesSent)))
		output.WriteString(fmt.Sprintf("  Bytes Received: %s\n", formatBytes(stat.BytesRecv)))
		output.WriteString(fmt.Sprintf("  Packets Sent: %d\n", stat.PacketsSent))
		output.WriteString(fmt.Sprintf("  Packets Received: %d\n", stat.PacketsRecv))
		output.WriteString(fmt.Sprintf("  Errors In: %d, Out: %d\n", stat.Errin, stat.Errout))
		output.WriteString(fmt.Sprintf("  Drops In: %d, Out: %d\n\n", stat.Dropin, stat.Dropout))
	}

	return output.String()
}

// getDiskIO returns disk I/O statistics
func getDiskIO() string {
	ioCounters, err := disk.IOCounters()
	if err != nil {
		return fmt.Sprintf("Error getting disk I/O: %v", err)
	}

	var output strings.Builder
	output.WriteString("Disk I/O Statistics:\n\n")

	for device, stat := range ioCounters {
		// Update Prometheus metrics
		diskIOBytes.WithLabelValues(device, "read").Set(float64(stat.ReadBytes))
		diskIOBytes.WithLabelValues(device, "write").Set(float64(stat.WriteBytes))

		output.WriteString(fmt.Sprintf("Device: %s\n", device))
		output.WriteString(fmt.Sprintf("  Read: %s (%d ops)\n", formatBytes(stat.ReadBytes), stat.ReadCount))
		output.WriteString(fmt.Sprintf("  Write: %s (%d ops)\n", formatBytes(stat.WriteBytes), stat.WriteCount))
		output.WriteString(fmt.Sprintf("  Read Time: %d ms\n", stat.ReadTime))
		output.WriteString(fmt.Sprintf("  Write Time: %d ms\n", stat.WriteTime))
		output.WriteString(fmt.Sprintf("  IO Time: %d ms\n\n", stat.IoTime))
	}

	return output.String()
}

// getProcessCountByState returns count of processes by state
func getProcessCountByState() string {
	procs, err := process.Processes()
	if err != nil {
		return fmt.Sprintf("Error getting processes: %v", err)
	}

	stateCounts := make(map[string]int)

	for _, p := range procs {
		status, err := p.Status()
		if err != nil {
			continue
		}
		if len(status) > 0 {
			// status is a string, use first character as state
			stateCounts[string(status[0])]++
		}
	}

	// Update Prometheus metrics
	for state, count := range stateCounts {
		processCount.WithLabelValues(state).Set(float64(count))
	}

	var output strings.Builder
	output.WriteString("Process Count by State:\n\n")
	output.WriteString(fmt.Sprintf("Total Processes: %d\n\n", len(procs)))

	for state, count := range stateCounts {
		stateDesc := getProcessStateDescription(state)
		output.WriteString(fmt.Sprintf("%s (%s): %d\n", stateDesc, state, count))
	}

	return output.String()
}

// getProcessStateDescription returns human-readable process state
func getProcessStateDescription(state string) string {
	descriptions := map[string]string{
		"R": "Running",
		"S": "Sleeping",
		"D": "Disk Sleep",
		"Z": "Zombie",
		"T": "Stopped",
		"t": "Tracing Stop",
		"W": "Paging",
		"X": "Dead",
		"x": "Dead",
		"K": "Wakekill",
		"P": "Parked",
		"I": "Idle",
	}

	if desc, ok := descriptions[state]; ok {
		return desc
	}
	return "Unknown"
}

// formatBytes formats bytes into human-readable format
func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %ciB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
