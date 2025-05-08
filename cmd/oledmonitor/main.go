package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/snehithgit/pi_oled_info/internal/display"
	"github.com/snehithgit/pi_oled_info/internal/sysinfo/cpu"
	"github.com/snehithgit/pi_oled_info/internal/sysinfo/disk"
	"github.com/snehithgit/pi_oled_info/internal/sysinfo/ipaddr"
	"github.com/snehithgit/pi_oled_info/internal/sysinfo/uptime"
)

func main() {
	// Initialize the display
	disp, err := display.New()
	if err != nil {
		log.Fatalf("Failed to initialize display: %v", err)
	}
	defer disp.Close()

	// Setup graceful shutdown
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	done := make(chan bool, 1)

	// Run the display update in a goroutine
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		// Initial display update
		updateDisplay(disp)

		// Loop for periodic updates
		for {
			select {
			case <-ticker.C:
				updateDisplay(disp)
			case <-done:
				return
			}
		}
	}()

	// Wait for signal
	<-sigs
	log.Println("Shutting down...")
	done <- true
	time.Sleep(250 * time.Millisecond) // Give display time to clean up
}

func updateDisplay(disp *display.Display) {
	disp.Clear()

	// Line 1: IP address
	ip, err := ipaddr.Get()
	if err != nil {
		ip = "IP: Error"
		log.Printf("Error getting IP address: %v", err)
	}
	disp.WriteLine(0, ip)

	// Line 2: System uptime
	upStr, err := uptime.Get()
	if err != nil {
		upStr = "Up: Error"
		log.Printf("Error getting uptime: %v", err)
	}
	disp.WriteLine(1, upStr)

	// Line 3: CPU usage and temperature
	cpuInfo, err := cpu.Get()
	if err != nil {
		cpuInfo = "CPU: Error"
		log.Printf("Error getting CPU info: %v", err)
	}
	disp.WriteLine(2, cpuInfo)

	// Line 4: Disk usage
	diskInfo, err := disk.Get()
	if err != nil {
		diskInfo = "Disk: Error"
		log.Printf("Error getting disk info: %v", err)
	}
	disp.WriteLine(3, diskInfo)

	// Update the display
	if err := disp.Update(); err != nil {
		log.Printf("Error updating display: %v", err)
	}
}
