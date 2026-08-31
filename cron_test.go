package main

import (
	"strings"
	"testing"
)

func TestValidateCronSchedule(t *testing.T) {
	valid := []string{
		"0 2 * * *",
		"*/5 * * * *",
		"0 12 1 */2 1-5",
		"30 4 * * JAN",
		"15,45 */6 * * mon-fri",
	}
	for _, s := range valid {
		if err := validateCronSchedule(s); err != nil {
			t.Errorf("validateCronSchedule(%q) unexpected error: %v", s, err)
		}
	}

	invalid := []string{
		"",                        // empty
		"0 2 * *",                 // 4 fields
		"0 2 * * * *",             // 6 fields
		"0 2 * * * reboot -h now", // injected command
		"a;b * * * *",             // forbidden char
		"0 2 * * *\nrm -rf /",     // newline injection
		"$(reboot) * * * *",       // command substitution
	}
	for _, s := range invalid {
		if err := validateCronSchedule(s); err == nil {
			t.Errorf("validateCronSchedule(%q) expected error, got nil", s)
		}
	}
}

func TestValidateCronCommand(t *testing.T) {
	valid := []string{
		"/usr/local/bin/backup.sh",
		"echo hello >> /var/log/osctl.log",
		"find /tmp -name 'osctl-*' -delete",
	}
	for _, c := range valid {
		if err := validateCronCommand(c); err != nil {
			t.Errorf("validateCronCommand(%q) unexpected error: %v", c, err)
		}
	}

	invalid := []string{
		"",
		strings.Repeat("a", 2000),  // over length cap
		"echo one\necho two\nsu -", // newline injection
		"echo\x00null",             // NUL byte
		"echo \x01bell",            // other control char
	}
	for _, c := range invalid {
		if err := validateCronCommand(c); err == nil {
			t.Errorf("validateCronCommand(%q) expected error, got nil", c)
		}
	}
}
