package main

import (
	"encoding/json"
	"fmt" // Added import
	"log"
	"net/http"
	"strings"
)

func handleRequest(w http.ResponseWriter, r *http.Request) {
	path := strings.Trim(r.URL.Path, "/")

	var result string

	switch path {
	case "ram":
		result = getRamUsage()
	case "disk":
		result = getDiskUsage()
	case "service":
		action := r.URL.Query().Get("action")
		service := r.URL.Query().Get("service")
		if action == "" || service == "" {
			http.Error(w, "Missing action or service parameter", http.StatusBadRequest)
			return
		}
		// Validate service name length
		if len(service) > 256 {
			http.Error(w, "Service name too long", http.StatusBadRequest)
			return
		}
		result = manageService(action, service)
	case "top":
		result = getTopProcesses()
	case "errors":
		result = getLastJournalErrors()
	case "users":
		result = getLastLoggedUsers()
	case "uptime":
		result = getUptime()
	case "osinfo":
		result = getOSInfo()
	case "shutdown", "reboot":
		// Destructive endpoints must not be triggerable by a navigational
		// GET (browsers attach cached Basic Auth to top-level navigations).
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", "POST")
			http.Error(w, "Method not allowed. shutdown and reboot require POST", http.StatusMethodNotAllowed)
			return
		}
		if path == "shutdown" {
			result = shutdownSystem()
		} else {
			result = rebootSystem()
		}
	case "ip":
		result = getIPAddresses()
	case "firewall":
		result = getFirewalldRules()
	case "update":
		result = updatePackages()
	case "containers":
		result = listDockerContainers()
	case "images":
		result = listDockerImages()
	case "dockerlogs":
		container := r.URL.Query().Get("container")
		if container == "" {
			http.Error(w, "Missing container parameter", http.StatusBadRequest)
			return
		}
		result = dockerContainerLogs(container)
	case "dockerrestart":
		// Restarting a container is destructive: POST only.
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", "POST")
			http.Error(w, "Method not allowed. dockerrestart requires POST", http.StatusMethodNotAllowed)
			return
		}
		container := r.URL.Query().Get("container")
		if container == "" {
			http.Error(w, "Missing container parameter", http.StatusBadRequest)
			return
		}
		result = dockerRestartContainer(container)
	case "userinfo":
		user := r.URL.Query().Get("user")
		if user == "" {
			http.Error(w, "Missing user parameter", http.StatusBadRequest)
			return
		}
		result = getUserInfo(user)
	case "useradd":
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", "POST")
			http.Error(w, "Method not allowed. useradd requires POST", http.StatusMethodNotAllowed)
			return
		}
		user := r.URL.Query().Get("user")
		if user == "" {
			http.Error(w, "Missing user parameter", http.StatusBadRequest)
			return
		}
		result = addUser(user)
	case "userdel":
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", "POST")
			http.Error(w, "Method not allowed. userdel requires POST", http.StatusMethodNotAllowed)
			return
		}
		user := r.URL.Query().Get("user")
		if user == "" {
			http.Error(w, "Missing user parameter", http.StatusBadRequest)
			return
		}
		result = deleteUser(user)
	case "firewallallow":
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", "POST")
			http.Error(w, "Method not allowed. firewallallow requires POST", http.StatusMethodNotAllowed)
			return
		}
		result = manageFirewallRule("allow", r.URL.Query().Get("port"))
	case "firewalldeny":
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", "POST")
			http.Error(w, "Method not allowed. firewalldeny requires POST", http.StatusMethodNotAllowed)
			return
		}
		result = manageFirewallRule("deny", r.URL.Query().Get("port"))
	case "doctor":
		result = getDoctorReport()
	case "version":
		result = getVersion()
	case "cpu":
		result = getCpuUsage()
	case "load":
		result = getLoadAverage()
	case "network":
		result = getNetworkStats()
	case "connections":
		result = getActiveConnections()
	case "filesystems":
		result = getMountedFilesystems()
	case "dmesg":
		result = getKernelMessages()
	case "who":
		result = getLoggedinUsers()
	case "services":
		result = getServiceStatuses()
	case "health":
		result = getHealthCheck()
	case "process":
		action := r.URL.Query().Get("action")
		pid := r.URL.Query().Get("pid")
		priority := r.URL.Query().Get("priority")

		switch action {
		case "kill":
			if pid == "" {
				http.Error(w, "Missing pid parameter", http.StatusBadRequest)
				return
			}
			result = killProcess(pid)
		case "killforce":
			if pid == "" {
				http.Error(w, "Missing pid parameter", http.StatusBadRequest)
				return
			}
			result = killProcessForce(pid)
		case "nice":
			if pid == "" || priority == "" {
				http.Error(w, "Missing pid or priority parameter", http.StatusBadRequest)
				return
			}
			result = setProcessPriority(pid, priority)
		case "info":
			if pid == "" {
				http.Error(w, "Missing pid parameter", http.StatusBadRequest)
				return
			}
			result = getProcessInfo(pid)
		case "tree":
			result = getProcessTree()
		default:
			http.Error(w, "Invalid process action. Valid: kill, killforce, nice, info, tree", http.StatusBadRequest)
			return
		}
	case "networkio":
		result = getNetworkIO()
	case "diskio":
		result = getDiskIO()
	case "procs":
		result = getProcessCountByState()
	case "audit":
		action := r.URL.Query().Get("action")
		switch action {
		case "ports":
			result = getOpenPorts()
		case "files":
			result = checkSuspiciousFiles()
		case "permissions":
			result = checkFilePermissions()
		case "users":
			result = checkUnusedUsers()
		case "ssh":
			result = checkSSHSecurity()
		case "summary":
			result = getSecurityAuditSummary()
		default:
			http.Error(w, "Invalid audit action. Valid: ports, files, permissions, users, ssh, summary", http.StatusBadRequest)
			return
		}
	case "cron":
		action := r.URL.Query().Get("action")
		schedule := r.URL.Query().Get("schedule")
		command := r.URL.Query().Get("command")
		line := r.URL.Query().Get("line")

		switch action {
		case "list":
			result = listCronJobsFormatted()
		case "add":
			if schedule == "" || command == "" {
				http.Error(w, "Missing schedule or command parameter", http.StatusBadRequest)
				return
			}
			result = addCronJob(schedule, command)
		case "remove":
			if line == "" {
				http.Error(w, "Missing line parameter", http.StatusBadRequest)
				return
			}
			result = removeCronJob(line)
		case "next":
			result = getCronNextRun()
		default:
			http.Error(w, "Invalid cron action. Valid: list, add, remove, next", http.StatusBadRequest)
			return
		}
	case "maintenance":
		action := r.URL.Query().Get("action")
		if action == "" {
			http.Error(w, "Missing action parameter. Valid: status, enable, disable, check-services, restart-failed, sync-time, clear-cache", http.StatusBadRequest)
			return
		}
		result = getMaintenanceActions(action)
	default:
		http.Error(w, "Unknown command", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatusFor(result))
	if err := json.NewEncoder(w).Encode(map[string]string{"result": result}); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

// httpStatusFor maps error-prefixed result strings to HTTP status codes so
// clients can rely on the response code instead of parsing the body.
func httpStatusFor(result string) int {
	switch {
	case strings.HasPrefix(result, "Usage:"),
		strings.HasPrefix(result, "Invalid"),
		strings.HasPrefix(result, "Unknown"),
		strings.HasPrefix(result, "Unsupported"):
		return http.StatusBadRequest
	case strings.HasPrefix(result, "Error"),
		strings.HasPrefix(result, "Failed"):
		return http.StatusInternalServerError
	default:
		return http.StatusOK
	}
}

// isErrorResult reports whether a result string represents a failure. The CLI
// uses it to pick a non-zero exit code.
func isErrorResult(result string) bool {
	return httpStatusFor(result) != http.StatusOK
}

func printHelp() {
	fmt.Println(`Usage: osctl [command]

Commands:
  ram             Show RAM usage
  disk            Show disk usage
  service         Manage system services
                  Usage: osctl service [start|stop|restart|status|enable|disable] [service_name]
  top             Show top processes by CPU usage
  errors          Show last 10 errors from the journal
  users           Show last 20 logged in users
  uptime          Show system uptime
  osinfo          Show operating system name and kernel version
  shutdown        Shutdown the system
  reboot          Reboot the system
  ip              Show IP addresses of all interfaces
  firewall        Show active firewall rules
  firewallallow   Allow a port: osctl firewallallow <port>[/<proto>] (ufw or firewalld)
  firewalldeny    Deny/remove a port rule: osctl firewalldeny <port>[/<proto>]
  update          Update OS packages
  containers      List all Docker containers
  images          List all Docker images
  dockerlogs      Show last 50 log lines of a container: osctl dockerlogs <container>
  dockerrestart   Restart a container: osctl dockerrestart <container>
  userinfo        Show user identity and password aging: osctl userinfo <username>
  useradd         Create a user with home directory: osctl useradd <username>
  userdel         Delete a user (and home directory): osctl userdel <username>
  cpu             Show CPU usage
  load            Show system load averages
  network         Show network statistics
  connections     List all active network connections
  filesystems     List all mounted filesystems
  dmesg           Show kernel messages
  who             List all currently logged in users
  services        Show status of all running services
  health          Show health check status
  doctor          One-shot diagnostic: health, failed services, disk, errors, updates
  process         Process management (kill, nice, info, tree)
  networkio       Show network I/O statistics
  diskio          Show disk I/O statistics
  procs           Show process count by state
  audit           Security audit (ports, files, permissions, users, ssh, summary)
  cron            Cron job management (list, add, remove, next)
  maintenance     Maintenance mode and system operations (status, enable, disable, check-services, restart-failed, sync-time, clear-cache)
  api             Run as an API server (default port: 12000)
  version         Show osctl version
  --json          Emit any command result as JSON: osctl --json <command>
  --help          Show this help message

Global options:
  --json, -j      Print the command result as {"result": "..."} JSON

Environment:
  OSCTL_CONFIG              Path to a KEY=VALUE config file
  OSCTL_PORT                API port (default 12000)
  OSCTL_USERNAME/PASSWORD   Basic auth credentials
  OSCTL_API_TOKEN           Bearer token alternative for the API
  OSCTL_TLS_CERT/OSCTL_TLS_KEY  Serve the API over HTTPS
  OSCTL_AUDIT_LOG           JSONL request audit log path
  OSCTL_STATE_DIR           Directory for persisted auth-failure state
  OSCTL_WEBHOOK_URL         POST JSON here when health status changes
  OSCTL_HEALTH_INTERVAL     Seconds between health checks (default 300, min 30)`)
}
