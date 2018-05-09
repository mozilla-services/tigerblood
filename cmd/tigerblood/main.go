package main

import (
	"fmt"
	"github.com/DataDog/datadog-go/statsd"
	"github.com/bmhatfield/go-runtime-metrics/collector"
	"github.com/peterbourgon/g2s"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"go.mozilla.org/mozlogrus"
	"go.mozilla.org/tigerblood"
	"net/http"
	_ "net/http/pprof"
	"strconv"
	"strings"
	"time"
)

func printConfig() {
	var fields = log.Fields{}
	for key, value := range viper.AllSettings() {
		switch key {
		case "credentials": // skip sensitive keys
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

func loadConfig() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.SetDefault("DATABASE_MAX_OPEN_CONNS", 75)
	viper.SetDefault("DATABASE_MAX_IDLE_CONNS", 75)
	viper.SetDefault("DATABASE_MAXLIFETIME", "0")
	viper.SetDefault("BIND_ADDR", "127.0.0.1:8080")
	viper.SetDefault("STATSD_ADDR", "127.0.0.1:8125")
	viper.SetDefault("STATSD_NAMESPACE", "tigerblood.")
	viper.SetDefault("HAWK", false)
	viper.SetDefault("APIKEY", false)
	viper.SetDefault("PUBLISH_RUNTIME_STATS", false)
	viper.SetDefault("RUNTIME_PAUSE_INTERVAL", 10)
	viper.SetDefault("RUNTIME_CPU", true)
	viper.SetDefault("RUNTIME_MEM", true)
	viper.SetDefault("RUNTIME_GC", true)
	viper.SetDefault("PROFILE", false)
	viper.SetDefault("MAX_ENTRIES", 1000)

	viper.SetEnvPrefix("tigerblood")
	viper.AutomaticEnv()
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Error loading config file: %s", err)
	}
}

func loadHawkCredentials() map[string]string {
	credentials := viper.GetStringMapString("HAWK_CREDENTIALS")
	if len(credentials) == 0 {
		log.Fatal("Hawk was enabled, but no credentials were found.")
	} else {
		log.Printf("Hawk enabled with %d credentials.", len(credentials))
	}
	return credentials
}

func loadAPIKeyCredentials() map[string]string {
	credentials := viper.GetStringMapString("APIKEY_CREDENTIALS")
	if len(credentials) == 0 {
		log.Fatal("API key authentication was enabled, but no credentials were found.")
	} else {
		log.Printf("API key authentication enabled with %d credentials.", len(credentials))
	}
	return credentials
}

func loadDB() *tigerblood.DB {
	if !viper.IsSet("DSN") {
		log.Fatalf("No DSN found. Cannot continue without a database")
	}
	db, err := tigerblood.NewDB(viper.GetString("DSN"))
	if err != nil {
		log.Fatalf("Could not connect to database: %s", err)
	}

	db.SetMaxOpenConns(viper.GetInt("DATABASE_MAX_OPEN_CONNS"))
	db.SetMaxIdleConns(viper.GetInt("DATABASE_MAX_IDLE_CONNS"))

	if viper.GetString("DATABASE_MAXLIFETIME") == "0" {
		db.SetConnMaxLifetime(0)
	} else {
		lifetime, err := time.ParseDuration(viper.GetString("DATABASE_MAXLIFETIME"))
		if err != nil {
			db.SetConnMaxLifetime(lifetime)
		} else {
			log.Warnf("Error parsing conn db max lifetime: %s", err)
		}
	}
	return db
}

func loadStatsd() *statsd.Client {
	statsdClient, err := statsd.New(viper.GetString("STATSD_ADDR"))
	if err != nil {
		log.Fatalf("Error loading statsdClient: %s", err)
	}
	statsdClient.Namespace = viper.GetString("STATSD_NAMESPACE")
	if viper.GetBool("PUBLISH_RUNTIME_STATS") {
		go startRuntimeCollector()
	}
	return statsdClient
}

func loadViolationPenalties() map[string]uint {
	if !viper.IsSet("VIOLATION_PENALTIES") {
		log.Fatal("No violation penalties found.")
	}

	// pass as violation_type=penalty (e.g. rateLimited=20) to
	// workaround for viper lowercasing everything
	// https://github.com/spf13/viper/issues/260
	var penalties = make(map[string]uint)
	for _, kv := range strings.Split(viper.GetString("VIOLATION_PENALTIES"), ",") {
		tmp := strings.Split(kv, "=")
		if len(tmp) != 2 {
			log.Printf("Error loading violation penalty %s (format should be type=penalty)", tmp)
			continue
		}
		violationType, penalty := tmp[0], tmp[1]
		parsedPenalty, err := strconv.ParseUint(penalty, 10, 64)
		if err != nil {
			log.Printf("Error parsing violation weight %s: %s", penalty, err)
			continue
		}
		if !tigerblood.IsValidViolationName(violationType) {
			log.Printf("Skipping invalid violation type: %s", violationType)
			continue
		}
		if !tigerblood.IsValidViolationPenalty(parsedPenalty) {
			log.Printf("Skipping invalid violation penalty: %s: %d", violationType, parsedPenalty)
			continue
		}
		penalties[violationType] = uint(parsedPenalty)
	}
	log.Printf("loaded violation map: %s", penalties)

	return penalties
}

func loadExceptions() {
	if !viper.IsSet("EXCEPTIONS") {
		return
	}

	for _, kv := range strings.Split(viper.GetString("EXCEPTIONS"), ",") {
		var ed, ec string
		tmp := strings.Split(kv, "=")
		if len(tmp) == 0 {
			continue
		}
		ed = tmp[0]
		if len(tmp) > 1 {
			ec = tmp[1]
		}
		switch ed {
		case "file":
			// Configuration is just the path to the file containing the
			// address list
			log.Printf("Adding exception source file %s", ec)
			err := tigerblood.AddFileException(ec)
			if err != nil {
				log.Fatalf("Error adding file exception: %s", err)
			}
		case "aws":
			// No configuration
			log.Print("Adding exception source AWS public address data")
			err := tigerblood.AddAWSException()
			if err != nil {
				log.Fatalf("Error adding AWS exception: %s", err)
			}
		default:
			log.Fatalf("Invalid exception source type %s", ed)
		}
	}
}

func main() {
	mozlogrus.Enable("tigerblood")
	loadConfig()
	printConfig()

	var (
		middleware []tigerblood.Middleware
		authmask   int
	)

	if viper.GetBool("HAWK") {
		tigerblood.SetHawkCredentials(loadHawkCredentials())
		authmask |= tigerblood.AuthEnableHawk
	}
	if viper.GetBool("APIKEY") {
		tigerblood.SetAPIKeyCredentials(loadAPIKeyCredentials())
		authmask |= tigerblood.AuthEnableAPIKey
	}
	middleware = append(middleware, tigerblood.RequireAuth())
	tigerblood.SetAuthMask(authmask)

	tigerblood.SetProfileHandlers(viper.GetBool("PROFILE"))

	tigerblood.SetDB(loadDB())

	loadExceptions()
	err := tigerblood.InitializeExceptions()
	if err != nil {
		log.Fatalf("Error initializing exception sources: %s", err)
	}

	if viper.IsSet("STATSD_ADDR") {
		tigerblood.SetStatsdClient(loadStatsd())
	} else {
		log.Println("statsd not found")
	}

	tigerblood.SetViolationPenalties(loadViolationPenalties())
	tigerblood.SetMaxEntries(viper.GetInt("MAX_ENTRIES"))

	middleware = append(middleware, tigerblood.SetResponseHeaders())

	if authmask == 0 {
		log.Warn("Warning, authentication is disabled")
	}
	log.Printf("Listening on %s", viper.GetString("BIND_ADDR"))
	err = http.ListenAndServe(
		viper.GetString("BIND_ADDR"),
		tigerblood.HandleWithMiddleware(tigerblood.NewRouter(), middleware))
	log.Fatal(err)
}
