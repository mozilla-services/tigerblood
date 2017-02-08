package tigerblood

import (
	"go.mozilla.org/hawk"
	"net/http"
	"time"
	"log"
	"bytes"
	"crypto/sha256"
	"github.com/willf/bloom"
	"io"
	"io/ioutil"
	"sync"
)

type HawkHandler struct {
	handler     http.Handler
	credentials map[string]string

	bloomPrev     *bloom.BloomFilter
	bloomNow      *bloom.BloomFilter
	bloomHalflife time.Duration
	lastRotate    time.Time
	bloomLock     sync.Mutex
}

func NewHawkHandler(handler http.Handler, secrets map[string]string) *HawkHandler {
	var requestsPerHalfLife uint = 2000 * 30
	var bitsPerRequest uint = 50
	return &HawkHandler{
		handler:       handler,
		credentials:   secrets,
		bloomPrev:     bloom.New(requestsPerHalfLife * bitsPerRequest, 5),
		bloomNow:      bloom.New(requestsPerHalfLife * bitsPerRequest, 5),
		bloomHalflife: 30 * time.Second,
		lastRotate:    time.Now(),
	}
}

func (h *HawkHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/__lbheartbeat__" &&
		r.URL.Path != "/__heartbeat__" &&
		r.URL.Path != "/__version__" {
		auth, err := hawk.NewAuthFromRequest(r, h.lookupCredentials, h.lookupNonce)
		if err != nil || auth.Valid() != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if r.Body != nil {
			buf, err := ioutil.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			r.Body = ioutil.NopCloser(bytes.NewBuffer(buf))
			hash := auth.PayloadHash(r.Header.Get("Content-Type"))

			bytesWritten, copyErr := io.Copy(hash, ioutil.NopCloser(bytes.NewBuffer(buf)))
			if copyErr != nil {
				log.Printf("Error copying request body: %s ", copyErr)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			if bytesWritten > 0 && !auth.ValidHash(hash) {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
		}
	}
	// Authentication successful, continue
	h.handler.ServeHTTP(w, r)
}

func (h *HawkHandler) rotate() {
	h.bloomNow, h.bloomPrev = h.bloomPrev, h.bloomNow
	h.bloomNow.ClearAll()
	h.lastRotate = time.Now()
}

func (h *HawkHandler) lookupNonce(nonce string, t time.Time, credentials *hawk.Credentials) bool {
	h.bloomLock.Lock()
	if time.Now().Sub(h.lastRotate) > h.bloomHalflife {
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

func (h *HawkHandler) lookupCredentials(creds *hawk.Credentials) error {
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
