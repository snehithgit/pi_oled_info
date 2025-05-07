package uptime

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"time"
)

// Get returns the system uptime as a formatted string
func Get() (string, error) {
	// Read uptime from proc filesystem
	content, err := ioutil.ReadFile("/proc/uptime")
	if err != nil {
		return "", err
	}

	// Parse the uptime value (first number in the file)
	fields := strings.Fields(string(content))
	if len(fields) < 1 {
		return "", fmt.Errorf("invalid uptime format")
	}

	uptimeSeconds, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return "", err
	}

	// Convert to duration
	uptime := time.Duration(uptimeSeconds) * time.Second

	// Format uptime in a compact way
	days := int(uptime.Hours() / 24)
	hours := int(uptime.Hours()) % 24
	mins := int(uptime.Minutes()) % 60

	if days > 0 {
		return fmt.Sprintf("Up: %dd %dh %dm", days, hours, mins), nil
	}
	return fmt.Sprintf("Up: %dh %dm", hours, mins), nil
}
