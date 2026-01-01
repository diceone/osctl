# Release Notes v0.0.8

## 🛠️ New Feature: Maintenance Mode

This release adds comprehensive maintenance mode functionality for system administrators.

### Maintenance Mode Operations

- **`maintenance status`**: View current maintenance mode state with JSON output
  - Shows enabled/disabled status
  - Displays activation timestamp
  - Shows user who enabled maintenance mode
  
- **`maintenance enable`**: Enter maintenance mode
  - Creates flag file tracking maintenance state
  - Broadcasts message to all logged-in users via `wall`
  - Records activation time and username

- **`maintenance disable`**: Exit maintenance mode
  - Removes maintenance flag
  - Notifies all users system is operational

- **`maintenance check-services`**: Verify critical services
  - Checks sshd, systemd-journald, systemd-logind
  - Reports status of each service

- **`maintenance restart-failed`**: Auto-restart failed services
  - Finds all failed systemd units
  - Attempts automatic restart
  - Provides detailed feedback

- **`maintenance sync-time`**: System time synchronization
  - Enables NTP via timedatectl
  - Restarts systemd-timesyncd

- **`maintenance clear-cache`**: Cache and log cleanup
  - Drops system caches (`/proc/sys/vm/drop_caches`)
  - Vacuums journal logs older than 7 days

### API Support

All maintenance operations available via HTTP API:
```bash
curl -u admin:password "http://localhost:12000/maintenance?action=status"
curl -u admin:password "http://localhost:12000/maintenance?action=enable"
curl -u admin:password "http://localhost:12000/maintenance?action=check-services"
```

## 📦 Installation

```bash
# Download binary for your platform
wget https://github.com/diceone/osctl/releases/download/v0.0.8/osctl-linux-amd64
chmod +x osctl-linux-amd64
sudo mv osctl-linux-amd64 /usr/local/bin/osctl

# Or build from source
git clone https://github.com/diceone/osctl.git
cd osctl
git checkout v0.0.8
go build -o osctl main.go auth.go metrics.go handlers.go system_info.go services.go health.go process.go extended_metrics.go security.go cron.go maintenance.go
```

## 🚀 What's Changed

- Added maintenance.go with system maintenance operations
- Updated CLI routing for maintenance commands
- Added API endpoint for maintenance operations
- Updated documentation with maintenance examples

Full Changelog: https://github.com/diceone/osctl/compare/v0.0.7...v0.0.8

---

# Release Notes v0.0.7

## 🎉 Major Feature Expansion

This release adds 5 major new feature categories, significantly expanding osctl's capabilities.

### 1. Health Check System

- **JSON-based health endpoint** for monitoring integration
- Configurable thresholds for memory, disk, and CPU
- Status levels: healthy, degraded, unhealthy
- Timestamp and uptime included in response

Example:
```bash
./osctl health
```

### 2. Process Management

- **`process kill <pid>`**: Send SIGTERM to process
- **`process killforce <pid>`**: Send SIGKILL for forced termination
- **`process nice <pid> <priority>`**: Adjust process priority (-20 to 19)
- **`process info <pid>`**: Display detailed process information
- **`process tree <pid>`**: Show process hierarchy

### 3. Extended Prometheus Metrics

New Prometheus gauges for enhanced monitoring:
- **Network I/O**: Bytes sent/received per interface
- **Disk I/O**: Read/write statistics per device  
- **Process counts**: Processes by state (running, sleeping, zombie, etc.)

Commands:
- `networkio`: Network I/O with human-readable formatting
- `diskio`: Disk I/O statistics with timing
- `procs`: Process count breakdown by state

### 4. Security Audit Tools

Comprehensive security scanning capabilities:
- **`audit ports`**: List all open listening ports (TCP/UDP)
- **`audit files`**: Detect suspicious file permissions (world-writable, SUID/SGID)
- **`audit permissions`**: Check critical system file permissions
- **`audit ssh`**: Audit SSH configuration for weak settings
- **`audit users`**: List user accounts with login shells
- **`audit summary`**: Complete security audit report

### 5. Cron Job Management

Full crontab management interface:
- **`cron list`**: Display all cron jobs with line numbers
- **`cron add "schedule" "command"`**: Add new cron job
- **`cron remove <line>`**: Remove job by line number
- **`cron next`**: Show next scheduled systemd timer runs

## ⚙️ Technical Updates

- Extended metrics.go with 3 new Prometheus gauge vectors
- Updated GitHub Actions workflow for all new files
- Comprehensive README updates with examples
- All features available via both CLI and HTTP API

## 📦 Installation

```bash
# Download binary
wget https://github.com/diceone/osctl/releases/download/v0.0.7/osctl-linux-amd64
chmod +x osctl-linux-amd64
sudo mv osctl-linux-amd64 /usr/local/bin/osctl

# Or build from source
git clone https://github.com/diceone/osctl.git
cd osctl
git checkout v0.0.7
go build -o osctl main.go auth.go metrics.go handlers.go system_info.go services.go health.go process.go extended_metrics.go security.go cron.go
```

## 🚀 What's Changed

New files added:
- health.go: Health monitoring system
- process.go: Process management operations
- extended_metrics.go: Enhanced Prometheus metrics
- security.go: Security audit utilities
- cron.go: Cron job management

Full Changelog: https://github.com/diceone/osctl/compare/v0.0.6...v0.0.7

---

# Release Notes v0.0.6

## 🔒 Security Enhancements

- **Environment Variable Configuration**: Configure credentials and port via `OSCTL_USERNAME`, `OSCTL_PASSWORD`, and `OSCTL_PORT`
- **Proper Basic Auth**: Added WWW-Authenticate headers for standard-compliant HTTP authentication
- **Input Validation**: Service names are validated to prevent command injection attacks
- **Parameter Limits**: Length limits enforced on API parameters

## 🛠️ Stability Improvements

- **Error Handling**: Replaced `log.Fatalf()` with error returns - API server no longer crashes on individual command failures
- **Graceful Degradation**: Server continues running even when individual system commands fail
- **Better Error Messages**: Improved error reporting for easier troubleshooting

## ⚙️ Configuration

All configuration now via environment variables:
- `OSCTL_PORT`: Server port (default: 12000)
- `OSCTL_USERNAME`: Basic auth username (default: admin)
- `OSCTL_PASSWORD`: Basic auth password (default: password)

Example:
```bash
export OSCTL_PORT=8080
export OSCTL_USERNAME=myuser
export OSCTL_PASSWORD=securepass
./osctl api
```

## 🐧 OS Compatibility

- **Modern OS Detection**: Uses `/etc/os-release` (modern standard) with fallback to legacy files
- **Better Distribution Support**: Improved detection for RHEL, CentOS, Fedora, Ubuntu, Debian, SUSE, openSUSE

## 📖 Documentation

- **AI Agent Instructions**: Added `.github/copilot-instructions.md` for AI coding assistants
- **Security Best Practices**: Comprehensive security section in README
- **Deployment Examples**: Updated examples with environment variable configuration

## 🔄 Breaking Changes

**None** - This release is fully backward compatible. Default values match previous behavior.

## 📦 Installation

```bash
# Download binary
wget https://github.com/diceone/osctl/releases/download/v0.0.6/osctl

# Or build from source
git clone https://github.com/diceone/osctl.git
cd osctl
git checkout v0.0.6
go build -o osctl main.go auth.go metrics.go handlers.go system_info.go services.go
```

## 🚀 What's Changed

Full Changelog: https://github.com/diceone/osctl/compare/v0.0.5...v0.0.6
