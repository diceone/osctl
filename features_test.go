package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestSummarizeUpdates(t *testing.T) {
	debian := summarizeUpdates("debian", "Reading package lists...\nopenssl/now 3.0 1 amd64 [upgradable from: 3.0]\nbase-files/now 1 amd64 [upgradable from: 1]\n")
	if !strings.Contains(debian, "Available package updates: 2") {
		t.Fatalf("debian summary wrong: %q", debian)
	}
	if strings.Contains(debian, "Reading package lists") {
		t.Fatalf("debian summary kept non-upgradable line: %q", debian)
	}

	suse := summarizeUpdates("suse", "v | repo | bash | 5.0 | 5.1 | x86_64\nu | repo | zsh | 5.9 | 6.0 | x86_64\nheader line\n")
	if !strings.Contains(suse, "Available package updates: 2") || strings.Contains(suse, "header line") {
		t.Fatalf("suse summary wrong: %q", suse)
	}

	rhel := summarizeUpdates("rhel", "  bash-5.1.8-9.el9.x86_64 \n\n kernel-5.14.0.x86_64 \n")
	if !strings.Contains(rhel, "Available package updates: 2") {
		t.Fatalf("rhel summary wrong: %q", rhel)
	}

	if got := summarizeUpdates("debian", "Reading package lists...\nBuilding dependency tree...\n"); got != "No package updates available." {
		t.Fatalf("empty summary wrong: %q", got)
	}
}

func TestGetServiceLogsInvalidUnit(t *testing.T) {
	cases := []struct{ unit, lines string }{
		{"", "50"},
		{"nginx;reboot", "50"},
		{"nginx|rm", "50"},
		{"nginx & shutdown", "50"},
		{"nginx$IFS", "50"},
		{"nginx`id`", "50"},
		{"nginx\nfoo", "50"},
		{"nginx\rfoo", "50"},
		{strings.Repeat("a", 129), "50"},
		{"nginx", "0"},
		{"nginx", "10001"},
		{"nginx", "abc"},
	}
	for _, tc := range cases {
		out := getServiceLogs(tc.unit, tc.lines)
		if !strings.HasPrefix(out, "Invalid") && !strings.HasPrefix(out, "Usage:") {
			t.Errorf("getServiceLogs(%q, %q) = %q, want Usage/Invalid prefix", tc.unit, tc.lines, out)
		}
	}
}

func TestParseDockerNumbers(t *testing.T) {
	if got := parsePercentValue("12.5%"); got != 12.5 {
		t.Errorf("parsePercentValue = %v, want 12.5", got)
	}
	if got := parsePercentValue(" 0.0 "); got != 0 {
		t.Errorf("parsePercentValue = %v, want 0", got)
	}

	used, limit := parseDockerMemUsage("12.5MiB / 1.944GiB")
	if used != 12.5*1024*1024 {
		t.Errorf("used = %v", used)
	}
	if limit != 1.944*1024*1024*1024 {
		t.Errorf("limit = %v", limit)
	}

	cases := map[string]float64{
		"0B":      0,
		"12.5MiB": 12.5 * 1024 * 1024,
		"318.7MB": 318.7 * 1e6,
		"2GiB":    2 * 1024 * 1024 * 1024,
		"5kB":     5e3,
		"1TB":     1e12,
		"1TiB":    1 * 1024 * 1024 * 1024 * 1024,
	}
	for in, want := range cases {
		if got := parseDockerSize(in); got != want {
			t.Errorf("parseDockerSize(%q) = %v, want %v", in, got, want)
		}
	}
	if got := parseDockerSize("nonsense"); got != 0 {
		t.Errorf("parseDockerSize(nonsense) = %v, want 0", got)
	}
}

func TestGetSensorsDir(t *testing.T) {
	dir := t.TempDir()
	dev := filepath.Join(dir, "hwmon0")
	if err := os.Mkdir(dev, 0o755); err != nil {
		t.Fatal(err)
	}
	files := map[string]string{
		"name":        "coretemp",
		"temp1_input": "45000",
		"temp1_label": "Package id 0",
		"temp2_input": "39500",
		"fan1_input":  "1200",
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dev, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	out := getSensorsDir(dir)
	for _, want := range []string{"coretemp", "Package id 0: 45.0°C", "39.5°C", "1200 RPM"} {
		if !strings.Contains(out, want) {
			t.Errorf("sensors output missing %q in %q", want, out)
		}
	}

	if got := getSensorsDir(filepath.Join(dir, "missing")); !strings.HasPrefix(got, "Failed") {
		t.Errorf("missing dir should fail, got %q", got)
	}
}

func TestSysctlChecks(t *testing.T) {
	if len(sysctlChecks) != 10 {
		t.Fatalf("expected 10 sysctl checks, got %d", len(sysctlChecks))
	}
	byName := map[string]sysctlCheck{}
	for _, c := range sysctlChecks {
		byName[c.name] = c
	}
	assert := func(name string, v int64, want bool) {
		t.Helper()
		c, ok := byName[name]
		if !ok {
			t.Fatalf("missing check %s", name)
		}
		if got := c.check(v); got != want {
			t.Errorf("%s(%d) = %v, want %v", name, v, got, want)
		}
	}
	assert("kernel.randomize_va_space", 2, true)
	assert("kernel.randomize_va_space", 1, false)
	assert("kernel.kptr_restrict", 1, true)
	assert("kernel.kptr_restrict", 0, false)
	assert("kernel.dmesg_restrict", 1, true)
	assert("fs.protected_hardlinks", 1, true)
	assert("fs.protected_hardlinks", 0, false)
	assert("fs.suid_dumpable", 0, true)
	assert("fs.suid_dumpable", 2, false)
	assert("net.ipv4.conf.all.accept_redirects", 0, true)
	assert("net.ipv4.conf.all.send_redirects", 0, true)
	assert("net.ipv4.ip_forward", 0, true)
	assert("net.ipv4.ip_forward", 1, false)

	for _, c := range sysctlChecks {
		if c.name == "net.ipv4.ip_forward" {
			if !c.infoOnly {
				t.Errorf("ip_forward should be infoOnly")
			}
		} else if c.infoOnly {
			t.Errorf("%s should not be infoOnly", c.name)
		}
	}
}

func TestApparmorLoadedProfiles(t *testing.T) {
	if got := apparmorLoadedProfiles("5 profiles are loaded.\n12 profiles are in enforce mode."); got != "5" {
		t.Errorf("apparmorLoadedProfiles = %q, want 5", got)
	}
	if got := apparmorLoadedProfiles("nothing here"); got != "unknown" {
		t.Errorf("apparmorLoadedProfiles = %q, want unknown", got)
	}
}

func TestCompletionScripts(t *testing.T) {
	names := commandNames()
	if len(names) < 40 {
		t.Fatalf("too few commands registered: %d", len(names))
	}
	for _, shell := range []string{"bash", "zsh", "fish"} {
		out := completionCommand(shell)
		for _, cmd := range names {
			if !strings.Contains(out, cmd) {
				t.Errorf("%s completion missing command %q", shell, cmd)
			}
		}
		for _, group := range []string{"audit", "cron", "maintenance", "process", "service", "completion"} {
			if !strings.Contains(out, group) {
				t.Errorf("%s completion missing subcommand group %q", shell, group)
			}
		}
	}
	if got := completionCommand(""); !strings.HasPrefix(got, "Usage:") {
		t.Errorf("completionCommand(\"\") = %q, want Usage prefix", got)
	}
	if got := completionCommand("powershell"); !strings.HasPrefix(got, "Unsupported shell:") {
		t.Errorf("completionCommand(powershell) = %q", got)
	}
}

func TestRunWatchUsage(t *testing.T) {
	if got := runWatch([]string{"--interval", "0", "version"}); !strings.HasPrefix(got, "Invalid watch interval:") {
		t.Errorf("interval 0: %q", got)
	}
	if got := runWatch([]string{"-i", "90000", "version"}); !strings.HasPrefix(got, "Invalid watch interval:") {
		t.Errorf("interval 90000: %q", got)
	}
	if got := runWatch([]string{"--interval", "abc", "version"}); !strings.HasPrefix(got, "Invalid watch interval:") {
		t.Errorf("interval abc: %q", got)
	}
	if got := runWatch(nil); !strings.HasPrefix(got, "Usage:") {
		t.Errorf("no command: %q", got)
	}
	if got := runWatch([]string{"--interval", "5"}); !strings.HasPrefix(got, "Usage:") {
		t.Errorf("interval only: %q", got)
	}
}

func TestOpenAPISpec(t *testing.T) {
	var spec map[string]any
	if err := json.Unmarshal(openAPISpecJSON(), &spec); err != nil {
		t.Fatalf("spec is not valid JSON: %v", err)
	}
	if spec["openapi"] != "3.0.3" {
		t.Errorf("openapi = %v, want 3.0.3", spec["openapi"])
	}
	info, _ := spec["info"].(map[string]any)
	if info == nil || info["version"] != buildVersion {
		t.Errorf("info.version = %v, want %s", info["version"], buildVersion)
	}
	paths, _ := spec["paths"].(map[string]any)
	if paths["/ram"] == nil {
		t.Errorf("missing /ram path")
	}
	// The live spec lists API command routes only; /openapi.json and
	// /metrics are mux-level routes and intentionally not part of it.
	shutdown, ok := paths["/shutdown"].(map[string]any)
	if !ok {
		t.Fatalf("/shutdown path malformed")
	}
	if _, has := shutdown["post"]; !has {
		t.Errorf("/shutdown should be POST-only, methods: %v", shutdown)
	}
	ram, ok := paths["/ram"].(map[string]any)
	if !ok {
		t.Fatalf("/ram path malformed")
	}
	if _, has := ram["get"]; !has {
		t.Errorf("/ram should expose get")
	}
}

func TestV1PrefixRouting(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/v1/version", nil)
	rec := httptest.NewRecorder()
	handleRequest(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("/v1/version status = %d, want 200", rec.Code)
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("body not JSON: %v", err)
	}
	result, _ := body["result"].(string)
	if !strings.Contains(result, "osctl") {
		t.Errorf("result = %q, want version string", result)
	}
}

func TestEnvEnabled(t *testing.T) {
	for _, v := range []string{"1", "true", "YES", "on", "On"} {
		t.Setenv("OSCTL_TEST_FLAG", v)
		if !envEnabled("OSCTL_TEST_FLAG") {
			t.Errorf("envEnabled(%q) = false", v)
		}
	}
	for _, v := range []string{"0", "false", "no", "off", "", "maybe"} {
		t.Setenv("OSCTL_TEST_FLAG", v)
		if envEnabled("OSCTL_TEST_FLAG") {
			t.Errorf("envEnabled(%q) = true", v)
		}
	}
}

func generateSelfSignedCert(t *testing.T, days int) []byte {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "test-cert"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Duration(days) * 24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}
	der, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
}

func TestCheckCertificates(t *testing.T) {
	t.Run("valid cert file", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "cert.pem")
		if err := os.WriteFile(path, generateSelfSignedCert(t, 90), 0o600); err != nil {
			t.Fatal(err)
		}
		out := checkCertificates([]string{path})
		if !strings.Contains(out, "test-cert") || !strings.Contains(out, "OK") {
			t.Errorf("checkCertificates output = %q", out)
		}
		if !strings.Contains(out, "1 certificate") {
			t.Errorf("summary missing count: %q", out)
		}
	})

	t.Run("missing file", func(t *testing.T) {
		out := checkCertificates([]string{filepath.Join(t.TempDir(), "nope.pem")})
		if !strings.Contains(out, "file not found") {
			t.Errorf("missing file output = %q", out)
		}
	})

	t.Run("non-PEM file", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "bad.pem")
		if err := os.WriteFile(path, []byte("not a certificate"), 0o600); err != nil {
			t.Fatal(err)
		}
		out := checkCertificates([]string{path})
		if !strings.Contains(out, "Could not inspect") {
			t.Errorf("bad PEM output = %q", out)
		}
	})

	t.Run("host port detection", func(t *testing.T) {
		// Host:port and extensionless hostnames are network targets; dotted
		// names are ambiguous and intentionally treated as file paths.
		if !looksLikeHostPort("example.com:443") || !looksLikeHostPort("localhost") {
			t.Errorf("host targets should look like host:port")
		}
		if looksLikeHostPort("/etc/ssl/certs/ca.crt") || looksLikeHostPort("/etc/pki/tls/certs") {
			t.Errorf("paths should not look like host:port")
		}
	})
}
