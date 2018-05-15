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
	credentials := make(map[string]string)
	SetHawkCredentials(credentials)
	SetAuthMask(AuthEnableHawk)
	handler := HandleWithMiddleware(EchoHandler, []Middleware{RequireAuth()})
	handler.ServeHTTP(recorder, req)
	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
}

func TestInvalidAuthorization(t *testing.T) {
	req, err := http.NewRequest("GET", "http://foo.bar/", bytes.NewReader([]byte("foo")))
	assert.Nil(t, err)
	req.Header.Set("Authorization", "Hawk This is clearly not a hawk header")
	recorder := httptest.NewRecorder()
	credentials := make(map[string]string)
	SetHawkCredentials(credentials)
	SetAuthMask(AuthEnableHawk)
	handler := HandleWithMiddleware(EchoHandler, []Middleware{RequireAuth()})
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
	credentials := map[string]string{"fxa": "foobar"}
	SetHawkCredentials(credentials)
	SetAuthMask(AuthEnableHawk)
	handler := HandleWithMiddleware(EchoHandler, []Middleware{RequireAuth()})
	handler.ServeHTTP(recorder, req)
	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
}

func TestValidPayload(t *testing.T) {
	req, err := http.NewRequest("GET", "http://foo.bar/", bytes.NewReader([]byte("foo")))
	assert.Nil(t, err)
	req.Header.Set("Content-Type", "application/json;charset=utf-8")
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
	credentials := map[string]string{"fxa": "foobar"}
	SetHawkCredentials(credentials)
	SetAuthMask(AuthEnableHawk)
	handler := HandleWithMiddleware(EchoHandler, []Middleware{RequireAuth()})
	handler.ServeHTTP(recorder, req)
	assert.Equal(t, http.StatusOK, recorder.Code)
}

func TestValidPayloadNoContentType(t *testing.T) {
	// use a POST to hit log the missing content type warning
	req, err := http.NewRequest("POST", "http://foo.bar/", bytes.NewReader([]byte("foo")))
	assert.Nil(t, err)
	req.Header.Set("Content-Type", "")
	auth := hawk.NewRequestAuth(req,
		&hawk.Credentials{
			ID:   "fxa",
			Key:  "foobar",
			Hash: sha256.New,
		},
		0,
	)
	hash := auth.PayloadHash("")
	hash.Write([]byte("foo"))
	auth.SetHash(hash)
	req.Header.Set("Authorization", auth.RequestHeader())
	recorder := httptest.NewRecorder()
	credentials := map[string]string{"fxa": "foobar"}
	SetHawkCredentials(credentials)
	SetAuthMask(AuthEnableHawk)
	handler := HandleWithMiddleware(EchoHandler, []Middleware{RequireAuth()})
	handler.ServeHTTP(recorder, req)
	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
}

func TestExpiration(t *testing.T) {
	req, err := http.NewRequest("GET", "http://foo.bar/", bytes.NewReader([]byte("foo")))
	assert.Nil(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", `Hawk id="fxa", mac="zcMu1EMcdseQ0J/LInTt73gHp3EiygoZnAC7KybGJBQ=", ts="1473887198", nonce="deYFZM4Z", hash="7wQDpR3QDtZYCfOpvQTEpR8cNz1dCX3sar9RLx5CmWk="`)
	recorder := httptest.NewRecorder()
	credentials := map[string]string{"fxa": "foobar"}
	SetHawkCredentials(credentials)
	handler := HandleWithMiddleware(EchoHandler, []Middleware{RequireAuth()})
	handler.ServeHTTP(recorder, req)
	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
}

func TestLoadbalancerEndpointsUnauthed(t *testing.T) {
	SetProfileHandlers(true)

	for _, path := range []string{
		"/__lbheartbeat__",
		"/__heartbeat__",
		"/__version__",
		"/debug/pprof/",
		"/debug/pprof/cmdline",
		"/debug/pprof/profile",
		"/debug/pprof/symbol",
	} {
		req, err := http.NewRequest("GET", "http://foo.bar"+path, nil)
		assert.Nil(t, err)
		recorder := httptest.NewRecorder()
		credentials := map[string]string{"fxa": "foobar"}
		SetHawkCredentials(credentials)
		SetAuthMask(AuthEnableHawk)
		handler := HandleWithMiddleware(EchoHandler, []Middleware{RequireAuth()})
		handler.ServeHTTP(recorder, req)
		assert.Equal(t, http.StatusOK, recorder.Code)
	}
}

func TestMissingCredentialsReturns401(t *testing.T) {
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
	credentials := map[string]string{"notFxa": "foobar"}
	SetHawkCredentials(credentials)
	SetAuthMask(AuthEnableHawk)
	handler := HandleWithMiddleware(EchoHandler, []Middleware{RequireAuth()})
	handler.ServeHTTP(recorder, req)
	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
}
