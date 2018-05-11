package tigerblood

import (
	"net/http"
)

// Middleware wraps an http.Handler with additional functionality.
type Middleware func(http.Handler) http.Handler

// HandleWithMiddleware runs a request through a stack of middleware then the http.Handler
func HandleWithMiddleware(h http.Handler, adapters []Middleware) http.Handler {
	// To make the middleware run in the order in which they are specified,
	// we reverse through them in the Middleware function, rather than just
	// ranging over them
	for i := len(adapters) - 1; i >= 0; i-- {
		h = adapters[i](h)
	}
	return h
}

type responseHeader struct {
	Field string
	Value string
}

// DefaultResponseHeaders is an array of default HTTP headers to return
var DefaultResponseHeaders = []responseHeader{
	responseHeader{"Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'; report-uri /__cspreport__"}, // APP-CSP
	responseHeader{"Content-Type", "application/json"},                                                                 // APP-NOHTML
	responseHeader{"X-Frame-Options", "DENY"},
	responseHeader{"X-Content-Type-Options", "nosniff"},
}

// SetResponseHeaders is middleware that adds default security and content type headers
func SetResponseHeaders() Middleware {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for _, header := range DefaultResponseHeaders {
				w.Header().Add(header.Field, header.Value)
			}
			h.ServeHTTP(w, r)
		})
	}
}
