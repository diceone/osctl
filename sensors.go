package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// getSensors lists temperatures and fan speeds from /sys/class/hwmon.
func getSensors() string {
	return getSensorsDir("/sys/class/hwmon")
}

// getSensorsDir implements getSensors against an injectable hwmon directory
// so tests can point it at a fixture tree.
func getSensorsDir(dir string) string {
	devices, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Sprintf("Failed to read sensors. Error: %v", err)
	}

	var output strings.Builder
	output.WriteString("Hardware Sensors:\n")
	found := false

	for _, dev := range devices {
		if !dev.IsDir() {
			continue
		}
		base := filepath.Join(dir, dev.Name())
		name := readTrimmed(filepath.Join(base, "name"))
		if name == "" {
			name = dev.Name()
		}

		files, err := os.ReadDir(base)
		if err != nil {
			continue
		}

		var lines []string
		for _, f := range files {
			switch {
			case strings.HasPrefix(f.Name(), "temp") && strings.HasSuffix(f.Name(), "_input"):
				prefix := strings.TrimSuffix(f.Name(), "_input")
				label := "temp"
				if l := readTrimmed(filepath.Join(base, prefix+"_label")); l != "" {
					label = l
				}
				if t, ok := readMilli(base, prefix); ok {
					sensorTempCelsius.WithLabelValues(name, label).Set(t)
					lines = append(lines, fmt.Sprintf("  %s: %.1f°C", label, t))
				}
			case strings.HasPrefix(f.Name(), "fan") && strings.HasSuffix(f.Name(), "_input"):
				prefix := strings.TrimSuffix(f.Name(), "_input")
				label := "fan"
				if l := readTrimmed(filepath.Join(base, prefix+"_label")); l != "" {
					label = l
				}
				if rpm, ok := readInt(filepath.Join(base, f.Name())); ok {
					sensorFanRPM.WithLabelValues(name, label).Set(float64(rpm))
					lines = append(lines, fmt.Sprintf("  %s: %d RPM", label, rpm))
				}
			}
		}
		if len(lines) > 0 {
			found = true
			output.WriteString(fmt.Sprintf("%s (%s):\n", name, base))
			for _, l := range lines {
				output.WriteString(l + "\n")
			}
		}
	}

	if !found {
		return "No temperature or fan sensors found."
	}
	return output.String()
}

// readMilli reads a milli-unit input file and returns the value scaled down
// (e.g. milli-Celsius -> Celsius).
func readMilli(base, prefix string) (float64, bool) {
	v, ok := readInt(filepath.Join(base, prefix+"_input"))
	if !ok {
		return 0, false
	}
	return float64(v) / 1000.0, true
}

func readInt(path string) (int64, bool) {
	s := readTrimmed(path)
	if s == "" {
		return 0, false
	}
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, false
	}
	return v, true
}

func readTrimmed(path string) string {
	data, err := os.ReadFile(path) //nolint:gosec // kernel sysfs paths
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}
