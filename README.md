# osctl

`osctl` is a command-line tool for Linux system administration. It provides various commands to monitor and manage the system, including checking RAM and disk usage, managing services, viewing top processes, checking system logs, and more. Additionally, `osctl` can run as an API server on port 12000.

## Features

- Show RAM usage
- Show disk usage
- Manage system services (start, stop, restart, status)
- Show top processes by CPU usage
- Show the last 10 errors from the journal
- Show the last 20 logged-in users
- Show system uptime
- Show operating system name and kernel version
- Shutdown the system
- Reboot the system
- Show IP addresses of all network interfaces
- Show active firewalld rules
- Update OS packages
- List all Docker containers
- List all Docker images

## Usage


### Building and Running the Go Program

Follow these steps to build and run the Go program:

1. **Clone the repository** (if you haven't already):

   ```bash
   git clone https://github.com/diceone/osctl.git
   cd osctl
   ```

2. **Initialize Go modules** (if not already initialized):

   ```bash
   go mod init github.com/diceone/osctl
   ```

3. **Download dependencies**:

   ```bash
   go mod tidy
   ```

4. **Build the `osctl` binary**:

   ```bash
   go build -o osctl osctl.go
   ```

5. **Run the tool**:

   ```bash
   ./osctl [command]
   ```

   Replace `[command]` with the command you want to execute, such as `ram`, `disk`, `service`, `top`, etc.


API Endpoints
Show RAM usage:

```bash

GET /ram
```
Show disk usage:

```bash

GET /disk
```
Manage system services:

```css

GET /service?action=[start|stop|restart|status]&service=[service_name]
```
Show top processes:

```bash

GET /top
```
Show the last errors from the journal:

```bash

GET /errors
```
Show the last 20 logged-in users:

```bash

GET /users
```
Show system uptime:

```bash

GET /uptime
```
Show operating system name and kernel version:

```bash

GET /osinfo
```
Shutdown the system:

```bash

GET /shutdown
```
Reboot the system:

```bash

GET /reboot
```

Show IP addresses of all interfaces:

```bash

GET /ip
```

Show active firewalld rules:

```bash

GET /firewall
```

Update OS packages:

``` bash

GET /update
```

List all Docker containers:

```bash

GET /containers
```
List all Docker images:

``` bash

GET /images
```
License
This project is licensed under the MIT License - see the LICENSE file for details.

Contributing
Contributions are welcome! Please feel free to submit a Pull Request.
