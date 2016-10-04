package tigerblood

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/DataDog/datadog-go/statsd"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

type TigerbloodHandler struct {
	db     *DB
	statsd *statsd.Client
}

func NewTigerbloodHandler(db *DB, statsd *statsd.Client) *TigerbloodHandler {
	return &TigerbloodHandler{
		db:     db,
		statsd: statsd,
	}
}

// IPAddressFromHTTPPath takes a HTTP path and returns an IPv4 IP if it's found, or an error if none is found.
func IPAddressFromHTTPPath(path string) (string, error) {
	path = path[1:len(path)]
	ip, network, err := net.ParseCIDR(path)
	if err != nil {
		if strings.Contains(path, "/") {
			return "", fmt.Errorf("Error getting IP from HTTP path: %s", err)
		}
		ip = net.ParseIP(path)
		if ip == nil {
			return "", fmt.Errorf("Error getting IP from HTTP path: %s", err)
		}
		network = &net.IPNet{}
		if ip.To4() != nil {
			network.Mask = net.CIDRMask(32, 32)
		} else if ip.To16() != nil {
			network.Mask = net.CIDRMask(128, 128)
		}
	}
	network.IP = ip
	return network.String(), nil
}

func (h *TigerbloodHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	switch r.URL.Path {
	case "/":
		switch r.Method {
		case "POST":
			h.CreateReputation(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	case "/__heartbeat__":
		err := h.db.Ping()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
		}
		return
	case "/__lbheartbeat__":
		w.WriteHeader(http.StatusOK)
		return
	case "/__version__":
		h.handleVersion(w, r)
		return
	default:
		switch r.Method {
		case "GET":
			h.ReadReputation(w, r)
		case "PUT":
			h.UpdateReputation(w, r)
		case "DELETE":
			h.DeleteReputation(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
	if h.statsd != nil {
		h.statsd.Histogram("request.duration", float64(time.Since(startTime).Nanoseconds())/float64(1e6), nil, 1)
	}
	if time.Since(startTime).Nanoseconds() > 1e7 {
		log.Printf("Request took %s to proces\n", time.Since(startTime))
	}
}

func (h *TigerbloodHandler) handleVersion(w http.ResponseWriter, req *http.Request) {
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

// CreateReputation takes a JSON formatted IP reputation entry from the http request and inserts it to the database.
func (h *TigerbloodHandler) CreateReputation(w http.ResponseWriter, r *http.Request) {
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
	err = h.db.InsertReputationEntry(nil, entry)
	if _, ok := err.(CheckViolationError); ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Reputation is outside of valid range [0-100]"))
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	if err != nil {
		log.Printf("Could not insert reputation entry: %s", err)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

// ReadReputation returns a JSON-formatted reputation entry from the database.
func (h *TigerbloodHandler) ReadReputation(w http.ResponseWriter, r *http.Request) {
	ip, err := IPAddressFromHTTPPath(r.URL.Path)
	if err != nil {
		// This means there was no IP address found in the path
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("No IP address found in path %s: %s", r.URL.Path, err)
		return
	}
	entry, err := h.db.SelectSmallestMatchingSubnet(ip)
	if err == sql.ErrNoRows {
		w.WriteHeader(http.StatusNotFound)
		log.Printf("No entries found for IP %s", ip)
		if h.statsd != nil {
			h.statsd.Incr("misses", nil, 1.0)
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
	if h.statsd != nil {
		h.statsd.Incr("hits", nil, 1.0)
	}
	w.WriteHeader(http.StatusOK)
	w.Write(json)
}

// UpdateReputation takes a JSON body from the http request and updates that reputation entry on the database.
// The HTTP requests path has to contain the IP to be updated, in CIDR notation. The body can contain the IP address, or it can be omitted. For example:
// {"Reputation": 50} or {"Reputation": 50, "IP":, "192.168.0.1"}. The IP in the JSON body will be ignored.
func (h *TigerbloodHandler) UpdateReputation(w http.ResponseWriter, r *http.Request) {
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
	err = h.db.UpdateReputationEntry(nil, entry)
	if _, ok := err.(CheckViolationError); ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Reputation is outside of valid range [0-100]"))
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	if err != nil {
		log.Printf("Could not update reputation entry: %s", err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// DeleteReputation deletes an entry based on the IP address provided on the path
func (h *TigerbloodHandler) DeleteReputation(w http.ResponseWriter, r *http.Request) {
	ip, err := IPAddressFromHTTPPath(r.URL.Path)
	if err != nil {
		// This means there was no IP address found in the path
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("No IP address found in path %s: %s", r.URL.Path, err)
		return
	}
	err = h.db.DeleteReputationEntry(nil, ReputationEntry{IP: ip})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Could not update reputation entry: %s", err)
		return
	}
	w.WriteHeader(http.StatusOK)
}
