package main

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

type portEntry struct {
	address string
	port    string
	process string
	pid     string
}

var ssProcessRe = regexp.MustCompile(`users:\(\("([^"]+)",pid=([0-9]+)`)

// getListeningPorts lists TCP ports with the owning process.
func getListeningPorts() string {
	out, err := exec.Command("ss", "-tlnp").CombinedOutput()
	if err != nil {
		// Fall back to netstat when iproute2 is not installed.
		if netOut, netErr := exec.Command("netstat", "-tlnp").CombinedOutput(); netErr == nil {
			return renderPorts(parseNetstatPorts(string(netOut)), "netstat")
		}
		msg := fmt.Sprintf("Failed to list listening ports. Error: %v\n%s", err, strings.TrimSpace(string(out)))
		if hint := permissionHint(err, string(out)); hint != "" {
			msg += "\n" + hint
		}
		return msg
	}
	return renderPorts(parseSSPorts(string(out)), "ss")
}

func renderPorts(entries []portEntry, source string) string {
	for _, e := range entries {
		listeningPort.WithLabelValues(e.port, e.process).Set(1)
	}

	if len(entries) == 0 {
		return "No listening TCP ports found."
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Listening TCP ports (%d, via %s):\n\n", len(entries), source)
	for _, e := range entries {
		if e.process != "" {
			fmt.Fprintf(&b, "  %-25s %s (pid %s)\n", e.address+":"+e.port, e.process, e.pid)
		} else {
			fmt.Fprintf(&b, "  %-25s (process unknown)\n", e.address+":"+e.port)
		}
	}
	return b.String()
}

// parseSSPorts parses `ss -tlnp` output rows.
func parseSSPorts(raw string) []portEntry {
	var entries []portEntry
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "State") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}
		local := fields[3]
		idx := strings.LastIndex(local, ":")
		if idx < 0 {
			continue
		}
		e := portEntry{
			address: strings.TrimSuffix(local[:idx], "."),
			port:    local[idx+1:],
		}
		if m := ssProcessRe.FindStringSubmatch(line); m != nil {
			e.process = m[1]
			e.pid = m[2]
		}
		entries = append(entries, e)
	}
	return entries
}

// parseNetstatPorts parses `netstat -tlnp` output rows:
// "tcp 0 0 0.0.0.0:22 0.0.0.0:* LISTEN 800/sshd".
func parseNetstatPorts(raw string) []portEntry {
	var entries []portEntry
	for _, line := range strings.Split(raw, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 7 || (fields[0] != "tcp" && fields[0] != "tcp6") {
			continue
		}
		local := fields[3]
		idx := strings.LastIndex(local, ":")
		if idx < 0 {
			continue
		}
		e := portEntry{
			address: strings.TrimSuffix(local[:idx], "."),
			port:    local[idx+1:],
		}
		if pid := fields[6]; pid != "-" && pid != "" {
			if i := strings.Index(pid, "/"); i >= 0 {
				e.pid = pid[:i]
				e.process = pid[i+1:]
			}
		}
		entries = append(entries, e)
	}
	return entries
}
