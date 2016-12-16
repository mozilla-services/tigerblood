package main

import (
	"github.com/DataDog/datadog-go/statsd"
	_ "github.com/bmhatfield/go-runtime-metrics"
	"github.com/spf13/viper"
	"flag"
	"go.mozilla.org/tigerblood"
	"log"
	"net/http"
)



func printConfig() {
	log.Println("Loaded viper config:")
	for key, value := range viper.AllSettings() {
		if key != "credentials" {
			log.Print("\t", key, ": ", value)
		}
	}
}

func main() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.SetDefault("DATABASE_MAX_OPEN_CONNS", 80)
	viper.SetDefault("BIND_ADDR", "127.0.0.1:8080")
	viper.SetDefault("STATSD_ADDR", "127.0.0.1:8125")
	viper.SetDefault("HAWK", false)
	viper.SetDefault("PUBLISH_RUNTIME_STATS", false)
	viper.SetDefault("RUNTIME_PAUSE_INTERVAL", 10)
	viper.SetDefault("RUNTIME_CPU", true)
	viper.SetDefault("RUNTIME_MEM", true)
	viper.SetDefault("RUNTIME_GC", true)

	viper.SetEnvPrefix("tigerblood")
	viper.AutomaticEnv()
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Error loading config file: %s", err)
	}

	printConfig()

	// Set flags for go-runtime-metrics
	flag.Set("statsd", viper.GetString("STATSD_ADDR"))
	flag.Set("metric-prefix", "tigerblood")
	flag.Set("pause", viper.GetString("RUNTIME_PAUSE_INTERVAL"))
	flag.Set("publish-runtime-stats", viper.GetString("PUBLISH_RUNTIME_STATS"))
	flag.Set("cpu", viper.GetString("RUNTIME_CPU"))
	flag.Set("mem", viper.GetString("RUNTIME_MEM"))
	flag.Set("gc", viper.GetString("RUNTIME_GC"))

	if !viper.IsSet("DSN") {
		log.Fatalf("No DSN found. Cannot continue without a database")
	}
	db, err := tigerblood.NewDB(viper.GetString("DSN"))
	if err != nil {
		log.Fatalf("Could not connect to database: %s", err)
	}
	db.SetMaxOpenConns(viper.GetInt("DATABASE_MAX_OPEN_CONNS"))

	var statsdClient *statsd.Client
	if viper.IsSet("STATSD_ADDR") {
		statsdClient, err = statsd.New(viper.GetString("STATSD_ADDR"))
		statsdClient.Namespace = "tigerblood."
		flag.Parse() // kick off go-runtime-stats collector
	} else {
		log.Println("statsd not found")
	}

	var handler http.Handler = tigerblood.NewTigerbloodHandler(db, statsdClient)
	if viper.GetBool("HAWK") {
		credentials := viper.GetStringMapString("CREDENTIALS")
		if len(credentials) == 0 {
			log.Fatal("Hawk was enabled, but no credentials were found.")
		} else {
			log.Printf("Hawk enabled with %d credentials.", len(credentials))
		}
		handler = tigerblood.NewHawkHandler(handler, credentials)
	}
	http.HandleFunc("/", handler.ServeHTTP)
	err = http.ListenAndServe(viper.GetString("BIND_ADDR"), nil)
	if err != nil {
		log.Fatal(err)
	}
}
