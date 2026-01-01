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

	switch os.Args[1] {
	case "ram":
		fmt.Println(getRamUsage())
	case "disk":
		fmt.Println(getDiskUsage())
	case "service":
		if len(os.Args) < 4 {
			fmt.Println("Usage: osctl service [start|stop|restart|status] [service_name]")
			return
		}
		action := os.Args[2]
		service := os.Args[3]
		fmt.Println(manageService(action, service))
	case "top":
		fmt.Println(getTopProcesses())
	case "errors":
		fmt.Println(getLastJournalErrors())
	case "users":
		fmt.Println(getLastLoggedUsers())
	case "uptime":
		fmt.Println(getUptime())
	case "osinfo":
		fmt.Println(getOSInfo())
	case "shutdown":
		fmt.Println(shutdownSystem())
	case "reboot":
		fmt.Println(rebootSystem())
	case "ip":
		fmt.Println(getIPAddresses())
	case "firewall":
		fmt.Println(getFirewalldRules())
	case "update":
		fmt.Println(updatePackages())
	case "containers":
		fmt.Println(listDockerContainers())
	case "images":
		fmt.Println(listDockerImages())
	case "cpu":
		fmt.Println(getCpuUsage())
	case "load":
		fmt.Println(getLoadAverage())
	case "network":
		fmt.Println(getNetworkStats())
	case "connections":
		fmt.Println(getActiveConnections())
	case "filesystems":
		fmt.Println(getMountedFilesystems())
	case "dmesg":
		fmt.Println(getKernelMessages())
	case "who":
		fmt.Println(getLoggedinUsers())
	case "services":
		fmt.Println(getServiceStatuses())
	case "health":
		fmt.Println(getHealthCheck())
	case "process":
		if len(os.Args) < 3 {
			fmt.Println("Usage: osctl process [kill|killforce|nice|info|tree] [options]")
			fmt.Println("  kill <pid>           - Terminate process")
			fmt.Println("  killforce <pid>      - Force kill process")
			fmt.Println("  nice <pid> <priority> - Set process priority (-20 to 19)")
			fmt.Println("  info <pid>           - Show process information")
			fmt.Println("  tree                 - Show process tree")
			return
		}
		action := os.Args[2]
		switch action {
		case "kill":
			if len(os.Args) < 4 {
				fmt.Println("Usage: osctl process kill <pid>")
				return
			}
			fmt.Println(killProcess(os.Args[3]))
		case "killforce":
			if len(os.Args) < 4 {
				fmt.Println("Usage: osctl process killforce <pid>")
				return
			}
			fmt.Println(killProcessForce(os.Args[3]))
		case "nice":
			if len(os.Args) < 5 {
				fmt.Println("Usage: osctl process nice <pid> <priority>")
				return
			}
			fmt.Println(setProcessPriority(os.Args[3], os.Args[4]))
		case "info":
			if len(os.Args) < 4 {
				fmt.Println("Usage: osctl process info <pid>")
				return
			}
			fmt.Println(getProcessInfo(os.Args[3]))
		case "tree":
			fmt.Println(getProcessTree())
		default:
			fmt.Println("Unknown process action")
		}
	case "networkio":
		fmt.Println(getNetworkIO())
	case "diskio":
		fmt.Println(getDiskIO())
	case "procs":
		fmt.Println(getProcessCountByState())
	case "audit":
		if len(os.Args) < 3 {
			fmt.Println("Usage: osctl audit [ports|files|permissions|users|ssh|summary]")
			return
		}
		action := os.Args[2]
		switch action {
		case "ports":
			fmt.Println(getOpenPorts())
		case "files":
			fmt.Println(checkSuspiciousFiles())
		case "permissions":
			fmt.Println(checkFilePermissions())
		case "users":
			fmt.Println(checkUnusedUsers())
		case "ssh":
			fmt.Println(checkSSHSecurity())
		case "summary":
			fmt.Println(getSecurityAuditSummary())
		default:
			fmt.Println("Unknown audit action")
		}
	case "cron":
		if len(os.Args) < 3 {
			fmt.Println("Usage: osctl cron [list|add|remove|next]")
			fmt.Println("  list              - List all cron jobs")
			fmt.Println("  add <schedule> <cmd> - Add new cron job")
			fmt.Println("  remove <line>     - Remove cron job by line number")
			fmt.Println("  next              - Show next scheduled runs")
			return
		}
		action := os.Args[2]
		switch action {
		case "list":
			fmt.Println(listCronJobsFormatted())
		case "add":
			if len(os.Args) < 5 {
				fmt.Println("Usage: osctl cron add \"schedule\" \"command\"")
				fmt.Println("Example: osctl cron add \"0 2 * * *\" \"/backup.sh\"")
				return
			}
			fmt.Println(addCronJob(os.Args[3], os.Args[4]))
		case "remove":
			if len(os.Args) < 4 {
				fmt.Println("Usage: osctl cron remove <line_number>")
				return
			}
			fmt.Println(removeCronJob(os.Args[3]))
		case "next":
			fmt.Println(getCronNextRun())
		default:
			fmt.Println("Unknown cron action")
		}
	case "api":
		runAPI()
	default:
		fmt.Println("Unknown command")
		printHelp()
	}
}
