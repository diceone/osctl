package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
	"github.com/shirou/gopsutil/process"
	"github.com/vishvananda/netlink"
)

const (
	username = "admin"
	password = "password"
)

// Prometheus metrics
var (
	ramUsage = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "osctl_ram_usage_bytes",
			Help: "RAM usage in bytes",
		},
		[]string{"type"},
	)
	diskUsage = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "osctl_disk_usage_bytes",
			Help: "Disk usage in bytes",
		},
		[]string{"type"},
	)
	cpuUsage = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "osctl_cpu_usage_percent",
			Help: "CPU usage in percent",
		},
	)
)

func init() {
	prometheus.MustRegister(ramUsage)
	prometheus.MustRegister(diskUsage)
	prometheus.MustRegister(cpuUsage)
}

func basicAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		payload, err := base64.StdEncoding.DecodeString(auth[len("Basic "):])
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		pair := strings.SplitN(string(payload), ":", 2)
		if len(pair) != 2 || pair[0] != username || pair[1] != password {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func getRamUsage() string {
	v, err := mem.VirtualMemory()
	if err != nil {
		log.Fatalf("Error getting RAM usage: %v", err)
	}

	ramUsage.WithLabelValues("total").Set(float64(v.Total))
	ramUsage.WithLabelValues("used").Set(float64(v.Used))
	ramUsage.WithLabelValues("free").Set(float64(v.Available))

	return fmt.Sprintf("Total: %v MB, Used: %v MB, Free: %v MB",
		v.Total/1024/1024, v.Used/1024/1024, v.Available/1024/1024)
}

func getDiskUsage() string {
	d, err := disk.Usage("/")
	if err != nil {
		log.Fatalf("Error getting disk usage: %v", err)
	}

	diskUsage.WithLabelValues("total").Set(float64(d.Total))
	diskUsage.WithLabelValues("used").Set(float64(d.Used))
	diskUsage.WithLabelValues("free").Set(float64(d.Free))

	return fmt.Sprintf("Total: %v GB, Used: %v GB, Free: %v GB",
		d.Total/1024/1024/1024, d.Used/1024/1024/1024, d.Free/1024/1024/1024)
}

func manageService(action, service string) string {
	cmd := exec.Command("systemctl", action, service)
	err := cmd.Run()
	if err != nil {
		return fmt.Sprintf("Failed to %s service %s. Error: %v", action, service, err)
	}
	return fmt.Sprintf("Service %s %sed successfully.", service, action)
}

func getTopProcesses() string {
	procs, err := process.Processes()
	if err
