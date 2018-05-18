package tigerblood

import (
	"fmt"
	"net"
	"strings"
)

// IPAddressFromHTTPPath takes a HTTP path and returns an IPv4 IP if it's found, or an error if none is found.
func IPAddressFromHTTPPath(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("Invalid path")
	}
	n := &net.IPNet{}
	comp := strings.Split(path, "/")
	ip := net.ParseIP(comp[len(comp)-1])
	if ip != nil {
		if ip.To4() != nil {
			n.Mask = net.CIDRMask(32, 32)
		} else if ip.To16() != nil {
			n.Mask = net.CIDRMask(128, 128)
		} else {
			return "", fmt.Errorf("Error getting IP from HTTP path")
		}
		n.IP = ip
		return n.String(), nil

	}
	// Otherwise, treat as CIDR
	if len(comp) < 2 {
		return "", fmt.Errorf("Error getting IP from HTTP path")
	}
	ip, n, err := net.ParseCIDR(strings.Join(comp[len(comp)-2:len(comp)], "/"))
	if err != nil {
		return "", fmt.Errorf("Error getting IP from HTTP path: %s", err)
	}
	return n.String(), nil
}
