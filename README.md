# osctl

`osctl` is a command-line tool for Linux system administration. It provides various commands to monitor and manage the system, including checking RAM and disk usage, managing services, viewing top processes, checking system logs, and more. Additionally, `osctl` can run as an API server to provide system information through HTTP endpoints.

## Features

- Check RAM and disk usage
- Manage system services
- View top processes by CPU usage
- Check system logs
- List last logged-in users
- Show system uptime
- Retrieve OS and kernel information
- Shutdown and reboot the system
- List IP addresses of all interfaces
- Show active firewalld rules
- Update OS packages
- List Docker containers and images
- Show CPU usage
- Show system load averages
- Show network statistics
- List active network connections
- List mounted filesystems
- Retrieve kernel messages
- List currently logged-in users
- Show status of all running services
- Run as an API server

## Installation

### Prerequisites

- Go (version 1.16 or higher)
- Docker (optional, for Docker-related commands)
- Systemd (for service management)

### Building from Source

1. Clone the repository:

    ```bash
    git clone https://github.com/yourusername/osctl.git
    cd osctl
    ```

2. Build the binary:

    ```bash
    go build -o osctl osctl.go
    ```

3. (Optional) Install the binary to `/usr/local/bin`:

    ```bash
    sudo mv osctl /usr/local/bin/
    ```

### Running as a Service

1. Create a systemd service unit file (`/etc/systemd/system/osctl.service`):

    ```ini
    [Unit]
    Description=osctl API server
    After=network.target

    [Service]
    ExecStart=/usr/local/bin/osctl api
    Restart=always
    User=nobody
    Group=nogroup

    [Install]
    WantedBy=multi-user.target
    ```

2. Reload systemd and start the service:

    ```bash
    sudo systemctl daemon-reload
    sudo systemctl start osctl
    sudo systemctl enable osctl
    ```

## Usage

### Command-Line Interface

Run `osctl` with one of the following commands:

```bash
osctl [command]
```

Available commands:

- `ram`: Show RAM usage
- `disk`: Show disk usage
- `service`: Manage system services
  - Usage: `osctl service [start|stop|restart|status] [service_name]`
- `top`: Show top processes by CPU usage
- `errors`: Show last 10 errors from the journal
- `users`: Show last 20 logged-in users
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

### Examples

Check RAM usage:

```bash
osctl ram
```

Check disk usage:

```bash
osctl disk
```

Start a service:

```bash
osctl service start nginx
```

Stop a service:

```bash
osctl service stop nginx
```

List top processes by CPU usage:

```bash
osctl top
```

Show last 10 journal errors:

```bash
osctl errors
```

Show last 20 logged-in users:

```bash
osctl users
```

Show system uptime:

```bash
osctl uptime
```

Show OS and kernel information:

```bash
osctl osinfo
```

Shutdown the system:

```bash
osctl shutdown
```

Reboot the system:

```bash
osctl reboot
```

Show IP addresses of all interfaces:

```bash
osctl ip
```

Show active firewalld rules:

```bash
osctl firewall
```

Update OS packages:

```bash
osctl update
```

List Docker containers:

```bash
osctl containers
```

List Docker images:

```bash
osctl images
```

Show CPU usage:

```bash
osctl cpu
```

Show system load averages:

```bash
osctl load
```

Show network statistics:

```bash
osctl network
```

List active network connections:

```bash
osctl connections
```

List mounted filesystems:

```bash
osctl filesystems
```

Show kernel messages:

```bash
osctl dmesg
```

List currently logged-in users:

```bash
osctl who
```

Show status of all running services:

```bash
osctl services
```

### Running as an API Server

Start the API server on port 12000:

```bash
osctl api
```

You can then access the endpoints using HTTP requests. For example, to get RAM usage:

```bash
curl -u admin:password http://localhost:12000/ram
```

### License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

### Contributing

Contributions are welcome! Please open an issue or submit a pull request on GitHub.

### Author

Michael Vogeler (diceone)
```

This README includes an overview of the project, installation and usage instructions, and a list of all available commands with examples. Adjust the author information and repository URL as needed.
