package tigerblood

import (
	"github.com/gorilla/mux"
	"net/http"
)

func NewRouter() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)

	for _, route := range routes {
		var handler http.Handler

		handler = route.HandlerFunc

		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler)
	}

	return router
}

type Route struct {
	Name         string
	Method       string
	Pattern      string
	HandlerFunc  http.HandlerFunc
}

type Routes []Route

var UnauthedRoutes = map[string]bool{
	"/__lbheartbeat__": true,
	"/__heartbeat__": true,
	"/__version__": true,
}

var routes = Routes{
	Route{
		"LoadBalancerHeartbeat",
		"GET",
		"/__lbheartbeat__",
		LoadBalancerHeartbeatHandler,
	},
	Route{
		"Heartbeat",
		"GET",
		"/__heartbeat__",
		HeartbeatHandler,
	},
	Route{
		"Version",
		"GET",
		"/__version__",
		VersionHandler,
	},
	Route{
		"ListViolations",
		"GET",
		"/violations",
		ListViolationsHandler,
	},
	Route{
		"UpsertReputationByViolation",
		"PUT",
		"/violations/{type:[[:punct:]\\w]{1,255}}",  // include all :punct: since gorilla/mux barfed trying to limit it to `:` (or as \x3a)
		UpsertReputationByViolationHandler,
	},
	Route{
		"ReadReputation",
		"GET",
		"/{ip:[[:punct:]\\/\\.\\w]{1,128}}", // see above note for all punct for IPs w/ colons e.g. 2001:db8::/32
		ReadReputationHandler,
	},
	Route{
		"CreateReputation",
		"POST",
		"/",
		CreateReputationHandler,
	},
	Route{
		"UpdateReputation",
		"PUT",
		"/{ip:[[:punct:]\\/\\.\\w]{1,128}}",
		UpdateReputationHandler,
	},
	Route{
		"DeleteReputation",
		"DELETE",
		"/{ip:[[:punct:]\\/\\.\\w]{1,128}}",
		DeleteReputationHandler,
	},
}
