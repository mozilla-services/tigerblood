package tigerblood

import (
	"encoding/json"
	"github.com/DataDog/datadog-go/statsd"
	log "github.com/sirupsen/logrus"
	"go.mozilla.org/mozlogrus"
	"runtime"
)

var db *DB
var statsdClient *statsd.Client
var violationPenalties map[string]uint
var violationPenaltiesJSON []byte
var useProfileHandlers = false
var maxEntries = int(100)
var exceptionSources []exceptionSource

func init() {
	mozlogrus.Enable("tigerblood")
}

// SetDB sets or updates the db handle
func SetDB(newDb *DB) {
	db = newDb
}

// SetProfileHandlers enables or disables runtime profile handlers
func SetProfileHandlers(profileHandlers bool) {
	useProfileHandlers = profileHandlers

	for route := range UnauthedDebugRoutes {
		UnauthedRoutes[route] = useProfileHandlers
	}
	var urs string
	for x := range UnauthedRoutes {
		if urs != "" {
			urs += ", "
		}
		urs += x
	}
	log.Printf("Unauthed routes: %s", urs)

	if profileHandlers {
		runtime.SetMutexProfileFraction(5)
		runtime.SetBlockProfileRate(1)
	} else {
		runtime.SetMutexProfileFraction(0)
		runtime.SetBlockProfileRate(0)
	}
}

// SetStatsdClient sets or updates the statsd client for handlers
func SetStatsdClient(newClient *statsd.Client) {
	statsdClient = newClient
}

// SetViolationPenalties sets or updates the violation penalties map
func SetViolationPenalties(newPenalties map[string]uint) {
	for violationType, penalty := range newPenalties {
		if !IsValidViolationName(violationType) {
			log.Fatalf("Invalid violation type: %s", violationType)
		}
		if !IsValidViolationPenalty(penalty) {
			log.Fatalf("Invalid violation penalty: %d", penalty)
		}
	}
	violationPenalties = newPenalties

	// set violationPenaltiesJSON
	json, err := json.Marshal(violationPenalties)
	if err != nil {
		log.WithFields(log.Fields{"errno": JSONMarshalError}).Fatalf(DescribeErrno(JSONMarshalError),
			"violations", err)
	}
	violationPenaltiesJSON = json
}

// SetMaxEntries updates the maximum number of entries in multi entry handlers
func SetMaxEntries(newMaxEntries int) {
	if newMaxEntries < 0 {
		log.Fatal("MAX_ENTRIES must be positive")
	}
	log.Debugf("Setting max entries: %s", newMaxEntries)
	maxEntries = newMaxEntries
}
