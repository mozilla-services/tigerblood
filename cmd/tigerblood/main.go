package main

import (
	"github.com/DataDog/datadog-go/statsd"
	"github.com/spf13/viper"
	"go.mozilla.org/tigerblood"
	"log"
	"net/http"
)

func main() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.SetDefault("DATABASE_MAX_OPEN_CONNS", 80)
	viper.SetDefault("BIND_ADDR", "127.0.0.1:8080")
	viper.SetDefault("STATSD_ADDR", "127.0.0.1:8125")
	viper.SetDefault("HAWK", false)
	viper.SetEnvPrefix("tigerblood")
	viper.AutomaticEnv()
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Error loading config file: %s", err)
	}
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
	} else {
		log.Println("statsd not found")
	}

	var handler http.Handler = tigerblood.NewTigerbloodHandler(db, statsdClient)
	if viper.GetBool("HAWK") {
		credentials := viper.GetStringMapString("CREDENTIALS")
		if len(credentials) == 0 {
			log.Fatal("Hawk was enabled, but no credentials were found.")
		}
		handler = tigerblood.NewHawkHandler(handler, credentials)
	}
	http.HandleFunc("/", handler.ServeHTTP)
	err = http.ListenAndServe(viper.GetString("BIND_ADDR"), nil)
	if err != nil {
		log.Fatal(err)
	}
}
