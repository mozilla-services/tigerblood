package tigerblood

import (
	log "github.com/Sirupsen/logrus"
	"go.mozilla.org/mozlogrus"
	"fmt"
	"net/http"
	"os"
	"path"
)

func init() {
	mozlogrus.Enable("tigerblood")
}

func LoadBalancerHeartbeatHandler(w http.ResponseWriter, req *http.Request) {
	SetResponseHeaders(w)

	if req.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	w.WriteHeader(http.StatusOK)
	return
}

func HeartbeatHandler(w http.ResponseWriter, req *http.Request) {
	SetResponseHeaders(w)

	if req.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

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
	SetResponseHeaders(w)

	if req.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

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
