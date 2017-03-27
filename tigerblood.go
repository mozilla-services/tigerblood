package tigerblood

import (
	"github.com/DataDog/datadog-go/statsd"
)

var db *DB = nil
var statsdClient *statsd.Client = nil

func SetDB(newDb *DB) {
	db = newDb
}

func SetStatsdClient(newClient *statsd.Client) {
	statsdClient = newClient
}
