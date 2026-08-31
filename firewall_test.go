package main

import "testing"

func TestValidatePortSpec(t *testing.T) {
	valid := []string{"8080", "80", "8080/tcp", "53/udp", "1", "65535"}
	for _, s := range valid {
		if _, err := validatePortSpec(s); err != nil {
			t.Errorf("validatePortSpec(%q) unexpected error: %v", s, err)
		}
	}

	invalid := []string{"", "0", "65536", "99999", "8080/icmp", "abc", "80; rm -rf /", "-1", "8080/", "/tcp"}
	for _, s := range invalid {
		if _, err := validatePortSpec(s); err == nil {
			t.Errorf("validatePortSpec(%q) expected error, got nil", s)
		}
	}
}

func TestManageFirewallRuleRejectsInvalidInput(t *testing.T) {
	if res := manageFirewallRule("explode", "80/tcp"); res == "" || res == "Firewall rule applied" {
		t.Fatalf("invalid action must be rejected, got %q", res)
	}
	if res := manageFirewallRule("allow", "65536"); res == "Firewall rule applied" {
		t.Fatalf("invalid port must be rejected, got %q", res)
	}
}
