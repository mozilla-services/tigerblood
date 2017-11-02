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
	log.Printf("Unauthed routes: %s", UnauthedRoutes)

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
		if !(IsValidViolationName(violationType) && IsValidViolationPenalty(uint64(penalty))) {
			delete(newPenalties, violationType)
			if !IsValidViolationName(violationType) {
				log.Printf("Skipping invalid violation type: %s", violationType)
			}
			if !IsValidViolationPenalty(uint64(penalty)) {
				log.Printf("Skipping invalid violation penalty: %s", penalty)
			}
		}
	}

	violationPenalties = newPenalties

	// set violationPenaltiesJSON
	json, err := json.Marshal(violationPenalties)
	if err != nil {
		log.WithFields(log.Fields{"errno": JSONMarshalError}).Warnf(DescribeErrno(JSONMarshalError), "violations", err)
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
