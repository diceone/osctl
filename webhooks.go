package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

// webhookMinInterval keeps misconfiguration from hammering the host or the
// webhook endpoint.
const minHealthInterval = 30 // seconds

// startHealthMonitor compares periodic health checks and POSTs a JSON payload
// to webhookURL whenever the overall status changes. Disabled without URL.
func startHealthMonitor(webhookURL string, intervalSeconds int) {
	if webhookURL == "" {
		return
	}
	if intervalSeconds < minHealthInterval {
		intervalSeconds = minHealthInterval
	}

	lastStatus := HealthStatus("")

	go func() {
		ticker := time.NewTicker(time.Duration(intervalSeconds) * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			checks := map[string]HealthCheck{}
			var hr HealthResponse
			if err := json.Unmarshal([]byte(getHealthCheck()), &hr); err != nil {
				log.Printf("Health monitor: cannot parse health check: %v", err)
				continue
			}
			checks = hr.Checks

			if lastStatus == "" {
				lastStatus = hr.Status // first iteration: baseline, no notify
				continue
			}
			if hr.Status == lastStatus {
				continue
			}

			previous := lastStatus
			lastStatus = hr.Status
			log.Printf("Health monitor: status changed %s -> %s", previous, hr.Status)
			notifyWebhook(webhookURL, previous, hr.Status, checks)
		}
	}()

	log.Printf("Health monitor active (every %ds) with webhook notifications enabled", intervalSeconds)
}

// notifyWebhook delivers the status-change payload; failures are logged only.
func notifyWebhook(webhookURL string, previous, current HealthStatus, checks map[string]HealthCheck) {
	payload, err := json.Marshal(map[string]interface{}{
		"previous": previous,
		"current":  current,
		"checks":   checks,
		"time":     time.Now().UTC().Format(time.RFC3339),
		"host":     hostname(),
	})
	if err != nil {
		return
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(webhookURL, "application/json", bytes.NewReader(payload)) //nolint:noctx // short-lived internal call
	if err != nil {
		log.Printf("Health monitor: webhook delivery failed: %v", err)
		return
	}
	resp.Body.Close()
	if resp.StatusCode >= 300 {
		log.Printf("Health monitor: webhook returned HTTP %d", resp.StatusCode)
	}
}

func hostname() string {
	h, err := os.Hostname()
	if err != nil {
		return fmt.Sprintf("unknown (%v)", err)
	}
	return h
}
