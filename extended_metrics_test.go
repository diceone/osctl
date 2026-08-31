package main

import "testing"

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		in   uint64
		want string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.00 KiB"},
		{1536, "1.50 KiB"},
		{1048576, "1.00 MiB"},
		{1073741824, "1.00 GiB"},
	}

	for _, tt := range tests {
		if got := formatBytes(tt.in); got != tt.want {
			t.Errorf("formatBytes(%d) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestGetProcessStateDescription(t *testing.T) {
	tests := map[string]string{
		"R":  "Running",
		"S":  "Sleeping",
		"Z":  "Zombie",
		"Q":  "Unknown", // unmapped state
		"":   "Unknown",
		"xx": "Unknown",
	}

	for state, want := range tests {
		if got := getProcessStateDescription(state); got != want {
			t.Errorf("getProcessStateDescription(%q) = %q, want %q", state, got, want)
		}
	}
}
