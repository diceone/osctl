# osctl

`osctl` is a command-line interface (CLI) tool for administrating Linux operating systems like RHEL, Ubuntu, and SUSE. It provides easy access to system statistics like RAM usage, disk usage, service management, and more. Additionally, it can run as an API server with a Prometheus metrics endpoint.

## Features

- **Version command** (`version`) and **JSON output** for all CLI commands (`--json`)
- **HTTPS/TLS support** for the API server
- **API token auth** (Bearer) as alternative to Basic auth
- **Request audit log** (JSONL) for the API
- **Config file support** (`OSCTL_CONFIG`)
- **Persistent rate limiting** across restarts
- **Docker container logs and restart**, **user management**, **firewall rule management**
- **Doctor command** for one-shot diagnostics and **health-change webhooks**
- Show RAM usage
- Show disk usage
- Manage system services (start, stop, restart, status, enable, disable)
- Show top processes by CPU usage
- Show the last 10 errors from the journal
- Show the last 20 logged-in users
- Show system uptime
- Show operating system name and kernel version
- Shutdown the system
- Reboot the system
- Show IP addresses of all interfaces
- Show active firewalld rules
- Update OS packages (RHEL/CentOS/Fedora, Ubuntu/Debian, SUSE/openSUSE)
- List all Docker containers
- List all Docker images
- Show CPU usage
- Show system load averages
- Show network statistics
- List all active network connections
- List all mounted filesystems
- Show kernel messages
- List all currently logged-in users
- Show status of all running services
- **Health check endpoint** for monitoring
- **Process management** (kill, nice, info, tree)
- **Extended metrics** (Network I/O, Disk I/O, Process counts)
- **Security audit** (port scan, file permissions, SSH config, suspicious files)
- **Cron job management** (list, add, remove, next runs)
- **Maintenance mode** (system maintenance operations, service checks, cache clearing)
- **Read-only package update list** (`updates`) and **service journal logs** (`logs <unit> [lines]`)
- **Docker container stats** (`dockerstats`) with Prometheus gauges
- **Hardware sensors** (`sensors`), **boot analysis** (`boot`), **TLS certificate expiry** (`certs`)
- **Diagnostics**: failed units (`failed`), listening ports (`ports`), hardware info (`hwinfo`), time sync (`timesync`), full JSON snapshot (`report`)
- **Watch mode** (`watch`) re-running a command on an interval with change webhooks
- **Extended security audit**: kernel hardening (`audit sysctl`) and SELinux/AppArmor status (`audit mac`)
- **Shell completions** (bash/zsh/fish) and **deb/rpm packages** via GoReleaser
- **OpenAPI document** served live at `/openapi.json`, versioned `/v1/` API routes, optional `/metrics` authentication
- Run as an API server with configurable port and Prometheus metrics endpoint

## Usage

```bash
osctl [command]
```

### Commands

- `ram`: Show RAM usage
- `disk`: Show disk usage
- `service [start|stop|restart|status|enable|disable] [service_name]`: Manage system services
- `top`: Show top processes by CPU usage
- `errors`: Show the last 10 errors from the journal
- `users`: Show the last 20 logged-in users
- `uptime`: Show system uptime
- `osinfo`: Show operating system name and kernel version
- `shutdown`: Shutdown the system
- `reboot`: Reboot the system
- `ip`: Show IP addresses of all interfaces
- `firewall`: Show active firewalld rules
- `update`: Update OS packages
- `updates`: List available package updates without installing them
- `containers`: List all Docker containers
- `images`: List all Docker images
- `dockerstats`: Show per-container CPU and memory stats
- `cpu`: Show CPU usage
- `load`: Show system load averages
- `network`: Show network statistics
- `connections`: List all active network connections
- `filesystems`: List all mounted filesystems
- `dmesg`: Show kernel messages
- `who`: List all currently logged-in users
- `services`: Show status of all running services
- `health`: Show system health check status
- `process [action]`: Process management
  - `kill <pid>`: Terminate process
  - `killforce <pid>`: Force kill process
  - `nice <pid> <priority>`: Set process priority (-20 to 19)
  - `info <pid>`: Show detailed process information
  - `tree`: Show process tree
- `networkio`: Show network I/O statistics with rates
- `diskio`: Show disk I/O statistics
- `procs`: Show process count by state
- `audit [action]`: Security audit tools
  - `ports`: List open listening ports
  - `files`: Check for suspicious file permissions
  - `permissions`: Check critical file permissions
  - `users`: List user accounts and last login
  - `ssh`: Audit SSH configuration
  - `sysctl`: Audit kernel hardening settings (ASLR, kptr_restrict, dump protection, ...)
  - `mac`: Show SELinux/AppArmor status
  - `summary`: Security audit summary
- `cron [action]`: Cron job management
  - `list`: List all cron jobs with line numbers
  - `add "<schedule>" "<command>"`: Add new cron job
  - `remove <line>`: Remove cron job by line number
  - `next`: Show next scheduled runs (systemd timers)
- `maintenance [action]`: Maintenance mode and system operations
  - `status`: Show maintenance mode status
  - `enable`: Enable maintenance mode (broadcasts message to users)
  - `disable`: Disable maintenance mode
  - `check-services`: Check status of critical services
  - `restart-failed`: Restart all failed systemd services
  - `sync-time`: Synchronize system time via NTP
  - `clear-cache`: Clear system caches and old journal logs
- `doctor`: One-shot diagnostic report (health, failed services, disk, journal errors, updates)
- `dockerlogs <container>`: Show last 50 log lines of a container
- `dockerrestart <container>`: Restart a container
- `userinfo <username>`: Show user identity and password aging
- `useradd <username>`: Create a user with home directory
- `userdel <username>`: Delete a user (and home directory)
- `firewallallow <port>[/<proto>]`: Allow a port (ufw or firewalld)
- `firewalldeny <port>[/<proto>]`: Deny/remove a port rule
- `logs <unit> [lines]`: Show recent journal entries for a systemd unit (default 50 lines)
- `boot`: Show boot time, slowest units and the critical chain (systemd-analyze)
- `sensors`: Show temperatures and fan speeds from `/sys/class/hwmon`
- `certs [path|host:port ...]`: Check TLS certificate expiry (scans common system locations by default)
- `failed`: List failed systemd units (with `osctl_failed_units` gauge)
- `ports`: List listening TCP ports and their owning processes (with `osctl_listening_port` gauge)
- `hwinfo`: Show PCI devices, USB devices and loaded kernel modules
- `timesync`: Show clock synchronization status (timedatectl, falls back to chrony)
- `report`: One-shot JSON snapshot of all read-only diagnostics
- `watch [--interval SECONDS] <command>`: Re-run a command on an interval (default 60) and POST to `OSCTL_WEBHOOK_URL` when the output changes
- `completion [bash|zsh|fish]`: Print a shell completion script
- `version`: Show osctl version
- `help`: Show this help message
- `api`: Run as an API server (default port: 12000)
- `--json`: Prefix for any command to emit the result as JSON (`osctl --json <command>`)
- `--help`: Show this help message

CLI commands exit with status `0` on success and `1` on failure (invalid arguments, unknown command, or command error), so they can be used safely in scripts.

## Installation

### Building from Source

1. Ensure you have Go 1.20 or later installed.
2. Clone the repository:

   ```bash
   git clone https://github.com/diceone/osctl.git
   cd osctl
   ```

3. Build the binary:

   ```bash
   go build -o osctl .
   ```

4. Run the `osctl` binary:

   ```bash
   ./osctl --help
   ```

### Running with systemd

For production deployments, you can run `osctl` as a systemd service. See the `systemd/` directory for service files and installation instructions.

Example systemd service configuration:
```bash
sudo cp osctl /usr/local/bin/
sudo cp systemd/osctl.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable osctl
sudo systemctl start osctl
```

For detailed instructions, see [systemd/README.md](systemd/README.md).

## Running as an API Server

To run `osctl` as an API server, use the `api` command:

```bash
./osctl api
```

By default, the API server listens on port 12000. You can customize the port and authentication credentials using environment variables.

### Configuration

Configure the API server using environment variables:

- `OSCTL_PORT`: Server port (default: `12000`)
- `OSCTL_USERNAME`: Basic auth username (default: `admin`)
- `OSCTL_PASSWORD`: Basic auth password (default: `password`)
- `OSCTL_API_TOKEN`: Bearer token as alternative to Basic auth (set to disable the password warning)
- `OSCTL_TLS_CERT` / `OSCTL_TLS_KEY`: PEM certificate/key to serve the API over HTTPS
- `OSCTL_AUDIT_LOG`: Path to a JSONL file where every API request (method, path, status, user, source IP) is logged
- `OSCTL_STATE_DIR`: Directory for persisted auth-failure state (default: `/var/lib/osctl`)
- `OSCTL_WEBHOOK_URL`: URL that receives a JSON POST whenever the health status changes (also used by `watch` for output changes)
- `OSCTL_HEALTH_INTERVAL`: Seconds between health checks for the webhook monitor (default: `300`, minimum: `30`)
- `OSCTL_METRICS_AUTH`: Require authentication on `/metrics` when set (e.g. `1`)

#### Config file

Any of these variables can be set in a `KEY=VALUE` config file passed via `OSCTL_CONFIG`.
Environment variables take precedence over the config file; unknown keys are ignored.

```bash
# /etc/osctl/osctl.conf
OSCTL_PORT=12000
OSCTL_USERNAME=ops
OSCTL_PASSWORD=correct-horse-battery-staple
OSCTL_API_TOKEN=...
OSCTL_TLS_CERT=/etc/osctl/tls.crt
OSCTL_TLS_KEY=/etc/osctl/tls.key
OSCTL_AUDIT_LOG=/var/log/osctl/audit.jsonl

export OSCTL_CONFIG=/etc/osctl/osctl.conf
./osctl api
```

Example:
```bash
export OSCTL_PORT=8080
export OSCTL_USERNAME=myuser
export OSCTL_PASSWORD=securepassword
./osctl api
```

The API server provides the same functionalities as the CLI commands. Additionally, it includes a Prometheus metrics endpoint at `/metrics` (no authentication required unless `OSCTL_METRICS_AUTH` is set) and a public OpenAPI document at `/openapi.json`.

## Authentication for API

The API uses Basic Authentication **or** a Bearer token for all endpoints except `/metrics`. Set `OSCTL_API_TOKEN` to enable token authentication:

```bash
export OSCTL_API_TOKEN=my-secret-token
curl -H "Authorization: Bearer my-secret-token" http://localhost:12000/ram
```

Basic auth continues to work alongside the token. 

**Default credentials:**
- Username: `admin`
- Password: `password`

**⚠️ Security Warning:** Change the default credentials using environment variables in production environments! The server logs a warning at startup when `OSCTL_PASSWORD` is not set.

Repeated failed login attempts from the same address are rate limited: after 10 failures within 5 minutes, further requests receive `429 Too Many Requests` until the window expires.

### API Usage Examples

Query RAM usage:
```bash
curl -u admin:password http://localhost:12000/ram
```

Manage a service:
```bash
curl -u admin:password "http://localhost:12000/service?action=status&service=nginx"
```

Shutdown the system (**POST required** — destructive endpoints reject `GET` with `405`):
```bash
curl -u admin:password -X POST http://localhost:12000/shutdown
```

Access Prometheus metrics (no auth required):
```bash
curl http://localhost:12000/metrics
```

### HTTP status codes

Responses use meaningful status codes: `200` success, `400` invalid input, `401` missing/invalid credentials, `405` wrong method (e.g. `GET /shutdown`), `429` rate-limited, `500` command failure. The body is always `{"result": "..."}`.
Failed authentication attempts are persisted to `OSCTL_STATE_DIR` so the rate-limit
survives restarts. When `OSCTL_AUDIT_LOG` is set, every request is appended as a JSONL entry:
`{"time":"...","ip":"...","method":"GET","path":"/ram","status":200,"user":"admin"}`.

### HTTPS / TLS

Set both variables to serve the API over HTTPS:

```bash
export OSCTL_TLS_CERT=/etc/osctl/tls.crt
export OSCTL_TLS_KEY=/etc/osctl/tls.key
./osctl api
# Server is listening on port 12000 (https)...
```

### Health webhooks

With `OSCTL_WEBHOOK_URL` set, osctl checks its health status in the background
(`OSCTL_HEALTH_INTERVAL`, default 5 minutes) and POSTs a JSON payload to the webhook
URL whenever the overall status changes (e.g. `healthy` → `unhealthy`), together with
the individual checks. Delivery failures are logged but never crash the server.

```bash
export OSCTL_WEBHOOK_URL=https://hooks.example.com/osctl
export OSCTL_HEALTH_INTERVAL=60
./osctl api
```

### OpenAPI

A machine-readable API description is served live at `/openapi.json` (no
authentication required) and is also checked into the repository at
[`docs/openapi.yaml`](docs/openapi.yaml). Either can be imported into Swagger UI,
Postman, or API gateways. All routes are additionally available under a `/v1/`
prefix (e.g. `/v1/ram`) so clients can pin to a stable API path.

### Releases

Releases are built with [GoReleaser](https://goreleaser.com) via GitHub Actions
(`.github/workflows/release.yml`): pushing a `v*` tag runs the tests and publishes
Linux binaries (amd64/arm64) plus checksums to the GitHub release, and build
`deb`/`rpm` packages (including the systemd service file). The version is
injected from the tag via `-ldflags -X main.buildVersion`.

## Example Usage

Show RAM usage:

```bash
./osctl ram
```

Show disk usage:

```bash
./osctl disk
```

Start a service:

```bash
./osctl service start apache2
```

Show top processes by CPU usage:

```bash
./osctl top
```

Update OS packages:

```bash
./osctl update
```

Check system health:

```bash
./osctl health
```

Kill a process:

```bash
./osctl process kill 1234
```

Security audit:

```bash
./osctl audit summary
./osctl audit ports
./osctl audit ssh
./osctl audit sysctl
./osctl audit mac
```

List available package updates (without installing):

```bash
./osctl updates
```

Show journal entries for a service:

```bash
./osctl logs nginx 100
```

Show per-container Docker stats:

```bash
./osctl dockerstats
```

Show hardware sensors and boot analysis:

```bash
./osctl sensors
./osctl boot
```

Check TLS certificate expiry (system locations by default, or explicit targets):

```bash
./osctl certs
./osctl certs /etc/letsencrypt/live/example.com/fullchain.pem
./osctl certs example.com:443
```

Run quick diagnostics (also available as API routes `/failed`, `/ports`, `/hwinfo`, `/timesync`, `/report`):

```bash
./osctl failed
./osctl ports
./osctl hwinfo
./osctl timesync
curl -u admin:password http://localhost:12000/report | jq .sections.failed_units
```

Watch a command on an interval (webhooks on output change):

```bash
export OSCTL_WEBHOOK_URL=https://hooks.example.com/osctl
./osctl watch --interval 30 health
```

Install shell completions:

```bash
./osctl completion bash | sudo tee /etc/bash_completion.d/osctl
```

Query the same information through the API:

```bash
curl -u admin:password http://localhost:12000/v1/ram
curl -u admin:password "http://localhost:12000/logs?unit=nginx&lines=100"
curl http://localhost:12000/openapi.json
```

Manage cron jobs:

```bash
./osctl cron list
./osctl cron add "0 2 * * *" "/backup.sh"
```

Maintenance operations:

```bash
./osctl maintenance status
./osctl maintenance enable
./osctl maintenance check-services
./osctl maintenance restart-failed
./osctl maintenance sync-time
./osctl maintenance clear-cache
./osctl maintenance disable
```

## Security Considerations

### Production Deployment

When deploying `osctl` in production, follow these security best practices:

1. **Change default credentials**: Always set custom username and password via environment variables
   ```bash
   export OSCTL_USERNAME=your_secure_username
   export OSCTL_PASSWORD=your_secure_password
   ```

2. **Run as root**: Most system commands require root privileges. The API server should run as root, but consider:
   - Using a reverse proxy (nginx/Apache) for SSL/TLS termination
   - Implementing additional authentication layers (OAuth, JWT)
   - Restricting network access via firewall rules

3. **Metrics endpoint**: The `/metrics` endpoint is public by default for Prometheus scraping. To secure it:
   - Set `OSCTL_METRICS_AUTH=1` to require authentication (supported natively)
   - Or use firewall rules to restrict access to your Prometheus server
   - Or place behind a reverse proxy with authentication

4. **Input validation**: The service management commands include validation to prevent command injection, but always:
   - Sanitize inputs when integrating with other systems
   - Monitor logs for suspicious activity
   - Use restricted service accounts where possible

### Supported Operating Systems

- **RHEL/CentOS/Fedora**: Full support with yum/dnf package management
- **Ubuntu/Debian**: Full support with apt package management  
- **SUSE/openSUSE**: Full support with zypper package management

OS detection uses `/etc/os-release` (modern standard) with fallback to legacy detection files.

## Changelog

All notable changes are documented in [CHANGELOG.md](CHANGELOG.md) (Keep a Changelog format).

## Contributing

Feel free to submit issues, fork the repository, and send pull requests. For major changes, please open an issue first to discuss what you would like to change.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
```
