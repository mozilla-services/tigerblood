package tigerblood

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"time"
)

var ipre = regexp.MustCompile(`/((((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\.|/(3[0-2]|(2|1)?[0-9])$|$))){4})`)

// IPAddressFromHTTPPath takes a HTTP path and returns an IPv4 IP if it's found, or an error if none is found.
func IPAddressFromHTTPPath(path string) (string, error) {
	match := ipre.FindStringSubmatch(path)
	if match == nil {
		return "", fmt.Errorf("No IPv4 address found on input string %s", path)
	}
	return match[1], nil
}

// Handler is the main HTTP handler for tigerblood.
func Handler(w http.ResponseWriter, r *http.Request, db *DB) {
	startTime := time.Now()
	switch r.Method {
	case "GET":
		ReadReputation(w, r, db)
	case "POST":
	case "PUT":
	case "DELETE":
	default:
	}
	if time.Since(startTime).Nanoseconds() > 1e7 {
		log.Printf("Request took %s to proces\n", time.Since(startTime))
	}
}

// CreateReputation takes a JSON formatted IP reputation entry from the http request and inserts it to the database.
func CreateReputation(w http.ResponseWriter, r *http.Request) {
}

// ReadReputation returns a JSON-formatted reputation entry from the database.
func ReadReputation(w http.ResponseWriter, r *http.Request, db *DB) {
	ip, err := IPAddressFromHTTPPath(r.URL.Path)
	if err != nil {
		// This means there was no IP address found in the path
		w.WriteHeader(http.StatusBadRequest)
		log.Println(err)
		return
	}
	entry, err := db.SelectSmallestMatchingSubnet(ip)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Panic(err)
	}
	json, err := json.Marshal(entry)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
	}
	w.Write(json)
}

// UpdateReputation takes a JSON body from the http request and updates that reputation entry on the database.
func UpdateReputation(w http.ResponseWriter, r *http.Request) {

}

// DeleteReputation deletes
func DeleteReputation(w http.ResponseWriter, r *http.Request) {

}
