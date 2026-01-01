# osctl

`osctl` is a command-line interface (CLI) tool for administrating Linux operating systems like RHEL, Ubuntu, and SUSE. It provides easy access to system statistics like RAM usage, disk usage, service management, and more. Additionally, it can run as an API server with a Prometheus metrics endpoint.

## Features

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
- `containers`: List all Docker containers
- `images`: List all Docker images
- `cpu`: Show CPU usage
- `load`: Show system load averages
- `network`: Show network statistics
- `connections`: List all active network connections
- `filesystems`: List all mounted filesystems
- `dmesg`: Show kernel messages
- `who`: List all currently logged-in users
- `services`: Show status of all running services
- `api`: Run as an API server (default port: 12000)
- `--help`: Show this help message

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
   go build -o osctl main.go auth.go metrics.go handlers.go system_info.go services.go
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

Example:
```bash
export OSCTL_PORT=8080
export OSCTL_USERNAME=myuser
export OSCTL_PASSWORD=securepassword
./osctl api
```

The API server provides the same functionalities as the CLI commands. Additionally, it includes a **public** Prometheus metrics endpoint at `/metrics` (no authentication required).

## Authentication for API

The API uses Basic Authentication for all endpoints except `/metrics`. 

**Default credentials:**
- Username: `admin`
- Password: `password`

**⚠️ Security Warning:** Change the default credentials using environment variables in production environments!

### API Usage Examples

Query RAM usage:
```bash
curl -u admin:password http://localhost:12000/ram
```

Manage a service:
```bash
curl -u admin:password "http://localhost:12000/service?action=status&service=nginx"
```

Access Prometheus metrics (no auth required):
```bash
curl http://localhost:12000/metrics
```

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
   - Use firewall rules to restrict access to your Prometheus server
   - Consider implementing IP whitelisting
   - Place behind a reverse proxy with authentication

4. **Input validation**: The service management commands include validation to prevent command injection, but always:
   - Sanitize inputs when integrating with other systems
   - Monitor logs for suspicious activity
   - Use restricted service accounts where possible

### Supported Operating Systems

- **RHEL/CentOS/Fedora**: Full support with yum/dnf package management
- **Ubuntu/Debian**: Full support with apt package management  
- **SUSE/openSUSE**: Full support with zypper package management

OS detection uses `/etc/os-release` (modern standard) with fallback to legacy detection files.

## Contributing

Feel free to submit issues, fork the repository, and send pull requests. For major changes, please open an issue first to discuss what you would like to change.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
```
