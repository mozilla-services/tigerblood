package tigerblood

import (
	"bytes"
	"crypto/sha256"
	"github.com/stretchr/testify/assert"
	"go.mozilla.org/hawk"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

var EchoHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		io.Copy(w, r.Body)
	}
	w.WriteHeader(http.StatusOK)
})

func TestMissingAuthorization(t *testing.T) {
	req, err := http.NewRequest("GET", "http://foo.bar/", bytes.NewReader([]byte("foo")))
	assert.Nil(t, err)
	recorder := httptest.NewRecorder()
	handler := NewHawkHandler(EchoHandler, nil)
	handler.ServeHTTP(recorder, req)
	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
}

func TestInvalidAuthorization(t *testing.T) {
	req, err := http.NewRequest("GET", "http://foo.bar/", bytes.NewReader([]byte("foo")))
	assert.Nil(t, err)
	req.Header.Set("Authorization", "Hawk This is clearly not a hawk header")
	recorder := httptest.NewRecorder()
	handler := NewHawkHandler(EchoHandler, nil)
	handler.ServeHTTP(recorder, req)
	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
}

func TestInvalidPayload(t *testing.T) {
	req, err := http.NewRequest("GET", "http://foo.bar/", bytes.NewReader([]byte("foo")))
	assert.Nil(t, err)
	auth := hawk.NewRequestAuth(req,
		&hawk.Credentials{
			ID:   "fxa",
			Key:  "foobar",
			Hash: sha256.New,
		},
		0,
	)
	hash := auth.PayloadHash("application/json")
	hash.Write([]byte("foobar"))
	auth.SetHash(hash)
	req.Header.Set("Authorization", auth.RequestHeader())
	recorder := httptest.NewRecorder()
	handler := NewHawkHandler(EchoHandler, map[string]string{"fxa": "foobar"})
	handler.ServeHTTP(recorder, req)
	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
}

func TestValidPayload(t *testing.T) {
	req, err := http.NewRequest("GET", "http://foo.bar/", bytes.NewReader([]byte("foo")))
	assert.Nil(t, err)
	req.Header.Set("Content-Type", "application/json")
	auth := hawk.NewRequestAuth(req,
		&hawk.Credentials{
			ID:   "fxa",
			Key:  "foobar",
			Hash: sha256.New,
		},
		0,
	)
	hash := auth.PayloadHash("application/json")
	hash.Write([]byte("foo"))
	auth.SetHash(hash)
	req.Header.Set("Authorization", auth.RequestHeader())
	recorder := httptest.NewRecorder()
	handler := NewHawkHandler(EchoHandler, map[string]string{"fxa": "foobar"})
	handler.ServeHTTP(recorder, req)
	assert.Equal(t, http.StatusOK, recorder.Code)
}

func TestExpiration(t *testing.T) {
	req, err := http.NewRequest("GET", "http://foo.bar/", bytes.NewReader([]byte("foo")))
	assert.Nil(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", `Hawk id="fxa", mac="zcMu1EMcdseQ0J/LInTt73gHp3EiygoZnAC7KybGJBQ=", ts="1473887198", nonce="deYFZM4Z", hash="7wQDpR3QDtZYCfOpvQTEpR8cNz1dCX3sar9RLx5CmWk="`)
	recorder := httptest.NewRecorder()
	handler := NewHawkHandler(EchoHandler, map[string]string{"fxa": "foobar"})
	handler.ServeHTTP(recorder, req)
	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
}

func TestReplayProtection(t *testing.T) {
	req, err := http.NewRequest("GET", "http://foo.bar/", bytes.NewReader([]byte("foo")))
	assert.Nil(t, err)
	req.Header.Set("Content-Type", "application/json")
	auth := hawk.NewRequestAuth(req,
		&hawk.Credentials{
			ID:   "fxa",
			Key:  "foobar",
			Hash: sha256.New,
		},
		0,
	)
	hash := auth.PayloadHash("application/json")
	hash.Write([]byte("foo"))
	auth.SetHash(hash)
	req.Header.Set("Authorization", auth.RequestHeader())
	recorder := httptest.NewRecorder()
	handler := NewHawkHandler(EchoHandler, map[string]string{"fxa": "foobar"})
	handler.ServeHTTP(recorder, req)
	assert.Equal(t, http.StatusOK, recorder.Code)
	recorder = httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
}
