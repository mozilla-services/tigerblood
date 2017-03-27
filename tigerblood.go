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
}

func SetStatsdClient(newClient *statsd.Client) {
	statsdClient = newClient
}

func SetViolationPenalties(newPenalties map[string]uint) {
	violationPenalties = newPenalties

	// set violationPenaltiesJson
	json, err := json.Marshal(violationPenalties)
	if err != nil {
		log.WithFields(log.Fields{"errno": JSONMarshalError}).Warnf(DescribeErrno(JSONMarshalError), "violations", err)
	}
	violationPenaltiesJson = json
}
