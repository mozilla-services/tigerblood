package tigerblood

import (
	"net/http"
)

// Middleware wraps an http.Handler with additional functionality.
type Middleware func(http.Handler) http.Handler

// Run the request through all middleware
func HandleWithMiddleware(h http.Handler, adapters []Middleware) http.Handler {
	// To make the middleware run in the order in which they are specified,
	// we reverse through them in the Middleware function, rather than just
	// ranging over them
	for i := len(adapters) - 1; i >= 0; i-- {
		h = adapters[i](h)
	}
	return h
}

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

func SetResponseHeaders(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, header := range DefaultResponseHeaders {
			w.Header().Add(header.Field, header.Value)
		}
		h.ServeHTTP(w, r)
	})
}
