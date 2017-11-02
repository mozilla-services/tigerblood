package tigerblood

import (
	"bytes"
	"crypto/sha256"
	log "github.com/Sirupsen/logrus"
	"go.mozilla.org/hawk"
	"go.mozilla.org/mozlogrus"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"time"
)

type HawkData struct {
	credentials map[string]string
}

func init() {
	mozlogrus.Enable("tigerblood")
}

func NewHawkData(secrets map[string]string) *HawkData {
	return &HawkData{
		credentials: secrets,
	}
}

func RequireHawkAuth(credentials map[string]string) Middleware {
	m := NewHawkData(credentials)

	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, ok := UnauthedRoutes[r.URL.Path]; ok {
				// Authentication not required, continue
				log.Debugf("Skipping auth for route: %s", r.URL.Path)
				h.ServeHTTP(w, r)
				return
			}

			// Validate the Hawk header format and credentials
			auth, err := hawk.NewAuthFromRequest(r, m.lookupCredentials, m.lookupNonceNop)
			if err != nil {
				switch err.(type) {
				case hawk.AuthFormatError:
					log.WithFields(log.Fields{"errno": HawkAuthFormatError}).Warn(err)
				case *hawk.CredentialError:
					log.WithFields(log.Fields{"errno": HawkCredError}).Warn(err)
				case hawk.AuthError:
					{
						switch err.(hawk.AuthError) {
						case hawk.ErrNoAuth:
							log.WithFields(log.Fields{"errno": HawkErrNoAuth}).Warn(err)
						case hawk.ErrReplay:
							log.WithFields(log.Fields{"errno": HawkReplayError}).Warn(err)
						}
					}
				default:
					log.WithFields(log.Fields{"errno": HawkOtherAuthError}).Warnf("other hawk auth error: %s", err)
				}
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			// Validate the header MAC and skew
			validationError := auth.Valid()
			if validationError != nil {
				log.WithFields(log.Fields{"errno": HawkValidationError}).Warnf("hawk validation error: %s", validationError)
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			// Validate the payload hash of the request Content-Type and body
			// assuming bodies will fit in memory always validate the body
			contentType := r.Header.Get("Content-Type")
			if r.Method != "GET" && r.Method != "DELETE" && contentType == "" {
				log.WithFields(log.Fields{"errno": HawkMissingContentType}).Warn("hawk: missing content-type")
			}

			mediaType, _, err := mime.ParseMediaType(contentType)
			if err != nil && contentType != "" {
				log.WithFields(log.Fields{"errno": HawkMissingContentType}).Warnf("hawk: invalid content-type %s", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			buf, err := ioutil.ReadAll(r.Body)
			if err != nil {
				log.WithFields(log.Fields{"errno": HawkReadBodyError}).Warnf("hawk: error reading body %s", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			r.Body = ioutil.NopCloser(bytes.NewBuffer(buf))
			hash := auth.PayloadHash(mediaType)
			io.Copy(hash, ioutil.NopCloser(bytes.NewBuffer(buf)))
			if !auth.ValidHash(hash) {
				log.WithFields(log.Fields{"errno": HawkInvalidBodyHash}).Warnf("hawk: invalid payload hash")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			// Authentication successful, continue
			h.ServeHTTP(w, r)
		})
	}
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
