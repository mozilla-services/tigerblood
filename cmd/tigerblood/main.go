package main

import (
	"flag"
	"fmt"
	"github.com/DataDog/datadog-go/statsd"
	"go.mozilla.org/tigerblood"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

func loadConfig(path string) (tigerblood.Config, error) {
	var config tigerblood.Config
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return config, err
	}
	err = yaml.Unmarshal(bytes, &config)
	return config, err
}

func main() {
	configFile := flag.String("config-file", "config.yml", "Path to the YAML config file")
	flag.Parse()
	config, err := loadConfig(*configFile)
	if err != nil {
		log.Println("Error loading config file:", err)
	}
	dsn, found := os.LookupEnv("TIGERBLOOD_DSN")
	if !found {
		log.Println("No database DSN found")
	}
	db, err := tigerblood.NewDB(dsn)
	if err != nil {
		log.Fatal(fmt.Errorf("Could not connect to the database: %s", err))
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
		if config.Credentials == nil {
			log.Fatal(fmt.Sprintf("Could not load hawk credentials! %s", err))
		}
		handler = tigerblood.NewHawkHandler(handler, config.Credentials)
	}
	http.HandleFunc("/", handler.ServeHTTP)
	bind, found := os.LookupEnv("TIGERBLOOD_BIND_ADDR")
	if !found {
		bind = "127.0.0.1:8080"
	}
	err = http.ListenAndServe(bind, nil)
	if err != nil {
		log.Fatal(err)
	}
}
