package tigerblood

import (
	"bytes"
	"crypto/sha256"
	log "github.com/sirupsen/logrus"
	"go.mozilla.org/hawk"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"time"
)

// HawkData is hawk config data (credentials is a map of Hawk IDs to passwords)
type HawkData struct {
	credentials map[string]string
}

// NewHawkData returns hawk config data for a map of hawk creds
func NewHawkData(secrets map[string]string) *HawkData {
	return &HawkData{
		credentials: secrets,
	}
}

// HawkAuth authenticates hawk requests, returns true if successful.
func HawkAuth(r *http.Request, m *HawkData) bool {
	// Validate the Hawk header format and credentials
	auth, err := hawk.NewAuthFromRequest(r, m.lookupCredentials, m.lookupNonceNop)
	if err != nil {
		switch err.(type) {
		case hawk.AuthFormatError:
			log.WithFields(log.Fields{"errno": HawkAuthFormatError}).Warn(err)
		case *hawk.CredentialError:
			log.WithFields(log.Fields{"errno": HawkCredError}).Warn(err)
		case hawk.AuthError:
			switch err.(hawk.AuthError) {
			case hawk.ErrNoAuth:
				log.WithFields(log.Fields{"errno": HawkErrNoAuth}).Warn(err)
			case hawk.ErrReplay:
				log.WithFields(log.Fields{"errno": HawkReplayError}).Warn(err)
			}
		default:
			log.WithFields(log.Fields{"errno": HawkOtherAuthError}).Warnf("other hawk auth error: %s",
				err)
		}
		return false
	}

	// Validate the header MAC and skew
	validationError := auth.Valid()
	if validationError != nil {
		log.WithFields(log.Fields{"errno": HawkValidationError}).Warnf("hawk validation error: %s",
			validationError)
		return false
	}

	// Validate the payload hash of the request Content-Type and body
	// assuming bodies will fit in memory always validate the body
	contentType := r.Header.Get("Content-Type")
	if r.Method != "GET" && r.Method != "DELETE" && contentType == "" {
		log.WithFields(log.Fields{"errno": HawkMissingContentType}).Warn("hawk: missing content-type")
		return false
	}

	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil && contentType != "" {
		log.WithFields(log.Fields{"errno": HawkMissingContentType}).Warnf("hawk: invalid content-type %s",
			err)
		return false
	}

	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.WithFields(log.Fields{"errno": HawkReadBodyError}).Warnf("hawk: error reading body %s", err)
		return false
	}

	r.Body = ioutil.NopCloser(bytes.NewBuffer(buf))
	hash := auth.PayloadHash(mediaType)
	io.Copy(hash, ioutil.NopCloser(bytes.NewBuffer(buf)))
	if !auth.ValidHash(hash) {
		log.WithFields(log.Fields{"errno": HawkInvalidBodyHash}).Warnf("hawk: invalid payload hash")
		return false
	}

	log.WithFields(log.Fields{"id": auth.Credentials.ID}).Infof("hawk: accepted request")
	return true
}

func (h *HawkData) lookupNonceNop(nonce string, t time.Time, credentials *hawk.Credentials) bool {
	return true
}

func (h *HawkData) lookupCredentials(creds *hawk.Credentials) error {
	creds.Key = "-"
	creds.Hash = sha256.New
	if cred, ok := h.credentials[creds.ID]; ok {
		creds.Key = cred
		return nil
	}
	return &hawk.CredentialError{
		Type:        hawk.UnknownID,
		Credentials: creds,
	}
}
