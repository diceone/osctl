package main

import (
	"encoding/json"
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
               Usage: osctl service [start|stop|restart|status] [service_name]
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
  api          Run as an API server on port 12000
  --help       Show this help message`)
}
