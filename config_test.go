package main

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "osctl.conf")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("cannot write config file: %v", err)
	}
	return path
}

func TestLoadConfigFile(t *testing.T) {
	path := writeTempConfig(t, `
# comment line
OSCTL_PORT=19877
OSCTL_USERNAME = alice
INVALID_KEY=ignored
nonsense line without equals
OSCTL_PASSWORD=secret123
`)

	t.Setenv("OSCTL_CONFIG", path)
	for _, k := range []string{"OSCTL_PORT", "OSCTL_USERNAME", "OSCTL_PASSWORD"} {
		os.Unsetenv(k)
	}
	t.Cleanup(func() {
		for _, k := range []string{"OSCTL_PORT", "OSCTL_USERNAME", "OSCTL_PASSWORD"} {
			os.Unsetenv(k)
		}
	})

	if err := loadConfigFile(path); err != nil {
		t.Fatalf("loadConfigFile error: %v", err)
	}

	if got := os.Getenv("OSCTL_PORT"); got != "19877" {
		t.Errorf("OSCTL_PORT = %q, want 19877", got)
	}
	if got := os.Getenv("OSCTL_USERNAME"); got != "alice" {
		t.Errorf("OSCTL_USERNAME = %q, want alice", got)
	}
	if got := os.Getenv("OSCTL_PASSWORD"); got != "secret123" {
		t.Errorf("OSCTL_PASSWORD = %q, want secret123", got)
	}
}

func TestLoadConfigFileDoesNotOverrideEnvironment(t *testing.T) {
	path := writeTempConfig(t, "OSCTL_USERNAME=bob\n")
	t.Setenv("OSCTL_USERNAME", "environment-wins")
	t.Cleanup(func() { os.Unsetenv("OSCTL_USERNAME") })

	if err := loadConfigFile(path); err != nil {
		t.Fatalf("loadConfigFile error: %v", err)
	}
	if got := os.Getenv("OSCTL_USERNAME"); got != "environment-wins" {
		t.Fatalf("environment must win over config file, got %q", got)
	}
}

func TestLoadConfigFileMissing(t *testing.T) {
	if err := loadConfigFile("/nonexistent/osctl.conf"); err == nil {
		t.Error("missing file should return an error for the caller to log")
	}
	if err := loadConfigFile(""); err != nil {
		t.Fatalf("empty path should be a no-op, got %v", err)
	}
}
