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
	dockerCPUPercent = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "osctl_docker_cpu_percent",
			Help: "CPU usage percent per Docker container",
		},
		[]string{"container"},
	)
	dockerMemBytes = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "osctl_docker_mem_bytes",
			Help: "Memory usage in bytes per Docker container",
		},
		[]string{"container", "type"},
	)
	sensorTempCelsius = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "osctl_sensor_temp_celsius",
			Help: "Temperature readings in Celsius from hwmon devices",
		},
		[]string{"device", "label"},
	)
	sensorFanRPM = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "osctl_sensor_fan_rpm",
			Help: "Fan speed readings in RPM from hwmon devices",
		},
		[]string{"device", "label"},
	)
	certExpirySeconds = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "osctl_cert_expiry_timestamp_seconds",
			Help: "Unix timestamp of TLS certificate expiry",
		},
		[]string{"source"},
	)
	sysctlCompliance = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "osctl_sysctl_compliance",
			Help: "Kernel hardening compliance per sysctl check (1 = pass, 0 = warning)",
		},
		[]string{"check"},
	)
)

func init() {
	prometheus.MustRegister(ramUsage)
	prometheus.MustRegister(diskUsage)
	prometheus.MustRegister(cpuUsage)
	prometheus.MustRegister(networkIOBytes)
	prometheus.MustRegister(diskIOBytes)
	prometheus.MustRegister(processCount)
	prometheus.MustRegister(dockerCPUPercent)
	prometheus.MustRegister(dockerMemBytes)
	prometheus.MustRegister(sensorTempCelsius)
	prometheus.MustRegister(sensorFanRPM)
	prometheus.MustRegister(certExpirySeconds)
	prometheus.MustRegister(sysctlCompliance)
}

func runAPI() {
	// Note: applyConfigFile() has already run in main(); do not load twice.
	port := os.Getenv("OSCTL_PORT")
	if port == "" {
		port = "12000"
	}
	portNum, err := strconv.Atoi(port)
	if err != nil || portNum < 1 || portNum > 65535 {
		log.Fatalf("Invalid OSCTL_PORT %q: must be an integer between 1 and 65535", port)
	}

	if os.Getenv("OSCTL_PASSWORD") == "" && os.Getenv("OSCTL_API_TOKEN") == "" {
		log.Printf("WARNING: neither OSCTL_PASSWORD nor OSCTL_API_TOKEN is set; the API will accept the default credentials (admin/password).")
	}

	// Restore brute-force protection state across restarts.
	loadAuthFailures(osctlStateDir())

	// Optional JSONL request audit log.
	auditLogPath := os.Getenv("OSCTL_AUDIT_LOG")
	if err := initAuditLog(auditLogPath); err != nil {
		log.Fatalf("Cannot open OSCTL_AUDIT_LOG %q: %v", auditLogPath, err)
	}
	if auditLogPath != "" {
		log.Printf("Request audit log: %s", auditLogPath)
	}

	// Protected endpoints with basic auth
	mux := http.NewServeMux()
	mux.Handle("/", basicAuth(auditMiddleware(http.HandlerFunc(handleRequest))))

	// Public OpenAPI document for the API routes.
	mux.HandleFunc("/openapi.json", serveOpenAPI)

	// Metrics endpoint; auth can be enforced with OSCTL_METRICS_AUTH=1.
	if envEnabled("OSCTL_METRICS_AUTH") {
		mux.Handle("/metrics", basicAuth(promhttp.Handler()))
	} else {
		mux.Handle("/metrics", promhttp.Handler())
	}

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

	scheme := "http"

	// Optional TLS. Both cert and key must be configured together.
	tlsCert := os.Getenv("OSCTL_TLS_CERT")
	tlsKey := os.Getenv("OSCTL_TLS_KEY")
	switch {
	case tlsCert != "" && tlsKey != "":
		scheme = "https"
	case tlsCert != "" || tlsKey != "":
		log.Fatalf("OSCTL_TLS_CERT and OSCTL_TLS_KEY must be set together")
	}

	// Optional health-change webhook notifications.
	if webhookURL := os.Getenv("OSCTL_WEBHOOK_URL"); webhookURL != "" {
		interval := 300
		if v, err := strconv.Atoi(os.Getenv("OSCTL_HEALTH_INTERVAL")); err == nil && v > 0 {
			interval = v
		}
		startHealthMonitor(webhookURL, interval)
	}

	log.Printf("Server is listening on port %s (%s)...", port, scheme)
	log.Printf("Metrics endpoint available at %s://localhost:%s/metrics", scheme, port)

	if scheme == "https" {
		log.Fatal(server.ListenAndServeTLS(tlsCert, tlsKey))
	}
	log.Fatal(server.ListenAndServe())
}
