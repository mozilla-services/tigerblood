package tigerblood

import (
	log "github.com/Sirupsen/logrus"
	"go.mozilla.org/mozlogrus"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"strings"
)

func init() {
	mozlogrus.Enable("tigerblood")
}

// Returns a list of known violations for debugging
func ListViolationsHandler(w http.ResponseWriter, req *http.Request) {
	SetResponseHeaders(w)

	if RequireHawkAuth(w, req) != nil {
		return
	}

	if req.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if violationPenalties == nil || violationPenaltiesJson == nil {
		log.WithFields(log.Fields{"errno": MissingViolations}).Warnf(DescribeErrno(MissingViolations))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(violationPenaltiesJson)
}

// Route{
// 	"UpsertReputationByViolation",
// 	"PUT",
// 	"/violations/{type:[[:punct:]\\w]{1,255}}",  // include all :punct: since gorilla/mux barfed trying to limit it to `:` (or as \x3a)
// 	UpsertReputationByViolationHandler,
// },

// UpsertReputationByViolation takes a JSON body from the http request
// and upserts the reputation entry on the database to the reputation
// given in reputation violation.  The HTTP requests path has to
// contain the IP to be updated, in CIDR notation. For example:
// {"Violation": "password-reset-rate-limit-exceeded"}
func UpsertReputationByViolationHandler(w http.ResponseWriter, req *http.Request) {
	SetResponseHeaders(w)

	if RequireHawkAuth(w, req) != nil {
		return
	}

	if req.Method != "PUT" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	splitPath := strings.Split(req.URL.Path, "/")
	if len(splitPath) != 3 {
		log.WithFields(log.Fields{"errno": InvalidIPError}).Infof(DescribeErrno(InvalidIPError), req.URL.Path)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	ip, err := IPAddressFromHTTPPath("/" + splitPath[2])

	if err != nil {
		log.WithFields(log.Fields{"errno": MissingIPError}).Infof(DescribeErrno(MissingIPError), req.URL.Path, err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !IsValidReputationCIDROrIP(ip) {
		log.WithFields(log.Fields{"errno": InvalidIPError}).Infof(DescribeErrno(InvalidIPError), splitPath[2])
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	body, err := ioutil.ReadAll(req.Body)
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
