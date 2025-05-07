package cpu

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	"github.com/shirou/gopsutil/cpu"
)

// Get returns CPU usage percentage and temperature as a formatted string
func Get() (string, error) {
	// Get CPU usage percentage
	percent, err := cpu.Percent(time.Second, false)
	if err != nil {
		return "", err
	}

	usage := 0.0
	if len(percent) > 0 {
		usage = percent[0]
	}

	// Get CPU temperature (Raspberry Pi specific)
	temp, err := getCPUTemperature()
	if err != nil {
		// If we can't get temperature, just show usage
		return fmt.Sprintf("CPU: %.1f%%", usage), nil
	}

	// Format both usage and temperature
	return fmt.Sprintf("CPU: %.1f%% %.1f°C", usage, temp), nil
}

// getCPUTemperature reads the CPU temperature from the Raspberry Pi thermal zone
func getCPUTemperature() (float64, error) {
	// Read from thermal zone
	content, err := ioutil.ReadFile("/sys/class/thermal/thermal_zone0/temp")
	if err != nil {
		return 0, err
	}

	// Convert the temperature (in milliCelsius) to Celsius
	tempStr := strings.TrimSpace(string(content))
	tempMilliC, err := strconv.ParseInt(tempStr, 10, 64)
	if err != nil {
		return 0, err
	}

	return float64(tempMilliC) / 1000.0, nil
}
