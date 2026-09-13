package main

import (
	"fmt"
	"os/exec"
	"strings"
)

// getHardwareInfo inventories PCI/USB devices and loaded kernel modules.
func getHardwareInfo() string {
	var b strings.Builder
	b.WriteString("Hardware Information:\n\n")

	b.WriteString("PCI devices (lspci):\n")
	if out, err := exec.Command("lspci").Output(); err == nil {
		b.WriteString(indentBlock(string(out)))
	} else {
		b.WriteString("  (lspci not installed)\n")
	}

	b.WriteString("\nUSB devices (lsusb):\n")
	if out, err := exec.Command("lsusb").Output(); err == nil {
		b.WriteString(indentBlock(string(out)))
	} else {
		b.WriteString("  (lsusb not installed)\n")
	}

	b.WriteString("\nKernel modules (lsmod):\n")
	if out, err := exec.Command("lsmod").Output(); err == nil {
		names, total := parseLsmod(string(out))
		if total == 0 {
			b.WriteString("  (no modules loaded)\n")
		} else {
			shown := names
			more := ""
			if len(shown) > 15 {
				shown = shown[:15]
				more = fmt.Sprintf(" … (%d more)", total-15)
			}
			fmt.Fprintf(&b, "  %d modules loaded: %s%s\n", total, strings.Join(shown, ", "), more)
		}
	} else {
		b.WriteString("  (lsmod not installed)\n")
	}

	return b.String()
}

// parseLsmod returns the module names and count from `lsmod` output
// (first line is the header).
func parseLsmod(raw string) (names []string, total int) {
	for _, line := range strings.Split(raw, "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 || fields[0] == "Module" || fields[0] == "Module:" {
			continue
		}
		names = append(names, fields[0])
	}
	return names, len(names)
}

func indentBlock(s string) string {
	var b strings.Builder
	for _, line := range strings.Split(strings.TrimRight(s, "\n"), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		b.WriteString("  " + line + "\n")
	}
	return b.String()
}
