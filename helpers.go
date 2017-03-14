package tigerblood

import (
	"fmt"
	"net"
	"net/http"
	"io"
	"io/ioutil"
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

func DumpBody(req *http.Request) {
	// when running behind nginx connection reset by peer issues arise
	// in issue https://github.com/golang/go/issues/15789 it could be that
	// nginx requires the whole request to be read before a response can be generated
	if req.Body != nil {
		io.Copy(ioutil.Discard, req.Body)
		req.Body.Close()
	}
}
