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
	log.Printf("loaded middleware")
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
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, addtoContext(r, ctxPenaltiesKey, violationPenalties))
		})
	}
}

func SetResponseHeaders() Middleware {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Strict-Transport-Security", "max-age=31536000;") // APP-HSTS
			w.Header().Add("Public-Key-Pins", `max-age=5184000; pin-sha256="WoiWRyIOVNa9ihaBciRSC7XHjliYS9VwUGOIud4PB18="; pin-sha256="r/mIkG3eEpVdm+u/ko/cwxzOMo1bk4TyHIlByibiA5E="; pin-sha256="YLh1dUR9y6Kja30RrAn7JKnbQG/uEtLMkBgFF2Fuihg="; pin-sha256="sRHdihwgkaib1P1gxX8HFszlD+7/gTfNvuAybgLPNis=";`) // APP-HPKP
			w.Header().Add("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'; report-uri /__cspreport__")  // APP-CSP
			w.Header().Add("Content-Type", "application/json") // APP-NOHTML

			w.Header().Add("X-Frame-Options", "DENY")
			w.Header().Add("X-Content-Type-Options", "nosniff")

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

func LogRequestDuration() Middleware {
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
			if time.Since(startTime).Nanoseconds() > 1e7 {
				log.Printf("Request took %s to process\n", time.Since(startTime))
			}

			h.ServeHTTP(w, r)
		})
	}
}
