package main

import (
	"fmt"
	"log"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
	"github.com/shirou/gopsutil/process"
)

func getRamUsage() string {
	v, err := mem.VirtualMemory()
	if err != nil {
		log.Fatalf("Error getting RAM usage: %v", err)
	}

	ramUsage.WithLabelValues("total").Set(float64(v.Total))
	ramUsage.WithLabelValues("used").Set(float64(v.Used))
	ramUsage.WithLabelValues("free").Set(float64(v.Available))

	return fmt.Sprintf("Total: %v MB, Used: %v MB, Free: %v MB",
		v.Total/1024/1024, v.Used/1024/1024, v.Available/1024/1024)
}

func getDiskUsage() string {
	d, err := disk.Usage("/")
	if err != nil {
		log.Fatalf("Error getting disk usage: %v", err)
	}

	diskUsage.WithLabelValues("total").Set(float64(d.Total))
	diskUsage.WithLabelValues("used").Set(float64(d.Used))
	diskUsage.WithLabelValues("free").Set(float64(d.Free))

	return fmt.Sprintf("Total: %v GB, Used: %v GB, Free: %v GB",
		d.Total/1024/1024/1024, d.Used/1024/1024/1024, d.Free/1024/1024/1024)
}

func getCpuUsage() string {
	cpuPercentages, err := cpu.Percent(0, false)
	if err != nil {
		log.Fatalf("Error getting CPU usage: %v", err)
	}
	cpuUsage.Set(cpuPercentages[0])
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
		output.WriteString(fmt.Sprintf("Type: %d, Local Address: %s:%d, Remote Address: %s:%d, Status: %s\n",
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
