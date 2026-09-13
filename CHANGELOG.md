# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.3.3] — 2026-09-13

### Fixed
- systemd service file: `ExecStart` now points to `/usr/bin/osctl`, the path used by the deb/rpm packages (v0.3.2 shipped the service file pointing at `/root/osctl`, which failed to start after a package install); obsolete `WorkingDirectory` removed

## [0.3.2] — 2026-09-13

### Added
- `updates` — read-only list of available package updates (apt/dnf/yum/zypper)
- `logs <unit> [lines]` — recent journal entries for a systemd unit (default 50, max 10000)
- `dockerstats` — per-container CPU/memory stats with `osctl_docker_cpu_percent` and `osctl_docker_mem_bytes` gauges
- `sensors` — temperatures and fan speeds from `/sys/class/hwmon` with `osctl_sensor_temp_celsius` and `osctl_sensor_fan_rpm` gauges
- `boot` — boot time, slowest units and critical chain (systemd-analyze)
- `certs [path|host:port ...]` — TLS certificate expiry check (OK / EXPIRING SOON / EXPIRED) with `osctl_cert_expiry_timestamp_seconds` gauge
- `watch [--interval SECONDS] <command>` — re-run a command on an interval and POST to `OSCTL_WEBHOOK_URL` when the output changes
- `completion [bash|zsh|fish]` — shell completion scripts with install hints
- `audit sysctl` — kernel hardening checks with `osctl_sysctl_compliance` gauges
- `audit mac` — SELinux/AppArmor status
- Live OpenAPI 3.0.3 document served at `/openapi.json` (no auth)
- Versioned `/v1/` route prefix (e.g. `/v1/ram`)
- `OSCTL_METRICS_AUTH` — require authentication on `/metrics` when set
- deb/rpm packages (amd64 + arm64) built by GoReleaser, including the systemd service file

### Fixed
- `logs` with no unit returns a usage error (HTTP 400) instead of a server error

## [0.3.1] — 2026-09-13

### Added
- Root pre-checks for `service start/stop/restart/enable/disable`, `shutdown`, `reboot`, `update`, `useradd`, `userdel`
- Permission-error hints appended for `audit ssh`, `audit files`, `audit summary`, and `errors` when run unprivileged

### Improved
- `audit summary` grouped by category; unreadable auth log reported with sudo hint instead of a silent `0`
- `audit files` distinguishes "(none found)" from permission-limited scans
- `audit permissions` prints three-digit octal notation; `audit users` lists UIDs
- `audit ssh` adds a marker legend
- `osctl help` documented as a command alongside `--help`

## [0.3.0] — 2026-08-31

### Added
- `version` command and `--json` output flag for all CLI commands
- HTTPS support via `OSCTL_TLS_CERT` / `OSCTL_TLS_KEY`
- Bearer token authentication via `OSCTL_API_TOKEN`
- JSONL request audit log via `OSCTL_AUDIT_LOG`
- KEY=VALUE config file via `OSCTL_CONFIG` (environment variables take precedence)
- Auth-failure rate limits persisted to `OSCTL_STATE_DIR` across restarts
- `dockerlogs` / `dockerrestart` with container-name validation
- `userinfo` / `useradd` / `userdel` with username validation and protected accounts
- `firewallallow` / `firewalldeny` (ufw/firewalld) with port validation
- `doctor` diagnostic command and API endpoint
- Health-change webhooks (`OSCTL_WEBHOOK_URL`, `OSCTL_HEALTH_INTERVAL`)

### Changed
- Switched to GoReleaser for release builds and CI builds with `go build .`

## [0.2.0] — 2026-08-31

### Security
- Constant-time basic auth comparison and per-IP rate limiting (10 failures / 5 min → 429)
- POST-only `shutdown` / `reboot` endpoints (GET returns 405)
- PID validation rejects negative/oversized values; no process-group kills
- Cron schedule/command validators block newline/control-character injection
- HTTP server timeouts, port validation, mux isolation, default-password warning

### Fixed
- Meaningful HTTP status codes (200/400/404/500) mapped from result strings
- Real output for service status; intermediate package-command errors surfaced
- Health checks: overall status can no longer be less severe than a single check
- Security audit: broken `find` arguments, stderr noise, RHEL secure log path
- Maintenance flag moved from `/tmp` to `/run/osctl` (0600, corrupt-JSON safe)
- `top`: instantaneous CPU delta instead of lifetime averages
- Working multi-stage Dockerfile; CLI exits non-zero on failure

### Added
- Unit tests for auth, handlers, PID validation, cron validation and metrics

## [0.0.8] — 2026-01-01

### Added
- `maintenance` command (system maintenance operations, service checks, cache clearing)

### Changed
- Go 1.23 compatibility fixes for SLSA release workflow

## [0.0.7] — 2026-01-01

### Added
- `health` endpoint with memory/disk/CPU thresholds
- Process management (`kill`, `nice`, `info`, `tree`)
- Extended Prometheus metrics (network I/O, disk I/O, process counts)
- Security audit tools (`ports`, `files`, `ssh`, `users`)
- Cron job management (list, add, remove, next runs)

## [0.0.6] — 2026-01-01

### Security
- Proper Basic Auth with `WWW-Authenticate` headers, applied to all API endpoints except `/metrics`
- Input validation for service parameters (command-injection prevention)

### Changed
- Environment variable configuration (`OSCTL_PORT`, `OSCTL_USERNAME`, `OSCTL_PASSWORD`)
- Replaced `log.Fatalf()` with error returns to prevent API server crashes
- OS detection via `/etc/os-release` parsing

## [0.0.5] — 2024-07-23

### Changed
- Release tooling (SLSA GoReleaser configuration)

## [0.0.4] — 2024-05-28

### Changed
- Codebase restructured from a single file into multiple source files
- GitHub Actions release workflow

## [0.0.3] — 2024-05-24

### Added
- RPM packaging specification (`osctl.spec`)

## [0.0.2] — 2024-05-22

### Changed
- SLSA release workflow updates

## [0.0.1] — 2024-05-22

### Added
- Initial release: CLI for Linux system administration (RAM, disk, CPU, services, Docker, network, shutdown/reboot) with HTTP API server and Prometheus metrics

[Unreleased]: https://github.com/diceone/osctl/compare/v0.3.3...HEAD
[0.3.3]: https://github.com/diceone/osctl/compare/v0.3.2...v0.3.3
[0.3.2]: https://github.com/diceone/osctl/compare/v0.3.1...v0.3.2
[0.3.1]: https://github.com/diceone/osctl/compare/v0.3.0...v0.3.1
[0.3.0]: https://github.com/diceone/osctl/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/diceone/osctl/compare/v0.0.8...v0.2.0
[0.0.8]: https://github.com/diceone/osctl/compare/v0.0.7...v0.0.8
[0.0.7]: https://github.com/diceone/osctl/compare/v0.0.6...v0.0.7
[0.0.6]: https://github.com/diceone/osctl/compare/v0.0.5...v0.0.6
[0.0.5]: https://github.com/diceone/osctl/compare/v0.0.4...v0.0.5
[0.0.4]: https://github.com/diceone/osctl/compare/v0.0.3...v0.0.4
[0.0.3]: https://github.com/diceone/osctl/compare/v0.0.2...v0.0.3
[0.0.2]: https://github.com/diceone/osctl/compare/v0.0.1...v0.0.2
[0.0.1]: https://github.com/diceone/osctl/releases/tag/v0.0.1