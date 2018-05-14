package tigerblood

import (
	"database/sql"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"os"
	"path"
)

// LoadBalancerHeartbeatHandler returns 200 if the server is up
func LoadBalancerHeartbeatHandler(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	return
}

// HeartbeatHandler pings the DB and returns 200 or 500
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

// VersionHandler returns the version.json file
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
		log.WithFields(log.Fields{"errno": FileNotFound}).Warnf(DescribeErrno(FileNotFound),
			"version.json", err)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	defer f.Close()
	stat, err := f.Stat()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	http.ServeContent(w, req, "__version__", stat.ModTime(), f)
}

// ListViolationsHandler returns a JSON array of known violations for debugging
func ListViolationsHandler(w http.ResponseWriter, req *http.Request) {
	if violationPenalties == nil || violationPenaltiesJSON == nil {
		log.WithFields(log.Fields{"errno": MissingViolations}).Warnf(DescribeErrno(MissingViolations))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(violationPenaltiesJSON)
}

// ListExceptionsHandler returns a JSON array of all active exceptions
func ListExceptionsHandler(w http.ResponseWriter, req *http.Request) {
	if db == nil {
		log.WithFields(log.Fields{"errno": MissingDB}).Warnf(DescribeErrno(MissingDB))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	entries, err := db.SelectAllExceptions()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.WithFields(log.Fields{"errno": DBError}).Warnf("Could not list exceptions: %s", err)
		return
	}
	if len(entries) == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	json, err := json.Marshal(entries)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.WithFields(log.Fields{"errno": JSONMarshalError}).Warnf(DescribeErrno(JSONMarshalError),
			"exceptions", err)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(json)
}

func writeEntryErrorResponse(w http.ResponseWriter, entryIndex int, entry IPViolationEntry, statusCode int,
	msg string) {
	entryError := struct {
		EntryIndex int
		Entry      IPViolationEntry
		Msg        string
	}{
		entryIndex,
		entry,
		msg,
	}
	j, err := json.Marshal(entryError)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.WithFields(log.Fields{"errno": JSONMarshalError}).Warnf(DescribeErrno(JSONMarshalError),
			"entry error", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(j)
	return
}

// MultiUpsertReputationByViolationHandler creates or update reputation entries for many IPViolationEntries
func MultiUpsertReputationByViolationHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.WithFields(log.Fields{"errno": BodyReadError}).Warnf(DescribeErrno(BodyReadError), err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	var entries []IPViolationEntry
	err = json.Unmarshal(body, &entries)
	if err != nil {
		log.WithFields(log.Fields{"errno": JSONUnmarshalError}).Warnf(DescribeErrno(JSONUnmarshalError),
			err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if len(entries) < 1 {
		log.WithFields(log.Fields{
			"errno": MissingIPViolationEntryError,
		}).Warn(DescribeErrno(MissingIPViolationEntryError))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if len(entries) > maxEntries {
		log.WithFields(log.Fields{
			"errno": TooManyIPViolationEntriesError,
		}).Warn(DescribeErrno(TooManyIPViolationEntriesError))
		w.WriteHeader(http.StatusRequestEntityTooLarge)
		return
	}

	if db == nil {
		log.WithFields(log.Fields{"errno": MissingDB}).Warnf(DescribeErrno(MissingDB))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var seenIps = make(map[string]bool)
	var ips = make([]string, len(entries))
	var penalties = make([]uint, len(entries))

	for i, entry := range entries {
		penalty, errno := ValidateIPViolationEntryAndGetPenalty(entry)
		if errno > 0 {
			switch errno {
			case MissingIPError:
				writeEntryErrorResponse(w, i, entry, http.StatusBadRequest,
					fmt.Sprintf(DescribeErrno(MissingIPError)))
			case MissingViolations:
				writeEntryErrorResponse(w, i, entry, http.StatusBadRequest,
					DescribeErrno(MissingViolations))
			case MissingViolationTypeError:
				writeEntryErrorResponse(w, i, entry, http.StatusBadRequest,
					fmt.Sprintf(DescribeErrno(MissingViolationTypeError), entry.Violation))
			case InvalidIPError:
				writeEntryErrorResponse(w, i, entry, http.StatusBadRequest,
					fmt.Sprintf(DescribeErrno(InvalidIPError), entry.IP))
			case InvalidViolationTypeError:
				writeEntryErrorResponse(w, i, entry, http.StatusBadRequest,
					fmt.Sprintf(DescribeErrno(InvalidViolationTypeError), entry.Violation))
			default:
				writeEntryErrorResponse(w, i, entry, http.StatusBadRequest, string(""))
			}
			return
		}

		if _, ok := seenIps[entry.IP]; ok {
			writeEntryErrorResponse(w, i, entry, http.StatusConflict,
				fmt.Sprintf(DescribeErrno(DuplicateIPError), entry.IP))
			return
		}
		seenIps[entry.IP] = true
		ips[i], penalties[i] = entry.IP, penalty
	}

	err = db.InsertOrUpdateReputationPenalties(nil, ips, penalties)
	if err != nil {
		log.WithFields(log.Fields{
			"errno": DBError,
		}).Warnf("Could not update reputation entry by violation: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Debugf("Updated %s reputations", len(entries))
	w.WriteHeader(http.StatusNoContent)
}

// UpdateReputationHandler takes a JSON body from the http request and updates that reputation
// entry in the database. The HTTP requests path has to contain the IP to be updated, in CIDR
// notation. The body can contain the IP address, or it can be omitted.
//
// For example:
//
// {"Reputation": 50} or {"Reputation": 50, "IP": "192.168.0.1"}.
//
// The IP in the JSON body will be ignored.
func UpdateReputationHandler(w http.ResponseWriter, r *http.Request) {
	ip, err := IPAddressFromHTTPPath(r.URL.Path)
	if err != nil {
		log.WithFields(log.Fields{"errno": MissingIPError}).Infof(DescribeErrno(MissingIPError))
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
		log.WithFields(log.Fields{"errno": BodyReadError}).Warnf(DescribeErrno(BodyReadError), err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = json.Unmarshal(body, &entry)
	if err != nil {
		log.WithFields(log.Fields{"errno": JSONUnmarshalError}).Warnf(DescribeErrno(JSONUnmarshalError),
			err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	entry.IP = ip

	if !IsValidReputationEntry(entry) {
		if !IsValidReputationCIDROrIP(entry.IP) {
			log.WithFields(log.Fields{"errno": InvalidIPError}).Infof(DescribeErrno(InvalidIPError),
				entry.IP)
		}
		if !IsValidReputation(entry.Reputation) {
			log.WithFields(log.Fields{
				"errno": InvalidReputationError,
			}).Infof(DescribeErrno(InvalidReputationError), entry.Reputation)
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if db == nil {
		log.WithFields(log.Fields{"errno": MissingDB}).Warnf(DescribeErrno(MissingDB))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	retrep, err := db.InsertOrUpdateReputationEntry(nil, entry)
	if _, ok := err.(CheckViolationError); ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Reputation is outside of valid range [0-100]"))
		return
	} else if err == ErrNoRowsAffected {
		// This can occur if the address is in the exception list, don't treat this as
		// an error
		w.WriteHeader(http.StatusOK)
		return
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.WithFields(log.Fields{"errno": DBError}).Warnf("Could not update reputation entry: %s", err)
		return
	}
	log.WithFields(log.Fields{"ip": entry.IP, "reputation": retrep}).Infof("reputation set")
	w.WriteHeader(http.StatusOK)
}

// DeleteReputationHandler deletes an entry based on the IP address provided on the path
func DeleteReputationHandler(w http.ResponseWriter, r *http.Request) {
	ip, err := IPAddressFromHTTPPath(r.URL.Path)
	if err != nil {
		log.WithFields(log.Fields{"errno": MissingIPError}).Infof(DescribeErrno(MissingIPError))
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

// ReadReputationHandler returns a JSON-formatted reputation entry from the database.
func ReadReputationHandler(w http.ResponseWriter, r *http.Request) {
	ip, err := IPAddressFromHTTPPath(r.URL.Path)
	if err != nil {
		log.WithFields(log.Fields{"errno": MissingIPError}).Infof(DescribeErrno(MissingIPError))
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
		log.WithFields(log.Fields{"errno": JSONMarshalError}).Warnf(DescribeErrno(JSONMarshalError),
			"reputation", err)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(json)
}
