package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// configKeys are the settings osctl understands in a config file. Anything
// else in the file is ignored, so unknown keys cannot smuggle in env vars.
var configKeys = map[string]bool{
	"OSCTL_PORT":            true,
	"OSCTL_USERNAME":        true,
	"OSCTL_PASSWORD":        true,
	"OSCTL_API_TOKEN":       true,
	"OSCTL_TLS_CERT":        true,
	"OSCTL_TLS_KEY":         true,
	"OSCTL_AUDIT_LOG":       true,
	"OSCTL_STATE_DIR":       true,
	"OSCTL_WEBHOOK_URL":     true,
	"OSCTL_HEALTH_INTERVAL": true,
}

// loadConfigFile reads simple KEY=VALUE pairs from path and sets them in the
// environment, WITHOUT overriding variables that are already set (environment
// wins over file). Missing or unreadable file is not an error.
func loadConfigFile(path string) error {
	if path == "" {
		return nil
	}

	f, err := os.Open(path) //nolint:gosec // path comes from trusted env/flag
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if !configKeys[key] {
			continue
		}
		if _, exists := os.LookupEnv(key); !exists {
			os.Setenv(key, value)
		}
	}
	return scanner.Err()
}

// applyConfig loads the file referenced by OSCTL_CONFIG (if any).
func applyConfigFile() {
	if path := os.Getenv("OSCTL_CONFIG"); path != "" {
		if err := loadConfigFile(path); err != nil {
			// Not fatal: environment variables still work; surface the reason.
			fmt.Fprintf(os.Stderr, "osctl: could not read config file %s: %v\n", path, err)
		}
	}
}
