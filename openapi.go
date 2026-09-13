package main

import (
	"encoding/json"
	"net/http"
)

// apiEndpoint describes one documented API route for /openapi.json.
type apiEndpoint struct {
	Path    string
	Method  string
	Summary string
	Params  []string // required query parameters
	Notes   string
}

// apiEndpoints is the source of truth for the generated OpenAPI document.
// CLI-only commands (watch, completion, api, help) have no HTTP route.
var apiEndpoints = []apiEndpoint{
	{"/ram", "GET", "Show RAM usage", nil, ""},
	{"/disk", "GET", "Show disk usage", nil, ""},
	{"/cpu", "GET", "Show CPU usage", nil, ""},
	{"/load", "GET", "Show system load averages", nil, ""},
	{"/uptime", "GET", "Show system uptime", nil, ""},
	{"/osinfo", "GET", "Show operating system name and kernel version", nil, ""},
	{"/top", "GET", "Show top processes by CPU usage", nil, ""},
	{"/procs", "GET", "Show process count by state", nil, ""},
	{"/process", "GET", "Process management", []string{"action"}, "action: kill, killforce, nice, info, tree; pid and priority for the mutating actions"},
	{"/network", "GET", "Show network statistics", nil, ""},
	{"/networkio", "GET", "Show network I/O statistics", nil, ""},
	{"/connections", "GET", "List all active network connections", nil, ""},
	{"/ip", "GET", "Show IP addresses of all interfaces", nil, ""},
	{"/filesystems", "GET", "List all mounted filesystems", nil, ""},
	{"/diskio", "GET", "Show disk I/O statistics", nil, ""},
	{"/services", "GET", "Show status of all running services", nil, ""},
	{"/service", "GET", "Manage a system service", []string{"action", "service"}, "action: start, stop, restart, status, enable, disable"},
	{"/logs", "GET", "Show recent journal entries for a systemd unit", []string{"unit"}, "lines: optional number of entries (1-10000, default 50)"},
	{"/errors", "GET", "Show last 10 errors from the journal", nil, ""},
	{"/dmesg", "GET", "Show kernel messages", nil, ""},
	{"/who", "GET", "List all currently logged in users", nil, ""},
	{"/users", "GET", "Show last 20 logged in users", nil, ""},
	{"/userinfo", "GET", "Show user identity and password aging", []string{"user"}, ""},
	{"/useradd", "POST", "Create a user with home directory", []string{"user"}, "POST only"},
	{"/userdel", "POST", "Delete a user (and home directory)", []string{"user"}, "POST only"},
	{"/firewall", "GET", "Show active firewall rules", nil, ""},
	{"/firewallallow", "POST", "Allow a port (ufw or firewalld)", []string{"port"}, "POST only"},
	{"/firewalldeny", "POST", "Deny/remove a port rule", []string{"port"}, "POST only"},
	{"/update", "GET", "Update OS packages", nil, ""},
	{"/updates", "GET", "List available package updates without installing", nil, ""},
	{"/containers", "GET", "List all Docker containers", nil, ""},
	{"/images", "GET", "List all Docker images", nil, ""},
	{"/dockerstats", "GET", "Show per-container CPU and memory stats", nil, ""},
	{"/dockerlogs", "GET", "Show last 50 log lines of a container", []string{"container"}, ""},
	{"/dockerrestart", "POST", "Restart a container", []string{"container"}, "POST only"},
	{"/audit", "GET", "Security audit", []string{"action"}, "action: ports, files, permissions, users, ssh, sysctl, mac, summary"},
	{"/cron", "GET", "Cron job management", []string{"action"}, "action: list, add, remove, next"},
	{"/maintenance", "GET", "Maintenance mode and system operations", []string{"action"}, "action: status, enable, disable, check-services, restart-failed, sync-time, clear-cache"},
	{"/health", "GET", "Show health check status", nil, ""},
	{"/doctor", "GET", "One-shot diagnostic: health, failed services, disk, errors, updates", nil, ""},
	{"/boot", "GET", "Show boot time and slowest units", nil, ""},
	{"/sensors", "GET", "Show temperatures and fan speeds", nil, ""},
	{"/certs", "GET", "Check TLS certificate expiry", nil, "target: optional certificate file path or host:port (repeatable)"},
	{"/version", "GET", "Show osctl version", nil, ""},
	{"/shutdown", "POST", "Shutdown the system", nil, "POST only"},
	{"/reboot", "POST", "Reboot the system", nil, "POST only"},
}

// buildOpenAPISpec assembles the OpenAPI 3.0.3 document as a generic map.
func buildOpenAPISpec() map[string]any {
	resultSchema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"result": map[string]string{"type": "string", "description": "Human-readable command output"},
		},
	}
	standardResponses := map[string]any{
		"200": map[string]any{
			"description": "Command executed; see the result field for the output",
			"content":     map[string]any{"application/json": map[string]any{"schema": resultSchema}},
		},
		"400": map[string]any{"description": "Invalid usage or missing parameters"},
		"401": map[string]any{"description": "Authentication required"},
		"405": map[string]any{"description": "Method not allowed (POST-only endpoint called with GET)"},
		"500": map[string]any{"description": "Command failed"},
	}

	paths := map[string]any{}
	for _, e := range apiEndpoints {
		item, ok := paths[e.Path].(map[string]any)
		if !ok {
			item = map[string]any{}
			paths[e.Path] = item
		}

		var params []map[string]any
		for _, name := range e.Params {
			params = append(params, map[string]any{
				"name":     name,
				"in":       "query",
				"required": true,
				"schema":   map[string]string{"type": "string"},
			})
		}

		op := map[string]any{
			"summary":   e.Summary,
			"responses": standardResponses,
		}
		if len(params) > 0 {
			op["parameters"] = params
		}
		if e.Notes != "" {
			op["description"] = e.Notes
		}
		item[mapLower(e.Method)] = op
	}

	return map[string]any{
		"openapi": "3.0.3",
		"info": map[string]any{
			"title":       "osctl API",
			"version":     buildVersion,
			"description": "HTTP API for Linux system administration with Prometheus metrics. All routes except /metrics and /openapi.json require Basic Auth or a Bearer token.",
		},
		"servers": []map[string]string{{"url": "http://localhost:12000"}},
		"components": map[string]any{
			"securitySchemes": map[string]any{
				"basicAuth":   map[string]string{"type": "http", "scheme": "basic"},
				"bearerToken": map[string]string{"type": "http", "scheme": "bearer"},
			},
		},
		"security": []map[string][]string{{"basicAuth": {}}, {"bearerToken": {}}},
		"paths":    paths,
	}
}

// openAPISpecJSON renders the OpenAPI document as pretty-printed JSON.
func openAPISpecJSON() []byte {
	spec := buildOpenAPISpec()
	data, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		// The spec is built from static data; failure is impossible in practice.
		return []byte(`{"error":"failed to render OpenAPI spec"}`)
	}
	return data
}

func serveOpenAPI(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(openAPISpecJSON())
}

func mapLower(method string) string {
	switch method {
	case "POST":
		return "post"
	default:
		return "get"
	}
}
