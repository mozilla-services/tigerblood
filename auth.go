package tigerblood

import (
	log "github.com/sirupsen/logrus"
	"net/http"
	"strings"
)

// Bits used in authentication modes bitmask
const (
	AuthEnableHawk = 1 << iota
	AuthEnableAPIKey
)

const (
	AuthRequestUnknown = iota
	AuthRequestAPIKey
	AuthRequestHawk
)

// authmodes is a bit mask indicating authentication modes to support and influences the
// behavior of the RequireAuth handler.
var authModes int

var hawkData *HawkData
var apiKeyData *APIKeyData

// APIKeyData is configuration data representing valid API authentication keys, where
// the key is just an identifier and the value is the actual secret.
type APIKeyData struct {
	credentials map[string]string
}

// SetAuthMask sets the forms of authentication tigerblood will accept for requests. mask
// is a bitmask indicating the supported authentication types; if no bits are set authentication
// is disabled.
func SetAuthMask(mask int) {
	authModes = mask
}

// SetHawkCredentials configures the credentials to be used for hawk authentication
func SetHawkCredentials(credentials map[string]string) {
	hawkData = NewHawkData(credentials)
}

// SetAPIKeyCredentials configures the credentials to be used for API key authentication
func SetAPIKeyCredentials(credentials map[string]string) {
	apiKeyData = NewAPIKeyData(credentials)
}

// NewAPIKeyData returns API key config data from a map of API key credentials
func NewAPIKeyData(secrets map[string]string) *APIKeyData {
	return &APIKeyData{
		credentials: secrets,
	}
}

func getAuthRequestType(h string) int {
	if strings.HasPrefix(h, "Hawk ") {
		return AuthRequestHawk
	} else if strings.HasPrefix(h, "APIKey ") {
		return AuthRequestAPIKey
	}
	return AuthRequestUnknown
}

// RequireAuth middleware for validating authentication credentials
func RequireAuth() Middleware {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, ok := UnauthedRoutes[r.URL.Path]; ok {
				// Authentication not required, continue
				log.Warnf("Skipping auth for route: %s", r.URL.Path)
				h.ServeHTTP(w, r)
				return
			}

			if authModes == 0 {
				// Authentication is disabled, continue
				h.ServeHTTP(w, r)
				return
			}

			success := false
			authtype := getAuthRequestType(r.Header.Get("Authorization"))
			if (authModes&AuthEnableAPIKey != 0) && authtype == AuthRequestAPIKey {
				success = APIKeyAuth(r, apiKeyData)
			} else if authModes&AuthEnableHawk != 0 && authtype == AuthRequestHawk {
				success = HawkAuth(r, hawkData)
			}
			if !success {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			// Authentication successful, continue
			h.ServeHTTP(w, r)
		})
	}
}

// APIKeyAuth authenticates API key based requests, returns true if successful
func APIKeyAuth(r *http.Request, m *APIKeyData) bool {
	hdr := r.Header.Get("Authorization")
	if hdr == "" {
		log.WithFields(log.Fields{"errno": APIKeyNotSpecified}).Warnf("apikey: no key specified")
		return false
	}

	hdr = strings.TrimPrefix(hdr, "APIKey ")
	for k, v := range m.credentials {
		if hdr == v {
			log.WithFields(log.Fields{"id": k}).Infof("apikey: accepted request")
			return true
		}
	}
	log.WithFields(log.Fields{"errno": APIKeyInvalid}).Warnf("apikey: invalid key specified")
	return false
}
