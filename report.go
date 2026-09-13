package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type reportSection struct {
	name string
	fn   func() string
}

// reportSections are the read-only checks included in the system report.
var reportSections = []reportSection{
	{"osinfo", getOSInfo},
	{"uptime", getUptime},
	{"load", getLoadAverage},
	{"cpu", getCpuUsage},
	{"ram", getRamUsage},
	{"disk", getDiskUsage},
	{"health", getHealthCheck},
	{"services", getServiceStatuses},
	{"failed_units", getFailedUnits},
	{"listening_ports", getListeningPorts},
	{"sensors", getSensors},
	{"timesync", getTimeSyncStatus},
	{"updates", listPackageUpdates},
	{"boot", getBootAnalysis},
	{"containers", listDockerContainers},
	{"dockerstats", getDockerStats},
	{"who", getLoggedinUsers},
}

// generateReport builds a one-shot JSON snapshot of all read-only checks
// for tickets and triage.
func generateReport() string {
	sections := make(map[string]string, len(reportSections))
	for _, s := range reportSections {
		sections[s.name] = strings.TrimSpace(s.fn())
	}

	payload := map[string]any{
		"generated": time.Now().UTC().Format(time.RFC3339),
		"host":      hostname(),
		"osctl":     buildVersion,
		"sections":  sections,
	}

	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Sprintf("Failed to generate report. Error: %v", err)
	}
	return string(data)
}
