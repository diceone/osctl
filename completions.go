package main

import (
	"fmt"
	"strings"
)

// commandDocs lists every top-level CLI command with a short description.
// It is the single source of truth for the generated completion scripts.
var commandDocs = [][2]string{
	{"ram", "Show RAM usage"},
	{"disk", "Show disk usage"},
	{"service", "Manage system services"},
	{"top", "Show top processes by CPU usage"},
	{"errors", "Show last 10 errors from the journal"},
	{"users", "Show last 20 logged in users"},
	{"uptime", "Show system uptime"},
	{"osinfo", "Show operating system name and kernel version"},
	{"shutdown", "Shutdown the system"},
	{"reboot", "Reboot the system"},
	{"ip", "Show IP addresses of all interfaces"},
	{"firewall", "Show active firewall rules"},
	{"firewallallow", "Allow a port (ufw or firewalld)"},
	{"firewalldeny", "Deny/remove a port rule"},
	{"update", "Update OS packages"},
	{"updates", "List available package updates without installing"},
	{"containers", "List all Docker containers"},
	{"images", "List all Docker images"},
	{"dockerlogs", "Show last 50 log lines of a container"},
	{"dockerstats", "Show per-container CPU and memory stats"},
	{"dockerrestart", "Restart a container"},
	{"userinfo", "Show user identity and password aging"},
	{"useradd", "Create a user with home directory"},
	{"userdel", "Delete a user (and home directory)"},
	{"cpu", "Show CPU usage"},
	{"load", "Show system load averages"},
	{"network", "Show network statistics"},
	{"networkio", "Show network I/O statistics"},
	{"connections", "List all active network connections"},
	{"filesystems", "List all mounted filesystems"},
	{"diskio", "Show disk I/O statistics"},
	{"dmesg", "Show kernel messages"},
	{"who", "List all currently logged in users"},
	{"services", "Show status of all running services"},
	{"health", "Show health check status"},
	{"doctor", "One-shot diagnostic: health, failed services, disk, errors, updates"},
	{"process", "Process management"},
	{"procs", "Show process count by state"},
	{"audit", "Security audit"},
	{"cron", "Cron job management"},
	{"maintenance", "Maintenance mode and system operations"},
	{"logs", "Show recent journal entries for a systemd unit"},
	{"boot", "Show boot time and slowest units"},
	{"sensors", "Show temperatures and fan speeds"},
	{"certs", "Check TLS certificate expiry"},
	{"watch", "Re-run a command on an interval and alert on changes"},
	{"completion", "Generate shell completion script"},
	{"api", "Run as an API server"},
	{"version", "Show osctl version"},
	{"help", "Show help message"},
}

// completionSubcommands maps parent commands to their subcommand words.
var completionSubcommands = map[string][]string{
	"audit":       {"ports", "files", "permissions", "users", "ssh", "sysctl", "mac", "summary"},
	"cron":        {"list", "add", "remove", "next"},
	"maintenance": {"status", "enable", "disable", "check-services", "restart-failed", "sync-time", "clear-cache"},
	"process":     {"kill", "killforce", "nice", "info", "tree"},
	"service":     {"start", "stop", "restart", "status", "enable", "disable"},
	"completion":  {"bash", "zsh", "fish"},
}

// commandNames returns the bare command words from commandDocs.
func commandNames() []string {
	names := make([]string, 0, len(commandDocs))
	for _, c := range commandDocs {
		names = append(names, c[0])
	}
	return names
}

// completionCommand returns the completion script for the requested shell.
func completionCommand(shell string) string {
	switch shell {
	case "":
		return "Usage: osctl completion [bash|zsh|fish]"
	case "bash":
		return bashCompletion()
	case "zsh":
		return zshCompletion()
	case "fish":
		return fishCompletion()
	default:
		return "Unsupported shell: " + shell + " (valid: bash, zsh, fish)"
	}
}

func bashCompletion() string {
	var subcases strings.Builder
	for parent, subs := range completionSubcommands {
		subcases.WriteString(fmt.Sprintf("    %s) COMPREPLY=($(compgen -W \"%s\" -- \"${cur}\")); return ;;\n",
			parent, strings.Join(subs, " ")))
	}

	return fmt.Sprintf(`# bash completion for osctl
# Install: osctl completion bash > /etc/bash_completion.d/osctl
_osctl() {
  local cur="${COMP_WORDS[COMP_CWORD]}"
  local prev="${COMP_WORDS[COMP_CWORD-1]}"
  local commands="%[1]s"
  case "${prev}" in
%[2]s    *)
      COMPREPLY=($(compgen -W "${commands}" -- "${cur}")) ;;
  esac
}
complete -F _osctl osctl
`, strings.Join(commandNames(), " "), subcases.String())
}

func zshCompletion() string {
	var lines strings.Builder
	for _, c := range commandDocs {
		lines.WriteString(fmt.Sprintf("    '%s:%s'\n", c[0], c[1]))
	}

	return fmt.Sprintf(`#compdef osctl
# zsh completion for osctl
# Install: osctl completion zsh > "${fpath[1]}/_osctl"
_osctl() {
  local -a commands
  commands=(
%s  )
  _describe 'osctl command' commands
}
_osctl "$@"
`, lines.String())
}

func fishCompletion() string {
	var lines strings.Builder
	for _, c := range commandDocs {
		lines.WriteString(fmt.Sprintf("complete -c osctl -f -n '__fish_use_subcommand' -a %s -d '%s'\n", c[0], c[1]))
	}
	for parent, subs := range completionSubcommands {
		for _, sub := range subs {
			lines.WriteString(fmt.Sprintf("complete -c osctl -f -n '__fish_seen_subcommand_from %s' -a %s\n", parent, sub))
		}
	}
	return fmt.Sprintf("# fish completion for osctl\n# Install: osctl completion fish > ~/.config/fish/completions/osctl.fish\n%s", lines.String())
}
