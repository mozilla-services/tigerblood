package main

import (
	log "github.com/Sirupsen/logrus"
	"go.mozilla.org/mozlogrus"
	"github.com/DataDog/datadog-go/statsd"
	"github.com/bmhatfield/go-runtime-metrics/collector"
	"github.com/spf13/viper"
	"go.mozilla.org/tigerblood"
	"github.com/peterbourgon/g2s"
	"fmt"
	"time"
	"strconv"
	"net/http"
)



func printConfig() {
	var fields = log.Fields{}
	for key, value := range viper.AllSettings() {
		switch key {
		case "credentials":  // skip sensitive keys
		case "dsn":
		default:
			fields[key] = value
		}
	}

	log.WithFields(fields).Info("Loaded viper config:")
}

func startRuntimeCollector() {
	statsd_addr := viper.GetString("STATSD_ADDR")
	s, err := g2s.Dial("udp", statsd_addr)
	if err != nil {
		panic(fmt.Sprintf("Unable to connect to Statsd on %s - %s", statsd_addr, err))
	}

	gaugeFunc := func(key string, val uint64) {
		s.Gauge(1.0, viper.GetString("STATSD_NAMESPACE")+key, strconv.FormatUint(val, 10))
	}
	c := collector.New(gaugeFunc)
	c.PauseDur = time.Duration(viper.GetInt("RUNTIME_PAUSE_INTERVAL")) * time.Second
	c.EnableCPU = viper.GetBool("RUNTIME_CPU")
	c.EnableMem = viper.GetBool("RUNTIME_MEM")
	c.EnableGC = viper.GetBool("RUNTIME_GC")
	c.Run()
}

func main() {
	mozlogrus.Enable("tigerblood")

	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.SetDefault("DATABASE_MAX_OPEN_CONNS", 80)
	viper.SetDefault("BIND_ADDR", "127.0.0.1:8080")
	viper.SetDefault("STATSD_ADDR", "127.0.0.1:8125")
	viper.SetDefault("STATSD_NAMESPACE", "tigerblood.")
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
		statsdClient.Namespace = viper.GetString("STATSD_NAMESPACE")
		if viper.GetBool("PUBLISH_RUNTIME_STATS") {
			go startRuntimeCollector()
		}
	} else {
		log.Println("statsd not found")
	}

	if !viper.IsSet("VIOLATION_PENALTIES") {
		log.Fatal("No violation penalties found.")
	}

	var penalties = make(map[string]uint)
	for k, penalty := range viper.GetStringMapString("VIOLATION_PENALTIES") {
		penalty, err := strconv.ParseUint(penalty, 10, 64)
		if err != nil {
			log.Printf("Error loading violation weight %s: %s", penalty, err)
		} else {
			penalties[k] = uint(penalty)
		}
	}
	log.Printf("loaded violation map: %s", penalties)
	var handler http.Handler = tigerblood.NewTigerbloodHandler(db, statsdClient, penalties)

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
	log.Printf("Listening on %s", viper.GetString("BIND_ADDR"))
	err = http.ListenAndServe(viper.GetString("BIND_ADDR"), nil)
	if err != nil {
		log.Fatal(err)
	}
}
