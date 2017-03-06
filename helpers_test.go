package tigerblood

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var cases = []struct {
	path string
	ip   string
	err  bool
}{
	{
		"/192.168.0.1",
		"192.168.0.1/32",
		false,
	},
	{
		"/192.168.0.1/32",
		"192.168.0.1/32",
		false,
	},
	{
		"/300.123.345.567",
		"",
		true,
	},
	{
		"/foobar",
		"",
		true,
	},
	{
		"/....",
		"",
		true,
	},
	{
		"/2001:0db8:0123:4567:89ab:cdef:1234:5678",
		"2001:db8:123:4567:89ab:cdef:1234:5678/128",
		false,
	},
	{
		"/2001:db8::ff00:42:8329",
		"2001:db8::ff00:42:8329/128",
		false,
	},
	{
		"/127.0.0.1' or '1' = '1",
		"",
		true,
	},
	{
		"/127.0.0.1; -- SELECT(2)",
		"",
		true,
	},
	{
		"/127.0.0.1/",
		"",
		true,
	},
}

func TestIPAddressFromHTTPPath(t *testing.T) {
	for _, c := range cases {
		ip, err := IPAddressFromHTTPPath(c.path)
		if c.err {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}
		assert.Equal(t, c.ip, ip)
	}
}
