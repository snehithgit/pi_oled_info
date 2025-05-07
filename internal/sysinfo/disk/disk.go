package disk

import (
	"fmt"

	"github.com/shirou/gopsutil/disk"
)

// Get returns disk usage information as a formatted string
func Get() (string, error) {
	// Get usage statistics for the root filesystem
	usage, err := disk.Usage("/")
	if err != nil {
		return "", err
	}

	// Format the disk usage to fit in the limited space
	usedGB := float64(usage.Used) / (1024 * 1024 * 1024)
	totalGB := float64(usage.Total) / (1024 * 1024 * 1024)
	percentUsed := usage.UsedPercent

	return fmt.Sprintf("Disk: %.1f/%.1fG %d%%", usedGB, totalGB, int(percentUsed)), nil
}
