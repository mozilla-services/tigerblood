package tigerblood

import (
	log "github.com/Sirupsen/logrus"
	"go.mozilla.org/mozlogrus"
	"net/http"
	"context"
	"time"
)

func init() {
	mozlogrus.Enable("tigerblood")
}

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

// addToContext add the given key value pair to the given request's context
func addtoContext(r *http.Request, key string, value interface{}) *http.Request {
	ctx := r.Context()
	return r.WithContext(context.WithValue(ctx, key, value))
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

func RecordStartTime() Middleware {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, addtoContext(r, ctxStartTimeKey, time.Now()))
		})
	}
}

func LogRequestDuration(slowRequestCutoff int) Middleware {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			val := r.Context().Value(ctxStartTimeKey)
			if val == nil {
				log.Printf("Could not find startTime in request context.")
				return
			}
			startTime := val.(time.Time)

			if statsdClient != nil {
				statsdClient.Histogram("request.duration", float64(time.Since(startTime).Nanoseconds())/float64(1e6), nil, 1)
			}
			if time.Since(startTime).Nanoseconds() > int64(slowRequestCutoff) {
				log.WithFields(log.Fields{
					"processing_time": time.Since(startTime).Nanoseconds(),
				}).Infof("Slow request completed successfully.")
			}

			h.ServeHTTP(w, r)
		})
	}
}
