package tigerblood

import (
	"bufio"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"time"
)

// exceptionTestHook if true will cause any go-routines spawned in the exception subsystem
// to immediately exit (e.g., no dynamic exception updates will run, and the expiry thread will
// not run
var exceptionTestHook bool

const awsUpdateInterval = time.Minute * 60
const awsIPRangeURL = "https://ip-ranges.amazonaws.com/ip-ranges.json"

// awsIPRanges is used to unmarshal the data we should get from awsIPRangeURL
type awsIPRanges struct {
	Prefixes []struct {
		IPPrefix string `json:"ip_prefix"`
	} `json:"prefixes"`
	IPV6Prefixes []struct {
		IPV6Prefix string `json:"ipv6_prefix"`
	} `json:"ipv6_prefixes"`
}

// ExceptionSource is an interface defining generic functions that all sources of
// exception information must implement
type exceptionSource interface {
	getName() string
	getExceptions() ([]ExceptionEntry, error)
	isStatic() bool
	getCreatorPrefix() string
	updateInterval() *time.Duration
}

// allExceptionTypes should contain a list of all types that represent a source of
// exception information, and is used by housekeeping operations for example to
// identify static exception sources and purge old database entries on initialization
var allExceptionTypes = []exceptionSource{
	&exceptionFile{},
	&exceptionAWS{},
}

// exceptionCalcExpiry is a helper function to calculate when an exception should expire
// based on the polling interval for the source. Return a time period 3 minutes beyond
// the polling interval to ensure out of date records are removed.
func exceptionCalcExpiry(e exceptionSource) time.Time {
	return time.Now().Add(*e.updateInterval()).Add(time.Minute * 3)
}

// exceptionFile is a type that stores exception information read from a file
// on the file system.
type exceptionFile struct {
	Path string
}

func (e *exceptionFile) getName() string {
	return fmt.Sprintf("%s:%s", e.getCreatorPrefix(), e.Path)
}

func (e *exceptionFile) getCreatorPrefix() string {
	return "file"
}

func (e *exceptionFile) isStatic() bool {
	return true
}

func (e *exceptionFile) updateInterval() *time.Duration {
	// noop for file
	return nil
}

func (e *exceptionFile) getExceptions() (ret []ExceptionEntry, err error) {
	fd, err := os.Open(e.Path)
	if err != nil {
		return
	}
	defer fd.Close()

	scn := bufio.NewScanner(fd)
	for scn.Scan() {
		buf := scn.Text()
		_, _, err = net.ParseCIDR(buf)
		if err != nil {
			return
		}
		ret = append(ret, ExceptionEntry{
			Creator: e.getName(),
			IP:      buf,
		})

	}
	err = scn.Err()
	return
}

// exceptionAWS is a type that stores exception information read from the AWS public IP
// address endpoint (see https://ip-ranges.amazonaws.com/ip-ranges.json)
type exceptionAWS struct {
}

func (e *exceptionAWS) getExceptions() (ret []ExceptionEntry, err error) {
	resp, err := http.Get(awsIPRangeURL)
	if err != nil {
		return ret, err
	}
	defer resp.Body.Close()
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ret, err
	}
	var awsp awsIPRanges
	err = json.Unmarshal(buf, &awsp)
	if err != nil {
		return ret, err
	}
	for _, v := range awsp.Prefixes {
		nent := ExceptionEntry{
			Creator: e.getName(),
			IP:      v.IPPrefix,
			Expires: exceptionCalcExpiry(e),
		}
		ret = append(ret, nent)
	}
	return
}

func (e *exceptionAWS) getName() string {
	return e.getCreatorPrefix() + ":"
}

func (e *exceptionAWS) getCreatorPrefix() string {
	return "awsiprange"
}

func (e *exceptionAWS) isStatic() bool {
	return false
}

func (e *exceptionAWS) updateInterval() *time.Duration {
	ret := awsUpdateInterval
	return &ret
}

// InitializeExceptions performs initial housekeeping of the exception table, purging
// old static data and starting routines that periodically import non-static exception
// information.
//
// InitializeExceptions should be called prior to performing any API processing of
// reputation information.
func InitializeExceptions() error {
	// Purge exceptions based on static information (files)
	for _, v := range allExceptionTypes {
		if !v.isStatic() {
			continue
		}
		err := db.DeleteExceptionCreatorType(nil, v.getCreatorPrefix())
		if err != nil {
			return err
		}
	}
	// Add any exceptions we have sourced from static information (files)
	for _, v := range exceptionSources {
		if !v.isStatic() {
			continue
		}
		except, err := v.getExceptions()
		if err != nil {
			return err
		}
		for _, w := range except {
			err = db.InsertOrUpdateExceptionEntry(nil, w)
			if err != nil {
				return err
			}
		}
	}
	// Start routine to purge expired exceptions
	go func() {
		if exceptionTestHook {
			return
		}
		log.Print("Starting expired exception purge routine")
		for {
			err := db.DeleteExpiredExceptions(nil)
			if err != nil {
				// If something goes wrong deleting expired exceptions treat
				// this as fatal
				log.Fatalf("Error removing expired exceptions: %s", err)
			}
			time.Sleep(time.Second * 60)
		}
	}()
	// For each dynamic exception source, start an update routine
	for i := range exceptionSources {
		if exceptionSources[i].isStatic() {
			continue
		}
		ne := exceptionSources[i]
		go func() {
			if exceptionTestHook {
				return
			}
			for {
				// XXX If anything fails here treat this as fatal right now, as we
				// no longer have up to date exception information. This could probably
				// be handled better, potentially by noting a critical error and
				// disabling purge of expired exceptions for this source until we have
				// valid data again.
				log.Printf("Update exceptions for %s", ne.getName())
				ent, err := ne.getExceptions()
				if err != nil {
					log.Fatalf("Error updating exception: %s", err)
				}
				for _, w := range ent {
					err = db.InsertOrUpdateExceptionEntry(nil, w)
					if err != nil {
						log.Fatalf("Error updating exception: %s", err)
					}
				}
				time.Sleep(*ne.updateInterval())
			}
		}()
	}
	return nil
}

// AddFileException adds a new exception source with exception data sourced from a file
// on the file system, with one CIDR entry per line
func AddFileException(path string) error {
	// Make sure we can open the file indicated in path
	fd, err := os.Open(path)
	if err != nil {
		return err
	}
	fd.Close()
	exceptionSources = append(exceptionSources, &exceptionFile{Path: path})
	return nil
}

// AddAWSException adds a new exception source that periodically fetches AWS public address
// data and adds exceptions for it
func AddAWSException() error {
	exceptionSources = append(exceptionSources, &exceptionAWS{})
	return nil
}
