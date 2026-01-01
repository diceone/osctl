package main

import (
	"encoding/json"
	"fmt" // Added import
	"net/http"
)

func handleRequest(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[1:]

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
	case "shutdown":
		result = shutdownSystem()
	case "reboot":
		result = rebootSystem()
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
	default:
		http.Error(w, "Unknown command", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"result": result})
}

func printHelp() {
	fmt.Println(`Usage: osctl [command]

Commands:
  ram          Show RAM usage
  disk         Show disk usage
  service      Manage system services
               Usage: osctl service [start|stop|restart|status|enable|disable] [service_name]
  top          Show top processes by CPU usage
  errors       Show last 10 errors from the journal
  users        Show last 20 logged in users
  uptime       Show system uptime
  osinfo       Show operating system name and kernel version
  shutdown     Shutdown the system
  reboot       Reboot the system
  ip           Show IP addresses of all interfaces
  firewall     Show active firewalld rules
  update       Update OS packages
  containers   List all Docker containers
  images       List all Docker images
  cpu          Show CPU usage
  load         Show system load averages
  network      Show network statistics
  connections  List all active network connections
  filesystems  List all mounted filesystems
  dmesg        Show kernel messages
  who          List all currently logged in users
  services     Show status of all running services
  health       Show health check status
  process      Process management (kill, nice, info, tree)
  networkio    Show network I/O statistics
  diskio       Show disk I/O statistics
  procs        Show process count by state
  audit        Security audit (ports, files, permissions, users, ssh, summary)
  cron         Cron job management (list, add, remove, next)
  api          Run as an API server (default port: 12000)
  --help       Show this help message`)
}
