package tigerblood

import (
	"net/http"
)

type ResponseHeader struct {
	Field string
	Value string
}
var DefaultResponseHeaders = []ResponseHeader{
	ResponseHeader{"Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'; report-uri /__cspreport__"},  // APP-CSP
	ResponseHeader{"Content-Type", "application/json"}, // APP-NOHTML

	ResponseHeader{"X-Frame-Options", "DENY"},
	ResponseHeader{"X-Content-Type-Options", "nosniff"},
}

func SetResponseHeaders(w http.ResponseWriter) {
	for _, header := range DefaultResponseHeaders {
		w.Header().Add(header.Field, header.Value)
	}
}
