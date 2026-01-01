# osctl AI Coding Agent Instructions

## Project Overview
`osctl` is a dual-mode Linux system administration tool: a CLI for interactive use and an HTTP API server with Prometheus metrics. Built in Go, it wraps system commands and exposes system stats (RAM, disk, CPU, services, Docker, network).

## Architecture

### Code Organization (Flat Structure)
- **main.go**: CLI entry point with switch-case command router
- **handlers.go**: HTTP endpoint router mapping paths to system functions + `printHelp()`
- **services.go**: systemctl operations, shutdown/reboot, package updates (RHEL/Ubuntu/SUSE), Docker commands, IP addresses
- **system_info.go**: System stats via gopsutil library (RAM, disk, CPU, network, processes)
- **metrics.go**: Prometheus metrics setup (`ramUsage`, `diskUsage`, `cpuUsage`) + `runAPI()` server
- **auth.go**: Basic auth middleware (hardcoded credentials: admin/password)

### Dual Interface Pattern
Both CLI and API expose identical functionality:
- CLI: `osctl [command]` → calls functions directly → prints to stdout
- API: `GET /[command]` → `handleRequest()` routes to same functions → JSON response `{"result": "..."}`

### Building
**Critical**: Build command must include ALL Go files (no package structure):
```bash
go build -o osctl main.go auth.go metrics.go handlers.go system_info.go services.go
```

## Key Dependencies
- `github.com/prometheus/client_golang`: Metrics endpoint at `/metrics`
- `github.com/shirou/gopsutil`: Cross-platform system stats
- `github.com/vishvananda/netlink`: Linux network interface info
- External commands: `systemctl`, `journalctl`, `docker`, `ps`, `shutdown`, `apt-get`/`yum`/`zypper`

## Configuration & Environment Variables
- `OSCTL_PORT`: API server port (default: `12000`)
- `OSCTL_USERNAME`: Basic auth username (default: `admin`)
- `OSCTL_PASSWORD`: Basic auth password (default: `password`)

**Important**: Credentials are read from environment variables at runtime via `getAuthCredentials()` in auth.go.

## Development Patterns

### Adding New Commands
1. Add case to `main.go` switch statement
2. Add matching case to `handlers.go` switch statement
3. Implement function in appropriate file (services/system_info)
4. Update `printHelp()` in handlers.go
5. Document in README.md under Features + Commands sections

Example: For new command `netstat`, add to both routers calling `getNetstat()`.

### OS Detection Pattern
Uses modern `/etc/os-release` with fallback to legacy files:
```go
if data, err := os.ReadFile("/etc/os-release"); err == nil {
    osRelease := string(data)
    if strings.Contains(strings.ToLower(osRelease), "ubuntu") {
        // Ubuntu/Debian logic
    } else if strings.Contains(strings.ToLower(osRelease), "rhel") {
        // RHEL/CentOS/Fedora logic
    }
}
```
See `updatePackages()` in services.go for reference.

### Input Validation Pattern
Prevent command injection in service management:
```go
// Validate allowed actions
validActions := map[string]bool{"start": true, "stop": true, ...}

// Sanitize service names
if strings.ContainsAny(service, ";|&$`\n\r") {
    return "Invalid service name: contains forbidden characters"
}
```

### Error Handling Convention
**Critical**: Never use `log.Fatalf()` - it terminates the entire API server. Always return formatted error strings:
```go
return fmt.Sprintf("Failed to %s. Error: %v", operation, err)
```

### Metrics Integration
When adding commands that expose numeric stats, register Prometheus gauges:
1. Declare `prometheus.NewGaugeVec()` in metrics.go
2. Register in `init()` function
3. Call `.Set()` or `.WithLabelValues().Set()` in your function

## Security Patterns

### Authentication
- Basic Auth implemented via middleware in auth.go
- Credentials configurable via `OSCTL_USERNAME`/`OSCTL_PASSWORD` env vars
- WWW-Authenticate header sent on 401 responses
- `/metrics` endpoint is public (no auth) for Prometheus scraping

### Input Validation
Always validate external inputs:
- Service names checked for shell injection characters
- Action parameters validated against whitelist
- Length limits enforced on user inputs

## Testing & Running

### Local Development
```bash
# CLI mode
go run main.go auth.go metrics.go handlers.go system_info.go services.go ram

# API mode (requires root for most commands)
sudo go run main.go auth.go metrics.go handlers.go system_info.go services.go api
curl -u admin:password http://localhost:12000/ram
```

### API Authentication
Configurable via environment variables. API server uses Basic Auth middleware wrapping all endpoints except `/metrics`.

Example with custom config:
```bash
export OSCTL_PORT=8080
export OSCTL_USERNAME=myuser
export OSCTL_PASSWORD=securepass
sudo ./osctl api
```

### Deployment
- systemd service file: `systemd/osctl.service` (runs as root on port 12000)
- Dockerfile provided but incomplete (base config only)
- See systemd/README.md for service setup steps

## Platform Constraints
- **Linux-only**: Uses systemctl, journalctl, /proc filesystem
- **Requires root**: Most commands need elevated privileges (systemctl, shutdown, package updates)
- **No Windows/macOS support**: OS detection checks Linux-specific files
