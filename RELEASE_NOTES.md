# Release Notes v0.3.4

> Per-version notes below. The full release history is kept in [CHANGELOG.md](CHANGELOG.md).

## 🔍 Diagnostics

Five new read-only diagnostic commands, all available as CLI commands and
API routes (`/failed`, `/ports`, `/hwinfo`, `/timesync`, `/report`, also
under `/v1/`):

- **`failed`** — list failed systemd units, with the `osctl_failed_units`
  gauge for Prometheus alerting.
- **`ports`** — listening TCP ports with their owning processes. Uses
  `ss -tlnp` with a `netstat -tlnp` fallback. Exposes
  `osctl_listening_port{port,process}`.
- **`hwinfo`** — PCI devices (`lspci`), USB devices (`lsusb`) and loaded
  kernel modules (`lsmod`); missing tools are noted instead of failing.
- **`timesync`** — clock synchronization status via `timedatectl`, with a
  `chronyc tracking` fallback. Exposes the `osctl_time_synced` gauge.
- **`report`** — one-shot JSON snapshot of all 17 read-only diagnostics
  (host, osctl version, per-section outputs) — handy for cron jobs,
  monitoring pipelines and support tickets.

Unit tests cover all new parsers and report generation.

## 📦 Installation

```bash
# Debian/Ubuntu
wget https://github.com/diceone/osctl/releases/download/v0.3.4/osctl_0.3.4_amd64.deb
sudo apt install ./osctl_0.3.4_amd64.deb
sudo systemctl restart osctl

# RHEL/Fedora/SUSE
sudo rpm -Uvh https://github.com/diceone/osctl/releases/download/v0.3.4/osctl_0.3.4_amd64.rpm

# Or download the binary for your platform from the release page
```

---

# Release Notes v0.3.3

> Per-version notes below. The full release history is kept in [CHANGELOG.md](CHANGELOG.md).

## 🔧 Packaging Fix

- **systemd service file**: `ExecStart` now points to `/usr/bin/osctl`, the
  path used by the GoReleaser deb/rpm packages (v0.3.2 shipped the service
  file pointing at `/root/osctl`, which would fail to start after a package
  install). The obsolete `WorkingDirectory=/root/osctl` was removed.
- No code changes; the binaries are identical to v0.3.2 apart from the
  version string.

## 📦 Installation

```bash
# Debian/Ubuntu
wget https://github.com/diceone/osctl/releases/download/v0.3.3/osctl_0.3.3_amd64.deb
sudo apt install ./osctl_0.3.3_amd64.deb

# RHEL/CentOS/Fedora/SUSE
wget https://github.com/diceone/osctl/releases/download/v0.3.3/osctl_0.3.3_amd64.rpm
sudo rpm -i osctl_0.3.3_amd64.rpm

# Or download the binary
wget https://github.com/diceone/osctl/releases/download/v0.3.3/osctl_0.3.3_linux_amd64.tar.gz
tar xzf osctl_0.3.3_linux_amd64.tar.gz
chmod +x osctl
sudo mv osctl /usr/local/bin/osctl
```

Full Changelog: https://github.com/diceone/osctl/compare/v0.3.2...v0.3.3

---

# Release Notes v0.3.2

## 🚀 New Commands

All of these work in the CLI and through the API (also under the `/v1/` prefix):

- **`updates`**: List available package updates without installing them (apt/dnf/yum/zypper)
- **`logs <unit> [lines]`**: Show recent journal entries for a systemd unit (default 50, max 10000)
- **`dockerstats`**: Per-container CPU and memory usage with Prometheus gauges (`osctl_docker_cpu_percent`, `osctl_docker_mem_bytes`)
- **`sensors`**: Temperatures and fan speeds from `/sys/class/hwmon` with gauges (`osctl_sensor_temp_celsius`, `osctl_sensor_fan_rpm`)
- **`boot`**: Boot time, slowest units and the critical chain (systemd-analyze)
- **`certs [path|host:port ...]`**: TLS certificate expiry check — scans common system locations by default; reports OK / EXPIRING SOON (< 30 days) / EXPIRED with a gauge (`osctl_cert_expiry_timestamp_seconds`)
- **`watch [--interval SECONDS] <command>`**: Re-run a command on an interval (default 60 s) and POST to `OSCTL_WEBHOOK_URL` when the output changes
- **`completion [bash|zsh|fish]`**: Shell completion scripts with install hints

## 🛡️ Security Audit Extensions

- **`audit sysctl`**: Kernel hardening check (ASLR, kptr_restrict, dmesg_restrict, hardlink/symlink protection, suid_dumpable, reverse path filtering, ICMP redirects, ip_forward) with PASS/WARN/INFO per setting and `osctl_sysctl_compliance` gauges
- **`audit mac`**: SELinux (getenforce / `/sys/fs/selinux/enforce`) and AppArmor (kernel parameter + `aa-status`) status

## 🌐 API & Packaging

- **Live OpenAPI 3.0.3 document** at `/openapi.json` (no auth) — also checked in at [`docs/openapi.yaml`](docs/openapi.yaml)
- **Versioned `/v1/` route prefix** (e.g. `/v1/ram`) so clients can pin to a stable API path
- **`OSCTL_METRICS_AUTH`**: Require authentication on `/metrics` when set (e.g. `1`)
- **deb/rpm packages** now published by GoReleaser for amd64 and arm64, including the systemd service file
- The `logs` command returns a usage error (HTTP 400) instead of a server error when no unit is given

## 📦 Installation

```bash
# Debian/Ubuntu
wget https://github.com/diceone/osctl/releases/download/v0.3.2/osctl_0.3.2_amd64.deb
sudo apt install ./osctl_0.3.2_amd64.deb

# RHEL/CentOS/Fedora/SUSE
wget https://github.com/diceone/osctl/releases/download/v0.3.2/osctl_0.3.2_amd64.rpm
sudo rpm -i osctl_0.3.2_amd64.rpm

# Or download the binary
wget https://github.com/diceone/osctl/releases/download/v0.3.2/osctl_0.3.2_linux_amd64.tar.gz
tar xzf osctl_0.3.2_linux_amd64.tar.gz
chmod +x osctl
sudo mv osctl /usr/local/bin/osctl

# Or build from source
git clone https://github.com/diceone/osctl.git
cd osctl
git checkout v0.3.2
go build -o osctl .
```

Full Changelog: https://github.com/diceone/osctl/compare/v0.3.1...v0.3.2

---

# Release Notes v0.3.1

## 🔒 Better Privilege Reporting (issue #26)

Commands that require root now clearly say so instead of only reporting failure:

- **Root pre-checks** for `service start/stop/restart/enable/disable`, `shutdown`, `reboot`, `update`, `useradd`, `userdel`
- **Permission-error hints** appended for `audit ssh`, `audit files`, `audit summary`, and `errors` when run unprivileged
- Messages keep the `Failed...` prefix so exit codes and HTTP status mapping are unchanged

## 📖 CLI Improvements

- **`help`**: `osctl help` is now a documented command alongside `--help`

## 🛡️ Security Audit Improvements

- **`audit summary`**: output grouped by category — `Ports:`, `Users:`, `Files:`, `System:` — and an unreadable auth log is reported (with sudo hint) instead of a silent `0`
- **`audit files`**: distinguishes `(none found)` from incomplete scans caused by permission denied, with a note to re-run with sudo for a full scan
- **`audit permissions`**: prints three-digit octal notation (`644`) matching the expected values
- **`audit users`**: lists each user's UID, e.g. `User: root (uid: 0)`
- **`audit ssh`**: adds a legend for the ✅/⚠️/❓ markers and notes when the sshd default applies

## 📦 Installation

```bash
# Download binary for your platform
wget https://github.com/diceone/osctl/releases/download/v0.3.1/osctl_0.3.1_linux_amd64.tar.gz
tar xzf osctl_0.3.1_linux_amd64.tar.gz
chmod +x osctl
sudo mv osctl /usr/local/bin/osctl

# Or build from source
git clone https://github.com/diceone/osctl.git
cd osctl
git checkout v0.3.1
go build -o osctl .
```

Full Changelog: https://github.com/diceone/osctl/compare/v0.3.0...v0.3.1

---

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
