package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/vishvananda/netlink"
)

func manageService(action, service string) string {
	// Validate action
	validActions := map[string]bool{
		"start":   true,
		"stop":    true,
		"restart": true,
		"status":  true,
		"enable":  true,
		"disable": true,
	}
	if !validActions[action] {
		return fmt.Sprintf("Invalid action '%s'. Valid actions: start, stop, restart, status, enable, disable", action)
	}

	// Basic validation for service name (prevent command injection)
	if strings.ContainsAny(service, ";|&$`\n\r") {
		return "Invalid service name: contains forbidden characters"
	}

	cmd := exec.Command("systemctl", action, service)
	err := cmd.Run()
	if err != nil {
		return fmt.Sprintf("Failed to %s service %s. Error: %v", action, service, err)
	}
	return fmt.Sprintf("Service %s %sed successfully.", service, action)
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

func updatePackages() string {
	var cmd *exec.Cmd

	// Check for /etc/os-release first (modern standard)
	if data, err := os.ReadFile("/etc/os-release"); err == nil {
		osRelease := string(data)
		if strings.Contains(strings.ToLower(osRelease), "ubuntu") || strings.Contains(strings.ToLower(osRelease), "debian") {
			cmd = exec.Command("apt-get", "update", "-y")
			cmd.Run()
			cmd = exec.Command("apt-get", "upgrade", "-y")
		} else if strings.Contains(strings.ToLower(osRelease), "rhel") || strings.Contains(strings.ToLower(osRelease), "centos") || strings.Contains(strings.ToLower(osRelease), "fedora") {
			cmd = exec.Command("yum", "update", "-y")
		} else if strings.Contains(strings.ToLower(osRelease), "suse") || strings.Contains(strings.ToLower(osRelease), "opensuse") {
			cmd = exec.Command("zypper", "refresh")
			cmd.Run()
			cmd = exec.Command("zypper", "update", "-y")
		} else {
			return "Unsupported OS for package update"
		}
	} else {
		// Fallback to old detection methods
		if _, err := os.Stat("/etc/redhat-release"); err == nil {
			cmd = exec.Command("yum", "update", "-y")
		} else if _, err := os.Stat("/etc/debian_version"); err == nil {
			cmd = exec.Command("apt-get", "update", "-y")
			cmd.Run()
			cmd = exec.Command("apt-get", "upgrade", "-y")
		} else {
			return "Unsupported OS for package update"
		}
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

func getIPAddresses() string {
	links, err := netlink.LinkList()
	if err != nil {
		return fmt.Sprintf("Error getting network interfaces: %v", err)
	}

	var output strings.Builder
	for _, link := range links {
		addrs, err := netlink.AddrList(link, syscall.AF_UNSPEC)
		if err != nil {
			output.WriteString(fmt.Sprintf("Error getting addresses for interface %v: %v\n", link.Attrs().Name, err))
			continue
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

func getServiceStatuses() string {
	cmd := exec.Command("systemctl", "list-units", "--type=service", "--state=running")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Sprintf("Failed to get service statuses. Error: %v", err)
	}
	return string(out)
}
