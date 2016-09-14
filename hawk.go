package tigerblood

import (
	"go.mozilla.org/hawk"
	"io"
	"net/http"
	"time"

	"crypto/sha256"
	"github.com/willf/bloom"
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
	var requestsPerHalfLife uint = 50000
	return &HawkHandler{
		handler:       handler,
		credentials:   secrets,
		bloomPrev:     bloom.New(requestsPerHalfLife, 5),
		bloomNow:      bloom.New(requestsPerHalfLife, 5),
		bloomHalflife: 30 * time.Second,
		lastRotate:    time.Now(),
	}
}

func (h *HawkHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	auth, err := hawk.NewAuthFromRequest(r, h.lookupCredentials, h.lookupNonce)
	if err != nil || auth.Valid() != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	hash := auth.PayloadHash(r.Header.Get("Content-Type"))
	io.Copy(hash, r.Body)
	if !auth.ValidHash(hash) {
		w.WriteHeader(http.StatusUnauthorized)
		return
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
