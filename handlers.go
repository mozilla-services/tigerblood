package tigerblood

import (
	log "github.com/Sirupsen/logrus"
	"go.mozilla.org/mozlogrus"
	"github.com/DataDog/datadog-go/statsd"
	"database/sql"
	"fmt"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"net"
	"os"
	"path"
	"strings"
)

// Context Keys
const (
	ctxDBKey = "db"
	ctxPenaltiesKey = "violationPenalties"
	ctxStatsdKey = "statsd"
	ctxStartTimeKey = "startTime"
)

func init() {
	mozlogrus.Enable("tigerblood")
}

func LoadBalancerHeartbeatHandler(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	return
}

func HeartbeatHandler(w http.ResponseWriter, req *http.Request) {
	val := req.Context().Value(ctxDBKey)
	if val == nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Could not find database handler in request context")
		return
	}
	db := val.(*DB)

	err := db.Ping()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

func VersionHandler(w http.ResponseWriter, req *http.Request) {
	dir, err := os.Getwd()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Could not get CWD")
		return
	}
	filename := path.Clean(dir + string(os.PathSeparator) + "version.json")
	f, err := os.Open(filename)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	stat, err := f.Stat()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	http.ServeContent(w, req, "__version__", stat.ModTime(), f)
}

// Returns a list of known violations for debugging
func ListViolationsHandler(w http.ResponseWriter, req *http.Request) {
	val := req.Context().Value(ctxPenaltiesKey)
	if val == nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Could not find violation penalties in request context")
		return
	}
	violationPenalties := val.(map[string]uint)

	json, err := json.Marshal(violationPenalties)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Error marshaling violations to JSON: %s", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(json)
}

// UpsertReputationByViolation takes a JSON body from the http request
// and upserts the reputation entry on the database to the reputation
// given in reputation violation.  The HTTP requests path has to
// contain the IP to be updated, in CIDR notation. For example:
// {"Violation": "password-reset-rate-limit-exceeded"}
func UpsertReputationByViolationHandler(w http.ResponseWriter, r *http.Request) {
	splitPath := strings.Split(r.URL.Path, "/")
	if len(splitPath) != 3 {
		log.Printf("Path format is invalid: %s", r.URL.Path)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	ip, err := IPAddressFromHTTPPath("/" + splitPath[2])

	if err != nil {
		// This means there was no IP address found in the path
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("No IP address found in path %s: %s", r.URL.Path, err)
		return
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Error reading body: %s", err)
		return
	}
	type ViolationBody struct {
		Violation string
	}
	var entry ViolationBody
	err = json.Unmarshal(body, &entry)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Could not unmarshal request body: %s", err)
		return
	}

	val := r.Context().Value(ctxPenaltiesKey)
	if val == nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Could not find violation penalties in request context")
		return
	}
	violationPenalties := val.(map[string]uint)

	// lookup violation weight in config map
	var penalty, ok = violationPenalties[entry.Violation]
	if !ok {
		log.Printf("Could not find violation type: %s", entry.Violation)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Violation type not found."))
		return
	}

	val = r.Context().Value(ctxDBKey)
	if val == nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Could not find database handler in request context")
		return
	}
	db := val.(*DB)

	err = db.InsertOrUpdateReputationPenalty(nil, ip, uint(penalty))
	if _, ok := err.(CheckViolationError); ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Reputation is outside of valid range [0-100]"))
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Could not update reputation entry by violation: %s", err)
	} else if err == nil {
		w.WriteHeader(http.StatusNoContent)
	}
}

// CreateReputation takes a JSON formatted IP reputation entry from
// the http request and inserts it to the database.
func CreateReputationHandler(w http.ResponseWriter, r *http.Request) {
	var entry ReputationEntry
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Error reading body: %s", err)
		return
	}
	err = json.Unmarshal(body, &entry)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Could not unmarshal request body: %s", err)
		return
	}

	if net.ParseIP(entry.IP) == nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Error parsing invalid IP from HTTP body: %s", body)
		return
	}

	val := r.Context().Value(ctxDBKey)
	if val == nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Could not find database handler in request context")
		return
	}
	db := val.(*DB)

	err = db.InsertReputationEntry(nil, entry)
	if _, ok := err.(CheckViolationError); ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Reputation is outside of valid range [0-100]"))
	} else if _, ok := err.(DuplicateKeyError); ok {
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte("Reputation is already set for that IP."))
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	if err != nil {
		log.Printf("Could not insert reputation entry: %s", err)
		return
	}
	w.WriteHeader(http.StatusCreated)
}


// UpdateReputation takes a JSON body from the http request and updates that reputation entry on the database.
// The HTTP requests path has to contain the IP to be updated, in CIDR notation. The body can contain the IP address, or it can be omitted. For example:
// {"Reputation": 50} or {"Reputation": 50, "IP":, "192.168.0.1"}. The IP in the JSON body will be ignored.
func UpdateReputationHandler(w http.ResponseWriter, r *http.Request) {
	ip, err := IPAddressFromHTTPPath(r.URL.Path)
	if err != nil {
		// This means there was no IP address found in the path
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("No IP address found in path %s: %s", r.URL.Path, err)
		return
	}
	var entry ReputationEntry
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Error reading body: %s", err)
		return
	}
	err = json.Unmarshal(body, &entry)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Could not unmarshal request body: %s", err)
		return
	}
	entry.IP = ip

	val := r.Context().Value(ctxDBKey)
	if val == nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Could not find database handler in request context")
		return
	}
	db := val.(*DB)

	err = db.UpdateReputationEntry(nil, entry)
	if _, ok := err.(CheckViolationError); ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Reputation is outside of valid range [0-100]"))
	} else if err == ErrNoRowsAffected {
		w.WriteHeader(http.StatusNotFound)
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Could not update reputation entry: %s", err)
	} else if err == nil {
		w.WriteHeader(http.StatusOK)
	}
}

// DeleteReputation deletes an entry based on the IP address provided on the path
func DeleteReputationHandler(w http.ResponseWriter, r *http.Request) {
	ip, err := IPAddressFromHTTPPath(r.URL.Path)
	if err != nil {
		// This means there was no IP address found in the path
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("No IP address found in path %s: %s", r.URL.Path, err)
		return
	}

	val := r.Context().Value(ctxDBKey)
	if val == nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Could not find database handler in request context")
		return
	}
	db := val.(*DB)

	err = db.DeleteReputationEntry(nil, ReputationEntry{IP: ip})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Could not delete reputation entry: %s", err)
		return
	}
	w.WriteHeader(http.StatusOK)
}


// ReadReputation returns a JSON-formatted reputation entry from the database.
func ReadReputationHandler(w http.ResponseWriter, r *http.Request) {
	ip, err := IPAddressFromHTTPPath(r.URL.Path)
	if err != nil {
		// This means there was no IP address found in the path
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("No IP address found in path %s: %s", r.URL.Path, err)
		return
	}

	val := r.Context().Value(ctxDBKey)
	if val == nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Could not find database handler in request context")
		return
	}
	db := val.(*DB)

	var statsdClient *statsd.Client = nil
	val = r.Context().Value(ctxStatsdKey)
	if val != nil {
		statsdClient = val.(*statsd.Client)
	}

	entry, err := db.SelectSmallestMatchingSubnet(ip)
	if err == sql.ErrNoRows {
		w.WriteHeader(http.StatusNotFound)
		log.Printf("No entries found for IP %s", ip)
		if statsdClient != nil {
			statsdClient.Incr("misses", nil, 1.0)
		}
		return
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Error executing SQL: %s", err)
		return
	}
	json, err := json.Marshal(entry)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Error marshaling JSON: %s", err)
		return
	}
	if statsdClient != nil {
		statsdClient.Incr("hits", nil, 1.0)
	}
	w.WriteHeader(http.StatusOK)
	w.Write(json)
}