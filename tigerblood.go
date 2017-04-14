package tigerblood

import (
	"github.com/DataDog/datadog-go/statsd"
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"go.mozilla.org/mozlogrus"
)

var db *DB = nil
var statsdClient *statsd.Client = nil
var violationPenalties map[string]uint = nil
var violationPenaltiesJson []byte = nil
var useProfileHandlers = false

func init() {
	mozlogrus.Enable("tigerblood")
}

func SetDB(newDb *DB) {
	db = newDb
}

func SetProfileHandlers(profileHandlers bool) {
	useProfileHandlers = profileHandlers

	for route, _ := range UnauthedDebugRoutes {
		UnauthedRoutes[route] = useProfileHandlers
	}
	log.Printf("Unauthed routes: %s", UnauthedRoutes)
}

func SetStatsdClient(newClient *statsd.Client) {
	statsdClient = newClient
}

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

	// set violationPenaltiesJson
	json, err := json.Marshal(violationPenalties)
	if err != nil {
		log.WithFields(log.Fields{"errno": JSONMarshalError}).Warnf(DescribeErrno(JSONMarshalError), "violations", err)
	}
	violationPenaltiesJson = json
}
