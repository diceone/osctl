package main

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/shirou/gopsutil/process"
)

// killProcess terminates a process by PID
func killProcess(pid string) string {
	// Validate PID
	_, err := strconv.Atoi(pid)
	if err != nil {
		return fmt.Sprintf("Invalid PID: %s", pid)
	}

	cmd := exec.Command("kill", pid)
	err = cmd.Run()
	if err != nil {
		return fmt.Sprintf("Failed to kill process %s. Error: %v", pid, err)
	}
	return fmt.Sprintf("Process %s killed successfully", pid)
}

// killProcessForce forcefully terminates a process by PID
func killProcessForce(pid string) string {
	// Validate PID
	_, err := strconv.Atoi(pid)
	if err != nil {
		return fmt.Sprintf("Invalid PID: %s", pid)
	}

	cmd := exec.Command("kill", "-9", pid)
	err = cmd.Run()
	if err != nil {
		return fmt.Sprintf("Failed to force kill process %s. Error: %v", pid, err)
	}
	return fmt.Sprintf("Process %s force killed successfully", pid)
}

// setProcessPriority sets the nice value (priority) of a process
func setProcessPriority(pid, priority string) string {
	// Validate PID
	_, err := strconv.Atoi(pid)
	if err != nil {
		return fmt.Sprintf("Invalid PID: %s", pid)
	}

	// Validate priority (-20 to 19)
	prio, err := strconv.Atoi(priority)
	if err != nil || prio < -20 || prio > 19 {
		return "Invalid priority. Must be between -20 (highest) and 19 (lowest)"
	}

	cmd := exec.Command("renice", "-n", priority, "-p", pid)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Sprintf("Failed to set priority for process %s. Error: %v", pid, err)
	}
	return string(out)
}

// getProcessInfo gets detailed information about a process
func getProcessInfo(pid string) string {
	// Validate PID
	pidInt, err := strconv.Atoi(pid)
	if err != nil {
		return fmt.Sprintf("Invalid PID: %s", pid)
	}

	proc, err := process.NewProcess(int32(pidInt))
	if err != nil {
		return fmt.Sprintf("Process %s not found. Error: %v", pid, err)
	}

	var output strings.Builder

	name, _ := proc.Name()
	output.WriteString(fmt.Sprintf("Name: %s\n", name))

	cmdline, _ := proc.Cmdline()
	output.WriteString(fmt.Sprintf("Command: %s\n", cmdline))

	status, _ := proc.Status()
	output.WriteString(fmt.Sprintf("Status: %s\n", status))

	cpuPercent, _ := proc.CPUPercent()
	output.WriteString(fmt.Sprintf("CPU%%: %.2f\n", cpuPercent))

	memPercent, _ := proc.MemoryPercent()
	output.WriteString(fmt.Sprintf("Memory%%: %.2f\n", memPercent))

	memInfo, _ := proc.MemoryInfo()
	if memInfo != nil {
		output.WriteString(fmt.Sprintf("RSS: %d MB\n", memInfo.RSS/1024/1024))
		output.WriteString(fmt.Sprintf("VMS: %d MB\n", memInfo.VMS/1024/1024))
	}

	numThreads, _ := proc.NumThreads()
	output.WriteString(fmt.Sprintf("Threads: %d\n", numThreads))

	createTime, _ := proc.CreateTime()
	output.WriteString(fmt.Sprintf("Started: %d\n", createTime))

	username, _ := proc.Username()
	output.WriteString(fmt.Sprintf("User: %s\n", username))

	cwd, _ := proc.Cwd()
	output.WriteString(fmt.Sprintf("CWD: %s\n", cwd))

	return output.String()
}

// getProcessTree shows the process tree
func getProcessTree() string {
	cmd := exec.Command("pstree", "-p")
	out, err := cmd.CombinedOutput()
	if err != nil {
		// Fallback if pstree is not available
		cmd = exec.Command("ps", "axjf")
		out, err = cmd.CombinedOutput()
		if err != nil {
			return fmt.Sprintf("Failed to get process tree. Error: %v", err)
		}
	}
	return string(out)
}
