package main

import (
	"strings"
	"testing"
)

func TestValidateUsername(t *testing.T) {
	valid := []string{"deploy", "_svc", "ci-runner", "user_1"}
	for _, n := range valid {
		if err := validateUsername(n); err != nil {
			t.Errorf("validateUsername(%q) unexpected error: %v", n, err)
		}
	}

	invalid := []string{
		"",
		"Root",        // uppercase
		"1abc",        // must start with letter/underscore
		"user name",   // space
		"../etc",      // traversal
		"user;id",     // metacharacters
		"user$whoami", // metacharacters
		"-leading",    // starts with dash (looks like a flag)
		string(make([]byte, 0)) + "a" + repeat('x', 40), // too long
	}
	for _, n := range invalid {
		if err := validateUsername(n); err == nil {
			t.Errorf("validateUsername(%q) expected error, got nil", n)
		}
	}
}

func repeat(c byte, n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = c
	}
	return string(b)
}

func TestDeleteUserProtectsSystemAccounts(t *testing.T) {
	for _, name := range []string{"root", "bin", "daemon"} {
		res := deleteUser(name)
		if !strings.Contains(res, "Refusing") {
			t.Errorf("deleteUser(%q) must refuse protected accounts, got %q", name, res)
		}
	}
}

func TestGetUserInfoRejectsBadName(t *testing.T) {
	if res := getUserInfo("bad name; id"); res[:8] != "Invalid " {
		t.Fatalf("expected rejection for invalid username, got %q", res)
	}
}
