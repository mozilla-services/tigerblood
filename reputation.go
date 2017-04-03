package tigerblood

import (
	log "github.com/Sirupsen/logrus"
	"go.mozilla.org/mozlogrus"
	"database/sql"
	"net/http"
	"io/ioutil"
	"encoding/json"
)

func init() {
	mozlogrus.Enable("tigerblood")
}

func ReputationHandler(w http.ResponseWriter, req *http.Request) {
	SetResponseHeaders(w)

	if RequireHawkAuth(w, req) != nil {
		return
	}

	switch req.Method {
	case "GET":
		ReadReputationHandler(w, req)
	case "POST":
		CreateReputationHandler(w, req)
	case "PUT":
		UpdateReputationHandler(w, req)
	case "DELETE":
		DeleteReputationHandler(w, req)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// Route{
// 	"ReadReputation",
// 	"GET",
// 	"/{ip:[[:punct:]\\/\\.\\w]{1,128}}", // see above note for all punct for IPs w/ colons e.g. 2001:db8::/32
// 	ReadReputationHandler,
// },
// Route{
// 	"CreateReputation",
// 	"POST",
// 	"/",
// 	CreateReputationHandler,
// },
// Route{
// 	"UpdateReputation",
// 	"PUT",
// 	"/{ip:[[:punct:]\\/\\.\\w]{1,128}}",
// 	UpdateReputationHandler,
// },
// Route{
// 	"DeleteReputation",
// 	"DELETE",
// 	"/{ip:[[:punct:]\\/\\.\\w]{1,128}}",
// 	DeleteReputationHandler,
// },


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
