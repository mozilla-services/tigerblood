package tigerblood

import (
	log "github.com/Sirupsen/logrus"
	"go.mozilla.org/mozlogrus"
	"github.com/DataDog/datadog-go/statsd"
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

func AddDB(db *DB) Middleware {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, addtoContext(r, ctxDBKey, db))
		})
	}
}

func AddStatsdClient(statsdClient *statsd.Client) Middleware {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, addtoContext(r, ctxStatsdKey, statsdClient))
		})
	}
}

func AddViolations(violationPenalties map[string]uint) Middleware {
	for violationType, penalty := range violationPenalties {
		if !(IsValidViolationName(violationType) && IsValidViolationPenalty(uint64(penalty))) {
			delete(violationPenalties, violationType)
			if !IsValidViolationName(violationType) {
				log.Printf("Skipping invalid violation type: %s", violationType)
			}
			if !IsValidViolationPenalty(uint64(penalty)) {
				log.Printf("Skipping invalid violation penalty: %s", penalty)
			}
		}
	}

	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, addtoContext(r, ctxPenaltiesKey, violationPenalties))
		})
	}
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

			var statsdClient *statsd.Client = nil
			val = r.Context().Value(ctxStatsdKey)
			if val != nil {
				statsdClient = val.(*statsd.Client)
			}

			if statsdClient != nil {
				statsdClient.Histogram("request.duration", float64(time.Since(startTime).Nanoseconds())/float64(1e6), nil, 1)
			}
			if time.Since(startTime).Nanoseconds() > int64(slowRequestCutoff) {
				log.Printf("Request took %s to process\n", time.Since(startTime))
			}

			h.ServeHTTP(w, r)
		})
	}
}
