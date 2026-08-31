package main

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// validPortSpec accepts "8080" or "8080/tcp", "8080/udp", "8080/sctp".
var validPortSpec = regexp.MustCompile(`^([0-9]{1,5})(/(tcp|udp))?$`)

func validatePortSpec(spec string) (string, error) {
	m := validPortSpec.FindStringSubmatch(spec)
	if m == nil {
		return "", fmt.Errorf("invalid port specification %q (expected <port> or <port>/tcp|udp)", spec)
	}
	port, err := strconv.Atoi(m[1])
	if err != nil || port < 1 || port > 65535 {
		return "", fmt.Errorf("port %s out of range (1-65535)", m[1])
	}
	return spec, nil
}

// firewallBackend returns the available firewall tool, preferring ufw.
func firewallBackend() (string, error) {
	if _, err := exec.LookPath("ufw"); err == nil {
		return "ufw", nil
	}
	if _, err := exec.LookPath("firewall-cmd"); err == nil {
		return "firewalld", nil
	}
	return "", fmt.Errorf("no supported firewall tool found (need ufw or firewall-cmd)")
}

// manageFirewallRule allows or removes a port rule. action must be "allow" or
// "deny" (deny == remove the rule on firewalld).
func manageFirewallRule(action, portSpec string) string {
	if action != "allow" && action != "deny" {
		return fmt.Sprintf("Invalid action %q. Valid: allow, deny", action)
	}
	if _, err := validatePortSpec(portSpec); err != nil {
		return "Invalid port specification: " + err.Error()
	}

	backend, err := firewallBackend()
	if err != nil {
		return "Failed to manage firewall: " + err.Error()
	}

	var cmd *exec.Cmd
	switch backend {
	case "ufw":
		if action == "allow" {
			cmd = exec.Command("ufw", "allow", portSpec)
		} else {
			cmd = exec.Command("ufw", "deny", portSpec)
		}
	case "firewalld":
		if action == "allow" {
			cmd = exec.Command("firewall-cmd", "--permanent", "--add-port="+portSpec)
		} else {
			cmd = exec.Command("firewall-cmd", "--permanent", "--remove-port="+portSpec)
		}
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Sprintf("Failed to %s %s on %s. Error: %v\n%s", action, portSpec, backend, err, string(out))
	}

	// firewalld only applies --permanent changes after a reload.
	if backend == "firewalld" {
		if reloadOut, err := exec.Command("firewall-cmd", "--reload").CombinedOutput(); err != nil {
			return fmt.Sprintf("Rule updated persistently but reload failed:\n%s%s", string(reloadOut), ruleSuccessHint(action, portSpec))
		}
	}

	return ruleSuccessHint(action, portSpec) + strings.TrimSpace(string(out)) + "\n"
}

func ruleSuccessHint(action, portSpec string) string {
	if action == "allow" {
		return fmt.Sprintf("Firewall rule applied: %s allowed (%s).\n", portSpec, firewallBackendName())
	}
	return fmt.Sprintf("Firewall rule applied: %s denied/removed (%s).\n", portSpec, firewallBackendName())
}

func firewallBackendName() string {
	name, _ := firewallBackend()
	return name
}
