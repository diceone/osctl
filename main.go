package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	applyConfigFile()

	args := os.Args[1:]

	// --json emits the result as {"result": "..."} for scripts.
	jsonOut := false
	if len(args) > 0 && (args[0] == "--json" || args[0] == "-j") {
		jsonOut = true
		args = args[1:]
	}

	if len(args) < 1 || args[0] == "--help" || args[0] == "-h" {
		printHelp()
		return
	}

	if args[0] == "api" {
		runAPI()
		return
	}

	result := executeCommand(args)

	if jsonOut {
		data, err := json.Marshal(map[string]string{"result": result})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to encode JSON output: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(data))
	} else {
		fmt.Println(result)
	}

	if isErrorResult(result) {
		os.Exit(1)
	}
}

func executeCommand(args []string) string {
	switch args[0] {
	case "version":
		return getVersion()
	case "ram":
		return getRamUsage()
	case "disk":
		return getDiskUsage()
	case "doctor":
		return getDoctorReport()
	case "service":
		if len(args) < 3 {
			fmt.Println("Usage: osctl service [start|stop|restart|status] [service_name]")
			os.Exit(1)
		}
		return manageService(args[1], args[2])
	case "top":
		return getTopProcesses()
	case "errors":
		return getLastJournalErrors()
	case "users":
		return getLastLoggedUsers()
	case "uptime":
		return getUptime()
	case "osinfo":
		return getOSInfo()
	case "shutdown":
		return shutdownSystem()
	case "reboot":
		return rebootSystem()
	case "ip":
		return getIPAddresses()
	case "firewall":
		return getFirewalldRules()
	case "firewallallow":
		if len(args) < 2 {
			fmt.Println("Usage: osctl firewallallow <port>[/<proto>]")
			os.Exit(1)
		}
		return manageFirewallRule("allow", args[1])
	case "firewalldeny":
		if len(args) < 2 {
			fmt.Println("Usage: osctl firewalldeny <port>[/<proto>]")
			os.Exit(1)
		}
		return manageFirewallRule("deny", args[1])
	case "update":
		return updatePackages()
	case "containers":
		return listDockerContainers()
	case "images":
		return listDockerImages()
	case "dockerlogs":
		if len(args) < 2 {
			fmt.Println("Usage: osctl dockerlogs <container>")
			os.Exit(1)
		}
		return dockerContainerLogs(args[1])
	case "dockerrestart":
		if len(args) < 2 {
			fmt.Println("Usage: osctl dockerrestart <container>")
			os.Exit(1)
		}
		return dockerRestartContainer(args[1])
	case "userinfo":
		if len(args) < 2 {
			fmt.Println("Usage: osctl userinfo <username>")
			os.Exit(1)
		}
		return getUserInfo(args[1])
	case "useradd":
		if len(args) < 2 {
			fmt.Println("Usage: osctl useradd <username>")
			os.Exit(1)
		}
		return addUser(args[1])
	case "userdel":
		if len(args) < 2 {
			fmt.Println("Usage: osctl userdel <username>")
			os.Exit(1)
		}
		return deleteUser(args[1])
	case "cpu":
		return getCpuUsage()
	case "load":
		return getLoadAverage()
	case "network":
		return getNetworkStats()
	case "connections":
		return getActiveConnections()
	case "filesystems":
		return getMountedFilesystems()
	case "dmesg":
		return getKernelMessages()
	case "who":
		return getLoggedinUsers()
	case "services":
		return getServiceStatuses()
	case "health":
		return getHealthCheck()
	case "process":
		if len(args) < 2 {
			fmt.Println("Usage: osctl process [kill|killforce|nice|info|tree] [options]")
			fmt.Println("  kill <pid>           - Terminate process")
			fmt.Println("  killforce <pid>      - Force kill process")
			fmt.Println("  nice <pid> <priority> - Set process priority (-20 to 19)")
			fmt.Println("  info <pid>           - Show process information")
			fmt.Println("  tree                 - Show process tree")
			os.Exit(1)
		}
		switch args[1] {
		case "kill":
			if len(args) < 3 {
				fmt.Println("Usage: osctl process kill <pid>")
				os.Exit(1)
			}
			return killProcess(args[2])
		case "killforce":
			if len(args) < 3 {
				fmt.Println("Usage: osctl process killforce <pid>")
				os.Exit(1)
			}
			return killProcessForce(args[2])
		case "nice":
			if len(args) < 4 {
				fmt.Println("Usage: osctl process nice <pid> <priority>")
				os.Exit(1)
			}
			return setProcessPriority(args[2], args[3])
		case "info":
			if len(args) < 3 {
				fmt.Println("Usage: osctl process info <pid>")
				os.Exit(1)
			}
			return getProcessInfo(args[2])
		case "tree":
			return getProcessTree()
		default:
			fmt.Println("Unknown process action")
			os.Exit(1)
		}
	case "networkio":
		return getNetworkIO()
	case "diskio":
		return getDiskIO()
	case "procs":
		return getProcessCountByState()
	case "audit":
		action := ""
		if len(args) >= 2 {
			action = args[1]
		}
		switch action {
		case "ports":
			return getOpenPorts()
		case "files":
			return checkSuspiciousFiles()
		case "permissions":
			return checkFilePermissions()
		case "users":
			return checkUnusedUsers()
		case "ssh":
			return checkSSHSecurity()
		case "summary":
			return getSecurityAuditSummary()
		default:
			fmt.Println("Usage: osctl audit [ports|files|permissions|users|ssh|summary]")
			os.Exit(1)
		}
	case "cron":
		if len(args) < 2 {
			fmt.Println("Usage: osctl cron [list|add|remove|next]")
			fmt.Println("  list              - List all cron jobs")
			fmt.Println("  add <schedule> <cmd> - Add new cron job")
			fmt.Println("  remove <line>     - Remove cron job by line number")
			fmt.Println("  next              - Show next scheduled runs")
			os.Exit(1)
		}
		switch args[1] {
		case "list":
			return listCronJobsFormatted()
		case "add":
			if len(args) < 4 {
				fmt.Println("Usage: osctl cron add \"schedule\" \"command\"")
				fmt.Println("Example: osctl cron add \"0 2 * * *\" \"/backup.sh\"")
				os.Exit(1)
			}
			return addCronJob(args[2], args[3])
		case "remove":
			if len(args) < 3 {
				fmt.Println("Usage: osctl cron remove <line_number>")
				os.Exit(1)
			}
			return removeCronJob(args[2])
		case "next":
			return getCronNextRun()
		default:
			fmt.Println("Unknown cron action")
			os.Exit(1)
		}
	case "maintenance":
		if len(args) < 2 {
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
		return getMaintenanceActions(args[1])
	default:
		fmt.Println("Unknown command")
		printHelp()
		os.Exit(1)
	}

	return "" // unreachable: every branch above returns or exits
}
