package main

import (
	"flag"
	"fmt"
	"github.com/DataDog/datadog-go/statsd"
	"go.mozilla.org/tigerblood"
	"log"
	"net/http"
)

func main() {
	configFile := flag.String("config-file", "config.yml", "Path to the YAML config file")
	flag.Parse()
	config, err := tigerblood.LoadConfigFromPath(*configFile)
	if err != nil {
		log.Println("Error loading config file:", err)
	}
	if config.DatabaseDsn == "" {
		log.Fatal("No database DSN found")
	}
	db, err := tigerblood.NewDB(config.DatabaseDsn)
	if err != nil {
		log.Fatal(fmt.Errorf("Could not connect to the database: %s", err))
	}
	db.SetMaxOpenConns(config.DatabaseMaxOpenConns)
	var statsdClient *statsd.Client
	if config.StatsdAddress != "" {
		statsdClient, err = statsd.New(config.StatsdAddress)
		statsdClient.Namespace = "tigerblood."
		if err != nil {
			log.Println("Could not connect to statsd:", err)
		}
	} 
	var handler http.Handler = tigerblood.NewTigerbloodHandler(db, statsdClient)
	if config.EnableHawk {
		if config.Credentials == nil {
			log.Fatal(fmt.Sprintf("Hawk is enabled but the Hawk credential map is nil!"))
		}
		handler = tigerblood.NewHawkHandler(handler, config.Credentials)
	}
	http.HandleFunc("/", handler.ServeHTTP)
	if config.BindAddress == "" {
		config.BindAddress = "127.0.0.1:8080"
	}
	err = http.ListenAndServe(config.BindAddress, nil)
	if err != nil {
		log.Fatal(err)
	}
}
