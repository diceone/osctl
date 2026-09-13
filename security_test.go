package main

import (
	"os"
	"strings"
	"testing"
)

func TestPermissionHint(t *testing.T) {
	err := &os.PathError{Op: "open", Path: "/etc/ssh/sshd_config", Err: os.ErrPermission}

	if hint := permissionHint(err, ""); !strings.Contains(hint, "sudo") {
		t.Fatalf("permissionHint(ErrPermission) = %q, want a sudo hint", hint)
	}
	if hint := permissionHint(nil, "anything"); hint != "" {
		t.Fatalf("permissionHint(nil, ...) = %q, want empty", hint)
	}
	if hint := permissionHint(os.ErrNotExist, ""); hint != "" {
		t.Fatalf("permissionHint(ErrNotExist) = %q, want empty", hint)
	}
	hint := permissionHint(os.ErrPermission, "")
	if os.Geteuid() == 0 {
		if hint != "" {
			t.Fatalf("running as root must suppress the hint, got %q", hint)
		}
	} else if hint == "" {
		t.Fatal("non-root should get a sudo hint")
	}
}

func TestUnprivilegedCommandsSuggestSudo(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("running as root: privilege hints are suppressed")
	}

	cases := map[string]func() string{
		"shutdown":       func() string { return shutdownSystem() },
		"reboot":         func() string { return rebootSystem() },
		"update":         func() string { return updatePackages() },
		"service start":  func() string { return manageService("start", "sshd") },
		"service stop":   func() string { return manageService("stop", "sshd") },
		"service status": func() string { return manageService("status", "sshd") },
		"useradd":        func() string { return addUser("testhintuser") },
		"userdel":        func() string { return deleteUser("testhintuser") },
	}

	for name, fn := range cases {
		msg := fn()
		switch name {
		case "service status":
			// Read-only: must not be blocked.
			if strings.Contains(msg, "sudo") {
				t.Errorf("%s must not require root, got: %s", name, msg)
			}
		default:
			if !strings.Contains(msg, "sudo") {
				t.Errorf("%s output lacks a sudo hint, got: %s", name, msg)
			}
			if !strings.HasPrefix(msg, "Failed") {
				t.Errorf("%s output must start with an error prefix for HTTP/exit-code mapping, got: %s", name, msg)
			}
		}
	}
}

func TestCheckFilePermissionsUsesThreeDigitOctal(t *testing.T) {
	out := checkFilePermissions()
	if strings.Contains(out, " 0644 ") || strings.Contains(out, " 0000 ") {
		t.Fatalf("expected three-digit octal notation, got: %s", out)
	}
	if !strings.Contains(out, ": 644 (expected: 644)") && !strings.Contains(out, "Not found or not accessible") {
		t.Fatalf("unexpected permissions output: %s", out)
	}
}

func TestCheckSuspiciousFilesExplainsScanResult(t *testing.T) {
	if os.Geteuid() != 0 {
		t.Skip("full-filesystem find scan is only reliable as root; skipping on non-root")
	}
	out := checkSuspiciousFiles()
	if strings.Contains(out, "none found or cannot scan") {
		t.Fatalf("ambiguous message should be gone, got: %s", out)
	}
	if !strings.Contains(out, "(none found)") && !strings.Contains(out, "World-writable") {
		t.Fatalf("unexpected audit files output: %s", out)
	}
}
