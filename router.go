package tigerblood

import (
	"github.com/gorilla/mux"
	"net/http"
	"log"
)

func init() {
	log.Printf("!!!!")
}

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
		"/violations/{type}",
		UpsertReputationByViolationHandler,
	},
	Route{
		"ReadReputation",
		"GET",
		"/{ip}",
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
		"/{ip}",
		UpdateReputationHandler,
	},
	Route{
		"DeleteReputation",
		"DELETE",
		"/{ip}",
		DeleteReputationHandler,
	},
}
