package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 || os.Args[1] == "--help" {
		printHelp()
		return
	}

	var result string

	switch os.Args[1] {
	case "ram":
		result = getRamUsage()
	case "disk":
		result = getDiskUsage()
	case "service":
		if len(os.Args) < 4 {
			fmt.Println("Usage: osctl service [start|stop|restart|status] [service_name]")
			os.Exit(1)
		}
		result = manageService(os.Args[2], os.Args[3])
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
		if len(os.Args) < 3 {
			fmt.Println("Usage: osctl process [kill|killforce|nice|info|tree] [options]")
			fmt.Println("  kill <pid>           - Terminate process")
			fmt.Println("  killforce <pid>      - Force kill process")
			fmt.Println("  nice <pid> <priority> - Set process priority (-20 to 19)")
			fmt.Println("  info <pid>           - Show process information")
			fmt.Println("  tree                 - Show process tree")
			os.Exit(1)
		}
		action := os.Args[2]
		switch action {
		case "kill":
			if len(os.Args) < 4 {
				fmt.Println("Usage: osctl process kill <pid>")
				os.Exit(1)
			}
			result = killProcess(os.Args[3])
		case "killforce":
			if len(os.Args) < 4 {
				fmt.Println("Usage: osctl process killforce <pid>")
				os.Exit(1)
			}
			result = killProcessForce(os.Args[3])
		case "nice":
			if len(os.Args) < 5 {
				fmt.Println("Usage: osctl process nice <pid> <priority>")
				os.Exit(1)
			}
			result = setProcessPriority(os.Args[3], os.Args[4])
		case "info":
			if len(os.Args) < 4 {
				fmt.Println("Usage: osctl process info <pid>")
				os.Exit(1)
			}
			result = getProcessInfo(os.Args[3])
		case "tree":
			result = getProcessTree()
		default:
			fmt.Println("Unknown process action")
			os.Exit(1)
		}
	case "networkio":
		result = getNetworkIO()
	case "diskio":
		result = getDiskIO()
	case "procs":
		result = getProcessCountByState()
	case "audit":
		if len(os.Args) < 3 {
			fmt.Println("Usage: osctl audit [ports|files|permissions|users|ssh|summary]")
			os.Exit(1)
		}
		action := os.Args[2]
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
			fmt.Println("Unknown audit action")
			os.Exit(1)
		}
	case "cron":
		if len(os.Args) < 3 {
			fmt.Println("Usage: osctl cron [list|add|remove|next]")
			fmt.Println("  list              - List all cron jobs")
			fmt.Println("  add <schedule> <cmd> - Add new cron job")
			fmt.Println("  remove <line>     - Remove cron job by line number")
			fmt.Println("  next              - Show next scheduled runs")
			os.Exit(1)
		}
		action := os.Args[2]
		switch action {
		case "list":
			result = listCronJobsFormatted()
		case "add":
			if len(os.Args) < 5 {
				fmt.Println("Usage: osctl cron add \"schedule\" \"command\"")
				fmt.Println("Example: osctl cron add \"0 2 * * *\" \"/backup.sh\"")
				os.Exit(1)
			}
			result = addCronJob(os.Args[3], os.Args[4])
		case "remove":
			if len(os.Args) < 4 {
				fmt.Println("Usage: osctl cron remove <line_number>")
				os.Exit(1)
			}
			result = removeCronJob(os.Args[3])
		case "next":
			result = getCronNextRun()
		default:
			fmt.Println("Unknown cron action")
			os.Exit(1)
		}
	case "maintenance":
		if len(os.Args) < 3 {
			fmt.Println("Usage: osctl maintenance [status|enable|disable|check-services|restart-failed|sync-time|clear-cache]")
			fmt.Println("  status            - Show maintenance mode status")
			fmt.Println("  enable            - Enable maintenance mode")
			fmt.Println("  disable           - Disable maintenance mode")
			fmt.Println("  check-services    - Check critical services status")
			fmt.Println("  restart-failed    - Restart all failed services")
			fmt.Println("  sync-time         - Synchronize system time")
			fmt.Println("  clear-cache       - Clear system caches")
			os.Exit(1)
		}
		action := os.Args[2]
		result = getMaintenanceActions(action)
	case "api":
		runAPI()
		return
	default:
		fmt.Println("Unknown command")
		printHelp()
		os.Exit(1)
	}

	fmt.Println(result)
	if isErrorResult(result) {
		os.Exit(1)
	}
}
