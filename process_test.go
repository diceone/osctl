package main

import "testing"

func TestValidatePID(t *testing.T) {
	tests := []struct {
		name    string
		pid     string
		wantErr bool
	}{
		{"empty", "", true},
		{"not a number", "abc", true},
		{"zero", "0", true},
		{"negative", "-1", true},
		{"negative process group", "-5", true},
		{"overflow", "99999999999", true},
		{"with space", " 5", true},
		{"one", "1", false},
		{"hundred", "100", false},
		{"max pid", "4194304", false},
		{"just over max", "4194305", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := validatePID(tt.pid)
			if tt.wantErr && err == nil {
				t.Fatalf("validatePID(%q) expected error, got pid %d", tt.pid, got)
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("validatePID(%q) unexpected error: %v", tt.pid, err)
			}
		})
	}
}

func TestValidatePIDNegativeRejectsProcessGroups(t *testing.T) {
	// The historic bug: "-5" passed Atoi and was forwarded to kill, where a
	// negative pid means "the whole process group".
	if _, err := validatePID("-5"); err == nil {
		t.Fatal("negative PIDs must be rejected so they cannot address process groups")
	}
}
