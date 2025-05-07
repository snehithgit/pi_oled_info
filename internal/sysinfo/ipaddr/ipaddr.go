package ipaddr

import (
	"fmt"
	"net"
)

// Get returns the primary IP address as a formatted string
func Get() (string, error) {
	// Get the list of network interfaces
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, iface := range ifaces {
		// Skip loopback and non-up interfaces
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			// Check if this is an IP network address
			switch v := addr.(type) {
			case *net.IPNet:
				ip := v.IP
				// Skip IPv6 and loopback addresses
				if ip.To4() != nil && !ip.IsLoopback() {
					return fmt.Sprintf("IP: %s", ip.String()), nil
				}
			}
		}
	}

	return "IP: Not found", nil
}
