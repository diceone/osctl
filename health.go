package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
)

type HealthStatus string

const (
	StatusHealthy   HealthStatus = "healthy"
	StatusDegraded  HealthStatus = "degraded"
	StatusUnhealthy HealthStatus = "unhealthy"
)

type HealthCheck struct {
	Status  HealthStatus `json:"status"`
	Message string       `json:"message,omitempty"`
	Value   string       `json:"value,omitempty"`
}

type HealthResponse struct {
	Status    HealthStatus           `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Checks    map[string]HealthCheck `json:"checks"`
	Uptime    string                 `json:"uptime"`
}

func getHealthCheck() string {
	checks := make(map[string]HealthCheck)
	overallStatus := StatusHealthy

	// Check Memory
	v, err := mem.VirtualMemory()
	if err != nil {
		checks["memory"] = HealthCheck{
			Status:  StatusUnhealthy,
			Message: fmt.Sprintf("Failed to get memory info: %v", err),
		}
		overallStatus = StatusUnhealthy
	} else {
		memStatus := StatusHealthy
		if v.UsedPercent > 90 {
			memStatus = StatusUnhealthy
			overallStatus = StatusDegraded
		} else if v.UsedPercent > 80 {
			memStatus = StatusDegraded
			if overallStatus == StatusHealthy {
				overallStatus = StatusDegraded
			}
		}
		checks["memory"] = HealthCheck{
			Status:  memStatus,
			Value:   fmt.Sprintf("%.2f%%", v.UsedPercent),
			Message: fmt.Sprintf("Used: %d MB / Total: %d MB", v.Used/1024/1024, v.Total/1024/1024),
		}
	}

	// Check Disk Space
	d, err := disk.Usage("/")
	if err != nil {
		checks["disk"] = HealthCheck{
			Status:  StatusUnhealthy,
			Message: fmt.Sprintf("Failed to get disk info: %v", err),
		}
		overallStatus = StatusUnhealthy
	} else {
		diskStatus := StatusHealthy
		if d.UsedPercent > 95 {
			diskStatus = StatusUnhealthy
			overallStatus = StatusDegraded
		} else if d.UsedPercent > 85 {
			diskStatus = StatusDegraded
			if overallStatus == StatusHealthy {
				overallStatus = StatusDegraded
			}
		}
		checks["disk"] = HealthCheck{
			Status:  diskStatus,
			Value:   fmt.Sprintf("%.2f%%", d.UsedPercent),
			Message: fmt.Sprintf("Used: %d GB / Total: %d GB", d.Used/1024/1024/1024, d.Total/1024/1024/1024),
		}
	}

	// Check CPU
	cpuPercent, err := cpu.Percent(time.Second, false)
	if err != nil {
		checks["cpu"] = HealthCheck{
			Status:  StatusUnhealthy,
			Message: fmt.Sprintf("Failed to get CPU info: %v", err),
		}
		overallStatus = StatusUnhealthy
	} else {
		cpuStatus := StatusHealthy
		if cpuPercent[0] > 95 {
			cpuStatus = StatusDegraded
			if overallStatus == StatusHealthy {
				overallStatus = StatusDegraded
			}
		}
		checks["cpu"] = HealthCheck{
			Status:  cpuStatus,
			Value:   fmt.Sprintf("%.2f%%", cpuPercent[0]),
			Message: "CPU usage",
		}
	}

	// Get uptime
	uptimeStr := getUptime()

	response := HealthResponse{
		Status:    overallStatus,
		Timestamp: time.Now(),
		Checks:    checks,
		Uptime:    uptimeStr,
	}

	jsonData, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return fmt.Sprintf("Error encoding health check: %v", err)
	}

	return string(jsonData)
}
