package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
	"github.com/shirou/gopsutil/process"
	"github.com/vishvananda/netlink"
)

const (
	username = "admin"
	password = "password"
)

func basicAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		payload, err := base64.StdEncoding.DecodeString(auth[len("Basic "):])
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		pair := strings.SplitN(string(payload), ":", 2)
		if len(pair) != 2 || pair[0] != username || pair[1] != password {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func getRamUsage() string {
	v, err := mem.VirtualMemory()
	if err != nil {
		log.Fatalf("Error getting RAM usage: %v", err)
	}

	return fmt.Sprintf("Total: %v MB, Used: %v MB, Free: %v MB",
		v.Total/1024/1024, v.Used/1024/1024, v.Available/1024/1024)
}

func getDiskUsage() string {
	d, err := disk.Usage("/")
	if err != nil {
		log.Fatalf("Error getting disk usage: %v", err)
	}

	return fmt.Sprintf("Total: %v GB, Used: %v GB, Free: %v GB",
		d.Total/1024/1024/1024, d.Used/1024/1024/1024, d.Free/1024/1024/1024)
}

func manageService(action, service string) string {
	cmd := exec.Command("systemctl", action, service)
	err := cmd.Run()
	if err != nil {
		return fmt.Sprintf("Failed to %s service %s. Error: %v", action, service, err)
	}
	return fmt.Sprintf("Service %s %sed successfully.", service, action)
}

func getTopProcesses() string {
	procs, err := process.Processes()
	if err != nil {
		log.Fatalf("Error getting processes: %v", err)
	}

	type procInfo struct {
		PID   int32
		Name  string
		CPU   float64
		Mem   float32
	}

	var procList []procInfo
	for _, p := range procs {
		name, err := p.Name()
		if err != nil {
			continue
		}
		cpu, err := p.CPUPercent()
		if err != nil {
			continue
		}
		mem, err := p.MemoryPercent()
		if err != nil {
			continue
		}
		procList = append(procList, procInfo{PID: p.Pid, Name: name, CPU: cpu, Mem: mem})
	}

	sort.Slice(procList, func(i, j int) bool {
		return procList[i].CPU > procList[j].CPU
	})

	output := "PID\tName\t\tCPU%\tMemory%\n"
	for i, p := range procList {
		if i >= 10 {
			break
		}
		output += fmt.Sprintf("%d\t%s\t\t%.2f\t%.2f\n", p.PID, p.Name, p.CPU, p.Mem)
	}
	return output
}

func getLastJournalErrors() string {
	cmd := exec.Command("journalctl", "-p", "err", "-n", "10", "--no-pager")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Sprintf("Failed to get journal errors. Error: %v", err)
	}
	return string(out)
}

func getLastLoggedUsers() string {
	cmd := exec.Command("last", "-n", "20")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Sprintf("Failed to get last logged users. Error: %v", err)
	}
	return string(out)
}

func getUptime() string {
	uptime, err := host.Uptime()
	if err != nil {
		log.Fatalf("Error getting uptime: %v", err)
	}
	return fmt.Sprintf("Uptime: %v", time.Duration(uptime)*time.Second)
}

func getOSInfo() string {
	info, err := host.Info()
	if err != nil {
		log.Fatalf("Error getting OS info: %v", err)
	}
	return fmt.Sprintf("OS: %s %s\nKernel: %s", info.Platform, info.PlatformVersion, info.KernelVersion)
}

func shutdownSystem() string {
	cmd := exec.Command("shutdown", "now")
	err := cmd.Run()
	if err != nil {
		return fmt.Sprintf("Failed to shutdown the system. Error: %v", err)
	}
	return "System is shutting down..."
}

func rebootSystem() string {
	cmd := exec.Command("reboot")
	err := cmd.Run()
	if err != nil {
		return fmt.Sprintf("Failed to reboot the system. Error: %v", err)
	}
	return "System is rebooting..."
}

func getIPAddresses() string {
	links, err := netlink.LinkList()
	if err != nil {
		log.Fatalf("Error getting network interfaces: %v", err)
	}

	var output strings.Builder
	for _, link := range links {
		addrs, err := netlink.AddrList(link, syscall.AF_UNSPEC)
		if err != nil {
			log.Fatalf("Error getting addresses for interface %v: %v", link.Attrs().Name, err)
		}
		if len(addrs) > 0 {
			output.WriteString(fmt.Sprintf("Interface %s:\n", link.Attrs().Name))
			for _, addr := range addrs {
				output.WriteString(fmt.Sprintf("  %s\n", addr.IP.String()))
			}
		}
	}

	return output.String()
}

func getFirewalldRules() string {
	cmd := exec.Command("firewall-cmd", "--list-all")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Sprintf("Failed to get firewalld rules. Error: %v", err)
	}
	return string(out)
}

func updatePackages() string {
	var cmd *exec.Cmd
	if _, err := os.Stat("/etc/redhat-release"); err == nil {
		cmd = exec.Command("yum", "update", "-y")
	} else if _, err := os.Stat("/etc/lsb-release"); err == nil {
		cmd = exec.Command("apt-get", "update", "-y")
		cmd.Run()
		cmd = exec.Command("apt-get", "upgrade", "-y")
	} else if _, err := os.Stat("/etc/SuSE-release"); err == nil {
		cmd = exec.Command("zypper", "refresh")
		cmd.Run()
		cmd = exec.Command("zypper", "update", "-y")
	} else {
		return "Unsupported OS for package update"
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Sprintf("Failed to update packages. Error: %v", err)
	}
	return string(out)
}

func listDockerContainers() string {
	cmd := exec.Command("docker", "ps", "-a")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Sprintf("Failed to list Docker containers. Error: %v", err)
	}
	return string(out)
}

func listDockerImages() string {
	cmd := exec.Command("docker", "images")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Sprintf("Failed to list Docker images. Error: %v", err)
	}
	return string(out)
}

func getCpuUsage() string {
	cpuPercentages, err := cpu.Percent(0, false)
	if err != nil {
		log.Fatalf("Error getting CPU usage: %v", err)
	}
	return fmt.Sprintf("CPU Usage: %.2f%%", cpuPercentages[0])
}

func getLoadAverage() string {
	avg, err := load.Avg()
	if err != nil {
		log.Fatalf("Error getting load average: %v", err)
	}
	return fmt.Sprintf("Load Average: 1 min: %.2f, 5 min: %.2f, 15 min: %.2f", avg.Load1, avg.Load5, avg.Load15)
}

func getNetworkStats() string {
	stats, err := net.IOCounters(true)
	if err != nil {
		log.Fatalf("Error getting network stats: %v", err)
	}

	var output strings.Builder
	for _, stat := range stats {
		output.WriteString(fmt.Sprintf("Interface: %s\n", stat.Name))
		output.WriteString(fmt.Sprintf("  Bytes Sent: %v\n", stat.BytesSent))
		output.WriteString(fmt.Sprintf("  Bytes Received: %v\n", stat.BytesRecv))
		output.WriteString(fmt.Sprintf("  Packets Sent: %v\n", stat.PacketsSent))
		output.WriteString(fmt.Sprintf("  Packets Received: %v\n", stat.PacketsRecv))
		output.WriteString(fmt.Sprintf("  Errors In: %v\n", stat.Errin))
		output.WriteString(fmt.Sprintf("  Errors Out: %v\n", stat.Errout))
		output.WriteString(fmt.Sprintf("  Drops In: %v\n", stat.Dropin))
		output.WriteString(fmt.Sprintf("  Drops Out: %v\n", stat.Dropout))
	}

	return output.String()
}

func getActiveConnections() string {
	connections, err := net.Connections("all")
	if err != nil {
		log.Fatalf("Error getting active connections: %v", err)
	}

	var output strings.Builder
	for _, conn := range connections {
		output.WriteString(fmt.Sprintf("Type: %s, Local Address: %s:%d, Remote Address: %s:%d, Status: %s\n",
			conn.Type, conn.Laddr.IP, conn.Laddr.Port, conn.Raddr.IP, conn.Raddr.Port, conn.Status))
	}

	return output.String()
}

func getMountedFilesystems() string {
	partitions, err := disk.Partitions(false)
	if err != nil {
		log.Fatalf("Error getting mounted filesystems: %v", err)
	}

	var output strings.Builder
	for _, partition := range partitions {
		usage, err := disk.Usage(partition.Mountpoint)
		if err != nil {
			log.Fatalf("Error getting usage for partition %s: %v", partition.Mountpoint, err)
		}
		output.WriteString(fmt.Sprintf("Mountpoint: %s, Total: %v GB, Used: %v GB, Free: %v GB, Usage: %.2f%%\n",
			partition.Mountpoint, usage.Total/1024/1024/1024, usage.Used/1024/1024/1024, usage.Free/1024/1024/1024, usage.UsedPercent))
	}

	return output.String()
}

func getKernelMessages() string {
	cmd := exec.Command("dmesg", "-T")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Sprintf("Failed to get kernel messages. Error: %v", err)
	}
	return string(out)
}

func getLoggedinUsers() string {
	cmd := exec.Command("who")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Sprintf("Failed to get logged-in users. Error: %v", err)
	}
	return string(out)
}

func getServiceStatuses() string {
	cmd := exec.Command("systemctl", "list-units", "--type=service", "--state=running")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Sprintf("Failed to get service statuses. Error: %v", err)
	}
	return string(out)
}

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

func runAPI() {
	http.HandleFunc("/", handleRequest)
	fmt.Println("Server is listening on port 12000...")
	log.Fatal(http.ListenAndServe(":12000", nil))
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
