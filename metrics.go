package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

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
	portNum, err := strconv.Atoi(port)
	if err != nil || portNum < 1 || portNum > 65535 {
		log.Fatalf("Invalid OSCTL_PORT %q: must be an integer between 1 and 65535", port)
	}

	if os.Getenv("OSCTL_PASSWORD") == "" {
		log.Printf("WARNING: OSCTL_PASSWORD is not set; the API will accept the default credentials (admin/password). Set OSCTL_USERNAME/OSCTL_PASSWORD to secure the API.")
	}

	// Protected endpoints with basic auth
	mux := http.NewServeMux()
	mux.Handle("/", basicAuth(http.HandlerFunc(handleRequest)))

	// Public metrics endpoint
	mux.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       60 * time.Second,
		// Long-running commands (e.g. package updates) return output only
		// when the command finishes, so allow generous write time.
		WriteTimeout:   300 * time.Second,
		IdleTimeout:    120 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Printf("Server is listening on port %s...", port)
	log.Printf("Metrics endpoint available at http://localhost:%s/metrics", port)
	log.Fatal(server.ListenAndServe())
}
