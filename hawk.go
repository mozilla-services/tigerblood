package tigerblood

import (
	"go.mozilla.org/hawk"
	"net/http"
	"time"
	"bytes"
	"crypto/sha256"
	"github.com/willf/bloom"
	"io"
	"io/ioutil"
	"sync"
)

type HawkData struct {
	credentials map[string]string

	bloomPrev     *bloom.BloomFilter
	bloomNow      *bloom.BloomFilter
	lastRotate    time.Time
	bloomLock     sync.Mutex
}

const (
	bloomHalflife time.Duration = 30 * time.Second
	requestsPerHalfLife uint = 2000 * 30
	bitsPerRequest uint = 50
)

func NewHawkData(secrets map[string]string) *HawkData {
	return &HawkData{
		credentials:   secrets,

		bloomPrev:     bloom.New(requestsPerHalfLife * bitsPerRequest, 5),
		bloomNow:      bloom.New(requestsPerHalfLife * bitsPerRequest, 5),
		lastRotate:    time.Now(),
	}
}

func RequireHawkAuth(credentials map[string]string) Middleware {
	m := NewHawkData(credentials)

	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, ok := UnauthedRoutes[r.URL.Path]; ok {
				// Authentication not required, continue
				h.ServeHTTP(w, r)
				return
			}

			auth, err := hawk.NewAuthFromRequest(r, m.lookupCredentials, m.lookupNonce)
			if err != nil || auth.Valid() != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			buf, err := ioutil.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			r.Body = ioutil.NopCloser(bytes.NewBuffer(buf))
			hash := auth.PayloadHash(r.Header.Get("Content-Type"))
			io.Copy(hash, ioutil.NopCloser(bytes.NewBuffer(buf)))
			if !auth.ValidHash(hash) {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			// Authentication successful, continue
			h.ServeHTTP(w, r)
		})
	}
}

func (h *HawkData) rotate() {
	h.bloomNow, h.bloomPrev = h.bloomPrev, h.bloomNow
	h.bloomNow.ClearAll()
	h.lastRotate = time.Now()
}

func (h *HawkData) lookupNonce(nonce string, t time.Time, credentials *hawk.Credentials) bool {
	h.bloomLock.Lock()
	if time.Now().Sub(h.lastRotate) > bloomHalflife {
		h.rotate()
	}
	h.bloomLock.Unlock()
	key := nonce + t.String() + credentials.ID
	if h.bloomNow.TestString(key) || h.bloomPrev.TestString(key) {
		return false
	}
	h.bloomNow.AddString(key)
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
