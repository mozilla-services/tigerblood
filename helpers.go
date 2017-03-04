package tigerblood

import (
	"fmt"
	"net"
	"strings"
)

// IPAddressFromHTTPPath takes a HTTP path and returns an IPv4 IP if it's found, or an error if none is found.
func IPAddressFromHTTPPath(path string) (string, error) {
	path = path[1:len(path)]
	ip, network, err := net.ParseCIDR(path)
	if err != nil {
		if strings.Contains(path, "/") {
			return "", fmt.Errorf("Error getting IP from HTTP path: %s", err)
		}
		ip = net.ParseIP(path)
		if ip == nil {
			return "", fmt.Errorf("Error getting IP from HTTP path: %s", err)
		}
		network = &net.IPNet{}
		if ip.To4() != nil {
			network.Mask = net.CIDRMask(32, 32)
		} else if ip.To16() != nil {
			network.Mask = net.CIDRMask(128, 128)
		}
	}
	network.IP = ip
	return network.String(), nil
}
