# osctl

`osctl` is a command-line interface (CLI) tool for administrating Linux operating systems like RHEL, Ubuntu, and SUSE. It provides easy access to system statistics like RAM usage, disk usage, service management, and more. Additionally, it can run as an API server with a Prometheus metrics endpoint.

## Features

- Show RAM usage
- Show disk usage
- Manage system services
- Show top processes by CPU usage
- Show the last 10 errors from the journal
- Show the last 20 logged-in users
- Show system uptime
- Show operating system name and kernel version
- Shutdown the system
- Reboot the system
- Show IP addresses of all interfaces
- Show active firewalld rules
- Update OS packages
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
- Run as an API server on port 12000 with a Prometheus metrics endpoint

## Usage

```bash
osctl [command]
```

### Commands

- `ram`: Show RAM usage
- `disk`: Show disk usage
- `service [start|stop|restart|status] [service_name]`: Manage system services
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
- `api`: Run as an API server on port 12000
- `--help`: Show this help message

## Installation

### Building from Source

1. Ensure you have Go 1.18 or later installed.
2. Clone the repository:

   ```bash
   git clone https://github.com/yourusername/osctl.git
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

## Running as an API Server

To run `osctl` as an API server on port 12000, use the `api` command:

```bash
./osctl api
```

The API server provides the same functionalities as the CLI commands. Additionally, it includes a Prometheus metrics endpoint at `/metrics`.

## Authentication for API

Basic authentication is used for the API server. The default credentials are:

- Username: `admin`
- Password: `password`

You can modify these credentials in the `auth.go` file.

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

## Contributing

Feel free to submit issues, fork the repository, and send pull requests. For major changes, please open an issue first to discuss what you would like to change.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
```
