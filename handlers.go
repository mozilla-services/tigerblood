package tigerblood

import (
	"database/sql"
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"go.mozilla.org/mozlogrus"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
)

func init() {
	mozlogrus.Enable("tigerblood")
}

func LoadBalancerHeartbeatHandler(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	return
}

func HeartbeatHandler(w http.ResponseWriter, req *http.Request) {
	if db == nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.WithFields(log.Fields{"errno": MissingDB}).Warnf(DescribeErrno(MissingDB))
		return
	}

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
		log.WithFields(log.Fields{"errno": CWDNotFound}).Warnf(DescribeErrno(CWDNotFound), err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Could not get CWD")
		return
	}
	filename := path.Clean(dir + string(os.PathSeparator) + "version.json")
	f, err := os.Open(filename)
	if err != nil {
		log.WithFields(log.Fields{"errno": FileNotFound}).Warnf(DescribeErrno(FileNotFound), "version.json", err)
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
	if violationPenalties == nil || violationPenaltiesJson == nil {
		log.WithFields(log.Fields{"errno": MissingViolations}).Warnf(DescribeErrno(MissingViolations))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(violationPenaltiesJson)
}

// UpsertReputationByViolation takes a JSON body from the http request
// and upserts the reputation entry on the database to the reputation
// given in reputation violation.  The HTTP requests path has to
// contain the IP to be updated, in CIDR notation. For example:
// {"Violation": "password-reset-rate-limit-exceeded"}
func UpsertReputationByViolationHandler(w http.ResponseWriter, r *http.Request) {
	splitPath := strings.Split(r.URL.Path, "/")
	if len(splitPath) != 3 {
		log.WithFields(log.Fields{"errno": InvalidIPError}).Infof(DescribeErrno(InvalidIPError), r.URL.Path)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	ip, err := IPAddressFromHTTPPath("/" + splitPath[2])

	if err != nil {
		log.WithFields(log.Fields{"errno": MissingIPError}).Infof(DescribeErrno(MissingIPError), r.URL.Path, err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !IsValidReputationCIDROrIP(ip) {
		log.WithFields(log.Fields{"errno": InvalidIPError}).Infof(DescribeErrno(InvalidIPError), splitPath[2])
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.WithFields(log.Fields{"errno": BodyReadError}).Warnf(DescribeErrno(BodyReadError))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	type ViolationBody struct {
		Violation string
	}
	var entry ViolationBody
	err = json.Unmarshal(body, &entry)
	if err != nil {
		log.WithFields(log.Fields{"errno": JSONUnmarshalError}).Warnf(DescribeErrno(JSONUnmarshalError), err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !IsValidViolationName(entry.Violation) {
		log.WithFields(log.Fields{"errno": InvalidViolationTypeError}).Infof(DescribeErrno(InvalidViolationTypeError), entry.Violation)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if violationPenalties == nil {
		log.WithFields(log.Fields{"errno": MissingViolations}).Warnf(DescribeErrno(MissingViolations))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// lookup violation weight in config map
	var penalty, ok = violationPenalties[entry.Violation]
	if !ok {
		log.WithFields(log.Fields{"errno": MissingViolationTypeError}).Infof("Could not find violation type: %s", entry.Violation)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Violation type not found."))
		return
	}

	if db == nil {
		log.WithFields(log.Fields{"errno": MissingDB}).Warnf(DescribeErrno(MissingDB))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = db.InsertOrUpdateReputationPenalty(nil, ip, uint(penalty))
	if _, ok := err.(CheckViolationError); ok {
		log.WithFields(log.Fields{"errno": InvalidReputationError}).Warnf("Reputation is outside of valid range [0-100]")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Reputation is outside of valid range [0-100]"))
	} else if err != nil {
		log.WithFields(log.Fields{"errno": DBError}).Warnf("Could not update reputation entry by violation: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
	} else if err == nil {
		log.Debugf("Updated reputation for %s due to %d", ip, penalty)
		w.WriteHeader(http.StatusNoContent)
	}
}

func writeEntryErrorResponse(w http.ResponseWriter, errno int, entryIndex int, entry IpViolationEntry, msg string) {
	type EntryError struct {
		Errno      int
		EntryIndex int
		Entry      IpViolationEntry
		Msg        string
	}

	entryError := EntryError{
		Errno:      errno,
		EntryIndex: entryIndex,
		Entry:      entry,
		Msg:        msg,
	}
	j, err := json.Marshal(entryError)
	if err != nil {
		log.Warnf("error marshaling error response JSON: %s", err)
	}
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusBadRequest)
	w.Write(j)
	return
}

func MultiUpsertReputationByViolationHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.WithFields(log.Fields{"errno": BodyReadError}).Warnf(DescribeErrno(BodyReadError))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	var entries []IpViolationEntry
	err = json.Unmarshal(body, &entries)
	if err != nil {
		log.WithFields(log.Fields{"errno": JSONUnmarshalError}).Warnf(DescribeErrno(JSONUnmarshalError), err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if len(entries) < 1 {
		log.WithFields(log.Fields{"errno": MissingIPViolationEntryError}).Warn(DescribeErrno(MissingIPViolationEntryError))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if len(entries) > maxEntries {
		log.WithFields(log.Fields{"errno": TooManyIpViolationEntriesError}).Warn(DescribeErrno(TooManyIpViolationEntriesError))
		w.WriteHeader(http.StatusRequestEntityTooLarge)
		return
	}

	if db == nil {
		log.WithFields(log.Fields{"errno": MissingDB}).Warnf(DescribeErrno(MissingDB))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	for i, entry := range entries {
		if !IsValidViolationName(entry.Violation) {
			log.WithFields(log.Fields{"errno": InvalidViolationTypeError}).Infof(DescribeErrno(InvalidViolationTypeError), entry.Violation)
			writeEntryErrorResponse(w, InvalidViolationTypeError, i, entry, string(""))
			return
		}

		if violationPenalties == nil {
			log.WithFields(log.Fields{"errno": MissingViolations}).Warnf(DescribeErrno(MissingViolations))
			writeEntryErrorResponse(w, MissingViolations, i, entry, string(""))
			return
		}

		// lookup violation weight in config map
		var penalty, ok = violationPenalties[entry.Violation]
		if !ok {
			log.WithFields(log.Fields{"errno": MissingViolationTypeError}).Infof("Could not find violation type: %s", entry.Violation)
			writeEntryErrorResponse(w, MissingViolationTypeError, i, entry, string("Violation type not found."))
			return
		}

		err = db.InsertOrUpdateReputationPenalty(nil, entry.Ip, uint(penalty))
		if _, ok := err.(CheckViolationError); ok {
			log.WithFields(log.Fields{"errno": InvalidReputationError}).Warnf("Reputation is outside of valid range [0-100]")
			writeEntryErrorResponse(w, InvalidReputationError, i, entry, string("Reputation is outside of valid range [0-100]"))
			return
		} else if err != nil {
			log.WithFields(log.Fields{"errno": DBError}).Warnf("Could not update reputation entry by violation: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		log.Debugf("Updated reputation for %s due to %d", entry.Ip, penalty)
	}

	w.WriteHeader(http.StatusNoContent)
}

// CreateReputation takes a JSON formatted IP reputation entry from
// the http request and inserts it to the database.
func CreateReputationHandler(w http.ResponseWriter, r *http.Request) {
	var entry ReputationEntry
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.WithFields(log.Fields{"errno": BodyReadError}).Warnf(DescribeErrno(BodyReadError))
		return
	}
	err = json.Unmarshal(body, &entry)
	if err != nil {
		log.WithFields(log.Fields{"errno": JSONUnmarshalError}).Warnf(DescribeErrno(JSONUnmarshalError), err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !IsValidReputationEntry(entry) {
		if !IsValidReputationCIDROrIP(entry.IP) {
			log.WithFields(log.Fields{"errno": InvalidIPError}).Infof(DescribeErrno(InvalidIPError), entry.IP)
		}
		if !IsValidReputation(entry.Reputation) {
			log.WithFields(log.Fields{"errno": InvalidReputationError}).Infof(DescribeErrno(InvalidReputationError), entry.Reputation)
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if db == nil {
		log.WithFields(log.Fields{"errno": MissingDB}).Warnf(DescribeErrno(MissingDB))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

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
		log.WithFields(log.Fields{"errno": DBError}).Warnf("Could not insert reputation entry: %s", err)
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
		log.WithFields(log.Fields{"errno": MissingIPError}).Infof(DescribeErrno(MissingIPError), r.URL.Path, err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if !IsValidReputationCIDROrIP(ip) {
		w.WriteHeader(http.StatusBadRequest)
		log.WithFields(log.Fields{"errno": InvalidIPError}).Infof(DescribeErrno(InvalidIPError), ip)
		return
	}

	var entry ReputationEntry
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.WithFields(log.Fields{"errno": BodyReadError}).Warnf(DescribeErrno(BodyReadError))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = json.Unmarshal(body, &entry)
	if err != nil {
		log.WithFields(log.Fields{"errno": JSONUnmarshalError}).Warnf(DescribeErrno(JSONUnmarshalError), err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	entry.IP = ip

	if !IsValidReputationEntry(entry) {
		if !IsValidReputationCIDROrIP(entry.IP) {
			log.WithFields(log.Fields{"errno": InvalidIPError}).Infof(DescribeErrno(InvalidIPError), entry.IP)
		}
		if !IsValidReputation(entry.Reputation) {
			log.WithFields(log.Fields{"errno": InvalidReputationError}).Infof(DescribeErrno(InvalidReputationError), entry.Reputation)
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if db == nil {
		log.WithFields(log.Fields{"errno": MissingDB}).Warnf(DescribeErrno(MissingDB))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = db.UpdateReputationEntry(nil, entry)
	if _, ok := err.(CheckViolationError); ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Reputation is outside of valid range [0-100]"))
	} else if err == ErrNoRowsAffected {
		w.WriteHeader(http.StatusNotFound)
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.WithFields(log.Fields{"errno": DBError}).Warnf("Could not update reputation entry: %s", err)
	} else if err == nil {
		w.WriteHeader(http.StatusOK)
	}
}

// DeleteReputation deletes an entry based on the IP address provided on the path
func DeleteReputationHandler(w http.ResponseWriter, r *http.Request) {
	ip, err := IPAddressFromHTTPPath(r.URL.Path)
	if err != nil {
		log.WithFields(log.Fields{"errno": MissingIPError}).Infof(DescribeErrno(MissingIPError), r.URL.Path, err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if !IsValidReputationCIDROrIP(ip) {
		log.WithFields(log.Fields{"errno": InvalidIPError}).Infof(DescribeErrno(InvalidIPError), ip)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if db == nil {
		log.WithFields(log.Fields{"errno": MissingDB}).Warnf(DescribeErrno(MissingDB))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = db.DeleteReputationEntry(nil, ReputationEntry{IP: ip})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.WithFields(log.Fields{"errno": DBError}).Warnf("Could not delete reputation entry: %s", err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// ReadReputation returns a JSON-formatted reputation entry from the database.
func ReadReputationHandler(w http.ResponseWriter, r *http.Request) {
	ip, err := IPAddressFromHTTPPath(r.URL.Path)
	if err != nil {
		log.WithFields(log.Fields{"errno": MissingIPError}).Infof(DescribeErrno(MissingIPError), r.URL.Path, err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if !IsValidReputationCIDROrIP(ip) {
		log.WithFields(log.Fields{"errno": InvalidIPError}).Infof(DescribeErrno(InvalidIPError), ip)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if db == nil {
		log.WithFields(log.Fields{"errno": MissingDB}).Warnf(DescribeErrno(MissingDB))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	entry, err := db.SelectSmallestMatchingSubnet(ip)
	if err == sql.ErrNoRows {
		w.WriteHeader(http.StatusNotFound)
		log.Debugf("No entries found for IP %s", ip)
		return
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.WithFields(log.Fields{"errno": DBError}).Warnf("Could not get reputation entry: %s", err)
		return
	}
	json, err := json.Marshal(entry)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.WithFields(log.Fields{"errno": JSONMarshalError}).Warnf(DescribeErrno(JSONMarshalError), "reputation", err)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(json)
}
