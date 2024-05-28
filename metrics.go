package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
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
)

func init() {
	prometheus.MustRegister(ramUsage)
	prometheus.MustRegister(diskUsage)
	prometheus.MustRegister(cpuUsage)
}

func runAPI() {
	http.HandleFunc("/", handleRequest)
	http.Handle("/metrics", promhttp.Handler())
	log.Println("Server is listening on port 12000...")
	log.Fatal(http.ListenAndServe(":12000", nil))
}
