package main

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// validUsername matches the Debian/Ubuntu login name rules: starts with a
// lowercase letter or underscore, then lowercase letters/digits/hyphen/
// underscore, max 32 chars.
var validUsername = regexp.MustCompile(`^[a-z_][a-z0-9_-]{0,31}$`)

func validateUsername(name string) error {
	if !validUsername.MatchString(name) {
		return fmt.Errorf("invalid username %q (allowed: lowercase letters, digits, '-', '_', max 32 chars, must start with a letter or '_')", name)
	}
	return nil
}

// getUserInfo shows identity and password-age info for a user.
func getUserInfo(name string) string {
	if err := validateUsername(name); err != nil {
		return "Invalid username: " + err.Error()
	}

	var output strings.Builder

	idOut, err := exec.Command("id", name).CombinedOutput()
	if err != nil {
		return fmt.Sprintf("Failed to find user %s. Error: %v", name, err)
	}
	output.WriteString(fmt.Sprintln(strings.TrimSpace(string(idOut))))

	if chageOut, err := exec.Command("chage", "-l", name).Output(); err == nil {
		output.WriteString("\nPassword aging:\n")
		output.WriteString(string(chageOut))
	}

	return strings.TrimRight(output.String(), "\n")
}

// addUser creates a user with a home directory.
func addUser(name string) string {
	if err := validateUsername(name); err != nil {
		return "Invalid username: " + err.Error()
	}

	out, err := exec.Command("useradd", "-m", name).CombinedOutput()
	if err != nil {
		return fmt.Sprintf("Failed to create user %s. Error: %v\n%s", name, err, string(out))
	}
	return fmt.Sprintf("User %s created (with home directory). Set a password with: passwd %s", name, name)
}

// deleteUser removes a user and their home directory. Protected accounts are
// refused so the tool can never gut the system it runs on.
var protectedUsers = map[string]bool{
	"root":   true,
	"daemon": true,
	"bin":    true,
	"sys":    true,
	"sync":   true,
	"man":    true,
}

func deleteUser(name string) string {
	if err := validateUsername(name); err != nil {
		return "Invalid username: " + err.Error()
	}
	if protectedUsers[name] {
		return fmt.Sprintf("Refusing to delete protected account %q", name)
	}

	out, err := exec.Command("userdel", "-r", name).CombinedOutput()
	if err != nil {
		return fmt.Sprintf("Failed to delete user %s. Error: %v\n%s", name, err, string(out))
	}
	return fmt.Sprintf("User %s deleted (including home directory).", name)
}
