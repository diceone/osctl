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
	case "api":
		runAPI()
	default:
		fmt.Println("Unknown command")
		printHelp()
	}
}
