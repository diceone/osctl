package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

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
	// Extended metrics
	networkIOBytes = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "osctl_network_io_bytes",
			Help: "Network I/O in bytes",
		},
		[]string{"interface", "direction"},
	)
	diskIOBytes = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "osctl_disk_io_bytes",
			Help: "Disk I/O in bytes",
		},
		[]string{"device", "direction"},
	)
	processCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "osctl_process_count",
			Help: "Number of processes by state",
		},
		[]string{"state"},
	)
)

func init() {
	prometheus.MustRegister(ramUsage)
	prometheus.MustRegister(diskUsage)
	prometheus.MustRegister(cpuUsage)
	prometheus.MustRegister(networkIOBytes)
	prometheus.MustRegister(diskIOBytes)
	prometheus.MustRegister(processCount)
}

func runAPI() {
	port := os.Getenv("OSCTL_PORT")
	if port == "" {
		port = "12000"
	}

	// Protected endpoints with basic auth
	http.Handle("/", basicAuth(http.HandlerFunc(handleRequest)))

	// Public metrics endpoint
	http.Handle("/metrics", promhttp.Handler())

	addr := fmt.Sprintf(":%s", port)
	log.Printf("Server is listening on port %s...", port)
	log.Printf("Metrics endpoint available at http://localhost:%s/metrics", port)
	log.Fatal(http.ListenAndServe(addr, nil))
}
