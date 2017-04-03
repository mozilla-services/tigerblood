package tigerblood

import (
	"net/http"
	"net/http/pprof"
)

func NewRouter() *http.ServeMux {
	router := http.NewServeMux()

	if useProfileHandlers {
		// Register pprof handlers
		router.HandleFunc("/debug/", pprof.Index)
		router.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		router.HandleFunc("/debug/pprof/profile", pprof.Profile)
		router.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	}

	router.HandleFunc("/__version__", VersionHandler)
	router.HandleFunc("/__heartbeat__", HeartbeatHandler)
	router.HandleFunc("/__lbheartbeat__", LoadBalancerHeartbeatHandler)

	router.HandleFunc("/violations/", UpsertReputationByViolationHandler)
	router.HandleFunc("/violations", ListViolationsHandler)
	router.HandleFunc("/", ReputationHandler)

	return router
}

var UnauthedRoutes = map[string]bool{
	"/__lbheartbeat__": true,
	"/__heartbeat__": true,
	"/__version__": true,
}

var UnauthedDebugRoutes = map[string]bool{
	"/debug/pprof/": true,
	"/debug/pprof/heap": true,
	"/debug/pprof/block": true,
	"/debug/pprof/cmdline": true,
	"/debug/pprof/profile": true,
	"/debug/pprof/symbol": true,
	"/debug/pprof/goroutine": true,
}
