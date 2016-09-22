package main

import (
	"fmt"
	"github.com/DataDog/datadog-go/statsd"
	"go.mozilla.org/tigerblood"
	"log"
	"net/http"
	"os"
)

func main() {
	dsn, found := os.LookupEnv("TIGERBLOOD_DSN")
	if !found {
		log.Println("No database DSN found")
	}
	db, err := tigerblood.NewDB(dsn)
	if err != nil {
		panic(fmt.Errorf("Could not connect to the database: %s", err))
	}
	db.SetMaxOpenConns(80)
	var statsdClient *statsd.Client
	if statsdAddr, found := os.LookupEnv("TIGERBLOOD_STATSD_ADDR"); found {
		statsdClient, err = statsd.New(statsdAddr)
		statsdClient.Namespace = "tigerblood."
	} else if !found || err != nil {
		log.Println("statsd not found")
	}
	var handler http.Handler = tigerblood.NewTigerbloodHandler(db, statsdClient)
	if _, found := os.LookupEnv("TIGERBLOOD_NO_HAWK"); !found {
		handler = tigerblood.NewHawkHandler(handler, nil)
	}
	http.HandleFunc("/", handler.ServeHTTP)
	bind, found := os.LookupEnv("TIGERBLOOD_BIND_ADDR")
	if !found {
		bind = "127.0.0.1:8080"
	}
	err = http.ListenAndServe(bind, nil)
	if err != nil {
		panic(err)
	}
}
