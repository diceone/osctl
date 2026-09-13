package main

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestParseFailedUnits(t *testing.T) {
	raw := "nginx.service   loaded failed failed A failed unit\nsshd.service    loaded failed failed B\n"
	units := parseFailedUnits(raw)
	if len(units) != 2 || units[0] != "nginx.service" || units[1] != "sshd.service" {
		t.Errorf("parseFailedUnits = %v", units)
	}
	if got := parseFailedUnits("\n \n"); len(got) != 0 {
		t.Errorf("expected no units, got %v", got)
	}
}

func TestParseSSPorts(t *testing.T) {
	raw := `State   Recv-Q  Send-Q  Local Address:Port  Peer Address:Port  Process
LISTEN 0      128        0.0.0.0:22         0.0.0.0:*    users:(("sshd",pid=800,fd=3))
LISTEN 0      128           [::]:22            [::]:*    users:(("sshd",pid=800,fd=4))
LISTEN 0      511        0.0.0.0:80         0.0.0.0:*`
	entries := parseSSPorts(raw)
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d: %v", len(entries), entries)
	}
	if entries[0].port != "22" || entries[0].process != "sshd" || entries[0].pid != "800" {
		t.Errorf("entry 0 wrong: %+v", entries[0])
	}
	if entries[1].address != "[::]" {
		t.Errorf("entry 1 address = %q", entries[1].address)
	}
	if entries[2].process != "" {
		t.Errorf("entry 2 should have no process: %+v", entries[2])
	}
}

func TestParseNetstatPorts(t *testing.T) {
	raw := `Active Internet connections (only servers)
Proto Recv-Q Send-Q Local Address           Foreign Address         State       PID/Program name
tcp        0      0 0.0.0.0:22              0.0.0.0:*               LISTEN      800/sshd
tcp6       0      0 :::80                   :::*                    LISTEN      -`
	entries := parseNetstatPorts(raw)
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].port != "22" || entries[0].process != "sshd" || entries[0].pid != "800" {
		t.Errorf("entry 0 wrong: %+v", entries[0])
	}
	if entries[1].process != "" {
		t.Errorf("entry 1 should have no process: %+v", entries[1])
	}
}

func TestParseLsmod(t *testing.T) {
	raw := "Module                  Size  Used by\nxt_tcpudp              16384  2\nip6table_filter        16384  1\nbridge                327680  0\n"
	names, total := parseLsmod(raw)
	if total != 3 || len(names) != 3 || names[0] != "xt_tcpudp" || names[1] != "ip6table_filter" || names[2] != "bridge" {
		t.Errorf("parseLsmod = %v, %d", names, total)
	}
}

func TestParseTimedatectl(t *testing.T) {
	raw := `               Local time: So 2026-09-13 16:00:00 CEST
           Universal time: So 2026-09-13 14:00:00 UTC
                 RTC time: So 2026-09-13 14:00:00
                Time zone: Europe/Berlin (CEST, +0200)
System clock synchronized: yes
              NTP service: active
          RTC in local TZ: no`
	status := parseTimedatectl(raw)
	if status["synchronized"] != "yes" {
		t.Errorf("synchronized = %q", status["synchronized"])
	}
	if status["ntp"] != "active" {
		t.Errorf("ntp = %q", status["ntp"])
	}
	if status["timezone"] != "Europe/Berlin (CEST, +0200)" {
		t.Errorf("timezone = %q", status["timezone"])
	}
	if status["rtcLocal"] != "no" {
		t.Errorf("rtcLocal = %q", status["rtcLocal"])
	}
	if status["local"] == "" {
		t.Errorf("local time not captured")
	}
}

func TestRenderPorts(t *testing.T) {
	entries := []portEntry{
		{address: "0.0.0.0", port: "22", process: "sshd", pid: "800"},
		{address: "0.0.0.0", port: "80"},
	}
	out := renderPorts(entries, "ss")
	if !strings.Contains(out, "(2, via ss)") {
		t.Errorf("missing count header: %q", out)
	}
	if !strings.Contains(out, "sshd (pid 800)") || !strings.Contains(out, "0.0.0.0:22") {
		t.Errorf("missing process entry: %q", out)
	}
	if !strings.Contains(out, "(process unknown)") {
		t.Errorf("missing unknown-process entry: %q", out)
	}
	if got := renderPorts(nil, "ss"); got != "No listening TCP ports found." {
		t.Errorf("empty output = %q", got)
	}
}

func TestGenerateReport(t *testing.T) {
	out := generateReport()
	var payload map[string]any
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("report is not valid JSON: %v\n%s", err, out)
	}
	if _, ok := payload["generated"]; !ok {
		t.Error("missing generated timestamp")
	}
	if _, ok := payload["host"]; !ok {
		t.Error("missing host")
	}
	if payload["osctl"] != buildVersion {
		t.Errorf("osctl = %v, want %s", payload["osctl"], buildVersion)
	}
	sections, ok := payload["sections"].(map[string]any)
	if !ok {
		t.Fatalf("sections missing")
	}
	for _, name := range []string{"ram", "disk", "failed_units", "listening_ports", "timesync", "updates", "osinfo"} {
		if _, ok := sections[name]; !ok {
			t.Errorf("missing section %s", name)
		}
	}
	if _, isStr := sections["ram"].(string); !isStr {
		t.Errorf("ram section is not a string")
	}
	if gen, _ := payload["generated"].(string); gen != "" {
		if _, err := time.Parse(time.RFC3339, gen); err != nil {
			t.Errorf("generated is not RFC3339: %q", gen)
		}
	}
}

func TestGetHardwareInfoShape(t *testing.T) {
	// On the build machine (macOS) lspci/lsusb/lsmod may not exist; the
	// function must always produce the section headers.
	out := getHardwareInfo()
	for _, want := range []string{"PCI devices (lspci):", "USB devices (lsusb):", "Kernel modules (lsmod):"} {
		if !strings.Contains(out, want) {
			t.Errorf("hwinfo output missing %q", want)
		}
	}
}
