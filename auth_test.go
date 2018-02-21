package tigerblood

import (
	"bytes"
	"crypto/sha256"
	"github.com/stretchr/testify/assert"
	"go.mozilla.org/hawk"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMissingAuthorizationAPIKey(t *testing.T) {
	req, err := http.NewRequest("GET", "http://foo.bar/", bytes.NewReader([]byte("foo")))
	assert.Nil(t, err)
	recorder := httptest.NewRecorder()
	credentials := make(map[string]string)
	SetAPIKeyCredentials(credentials)
	SetAuthMask(AuthEnableAPIKey)
	handler := HandleWithMiddleware(EchoHandler, []Middleware{RequireAuth()})
	handler.ServeHTTP(recorder, req)
	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
}

func TestInvalidAuthorizationAPIKeyEmptyConfig(t *testing.T) {
	req, err := http.NewRequest("GET", "http://foo.bar/", bytes.NewReader([]byte("foo")))
	assert.Nil(t, err)
	req.Header.Set("Authorization", "APIKey some_invalid_key")
	recorder := httptest.NewRecorder()
	credentials := make(map[string]string)
	SetAPIKeyCredentials(credentials)
	SetAuthMask(AuthEnableAPIKey)
	handler := HandleWithMiddleware(EchoHandler, []Middleware{RequireAuth()})
	handler.ServeHTTP(recorder, req)
	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
}

func TestInvalidAuthorizationAPIKeySetConfig(t *testing.T) {
	req, err := http.NewRequest("GET", "http://foo.bar/", bytes.NewReader([]byte("foo")))
	assert.Nil(t, err)
	req.Header.Set("Authorization", "APIKey some_invalid_key")
	recorder := httptest.NewRecorder()
	credentials := map[string]string{"test": "valid_key", "test2": "valid_key2"}
	SetAPIKeyCredentials(credentials)
	SetAuthMask(AuthEnableAPIKey)
	handler := HandleWithMiddleware(EchoHandler, []Middleware{RequireAuth()})
	handler.ServeHTTP(recorder, req)
	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
}

func TestValidAuthorizationAPIKey(t *testing.T) {
	req, err := http.NewRequest("GET", "http://foo.bar/", bytes.NewReader([]byte("foo")))
	assert.Nil(t, err)
	req.Header.Set("Authorization", "APIKey valid_key2")
	recorder := httptest.NewRecorder()
	credentials := map[string]string{"test": "valid_key", "test2": "valid_key2"}
	SetAPIKeyCredentials(credentials)
	SetAuthMask(AuthEnableAPIKey)
	handler := HandleWithMiddleware(EchoHandler, []Middleware{RequireAuth()})
	handler.ServeHTTP(recorder, req)
	assert.Equal(t, http.StatusOK, recorder.Code)
}

func TestLoadbalancerEndpointsUnauthedAPIKey(t *testing.T) {
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
		SetAPIKeyCredentials(credentials)
		SetAuthMask(AuthEnableAPIKey)
		handler := HandleWithMiddleware(EchoHandler, []Middleware{RequireAuth()})
		handler.ServeHTTP(recorder, req)
		assert.Equal(t, http.StatusOK, recorder.Code)
	}
}

func TestMissingMixedAuthorization(t *testing.T) {
	req, err := http.NewRequest("GET", "http://foo.bar/", bytes.NewReader([]byte("foo")))
	assert.Nil(t, err)
	recorder := httptest.NewRecorder()
	credentials := make(map[string]string)
	SetHawkCredentials(credentials)
	SetAPIKeyCredentials(credentials)
	SetAuthMask(AuthEnableHawk | AuthEnableAPIKey)
	handler := HandleWithMiddleware(EchoHandler, []Middleware{RequireAuth()})
	handler.ServeHTTP(recorder, req)
	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
}

func TestInvalidMixedAuthorizationEmptyConfig(t *testing.T) {
	req, err := http.NewRequest("GET", "http://foo.bar/", bytes.NewReader([]byte("foo")))
	assert.Nil(t, err)
	req.Header.Set("Authorization", "Hawk This is clearly not a hawk header")
	recorder := httptest.NewRecorder()
	credentials := make(map[string]string)
	SetHawkCredentials(credentials)
	SetAPIKeyCredentials(credentials)
	SetAuthMask(AuthEnableHawk | AuthEnableAPIKey)
	handler := HandleWithMiddleware(EchoHandler, []Middleware{RequireAuth()})
	handler.ServeHTTP(recorder, req)
	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
}

func TestInvalidMixedAuthorizationAPIKey(t *testing.T) {
	req, err := http.NewRequest("GET", "http://foo.bar/", bytes.NewReader([]byte("foo")))
	assert.Nil(t, err)
	req.Header.Set("Authorization", "APIKey invalid_key")
	recorder := httptest.NewRecorder()
	apicredentials := map[string]string{"test": "valid_key", "test2": "valid_key2"}
	hawkcredentials := map[string]string{"fxa": "foobar"}
	SetHawkCredentials(hawkcredentials)
	SetAPIKeyCredentials(apicredentials)
	SetAuthMask(AuthEnableHawk | AuthEnableAPIKey)
	handler := HandleWithMiddleware(EchoHandler, []Middleware{RequireAuth()})
	handler.ServeHTTP(recorder, req)
	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
}

func TestInvalidMixedAuthorizationHawk(t *testing.T) {
	req, err := http.NewRequest("GET", "http://foo.bar/", bytes.NewReader([]byte("foo")))
	assert.Nil(t, err)
	req.Header.Set("Content-Type", "application/json;charset=utf-8")
	auth := hawk.NewRequestAuth(req,
		&hawk.Credentials{
			ID:   "fxa",
			Key:  "invalid",
			Hash: sha256.New,
		},
		0,
	)
	hash := auth.PayloadHash("application/json")
	hash.Write([]byte("foo"))
	auth.SetHash(hash)
	req.Header.Set("Authorization", auth.RequestHeader())
	recorder := httptest.NewRecorder()
	apicredentials := map[string]string{"test": "valid_key", "test2": "valid_key2"}
	hawkcredentials := map[string]string{"fxa": "foobar"}
	SetHawkCredentials(hawkcredentials)
	SetAPIKeyCredentials(apicredentials)
	SetAuthMask(AuthEnableHawk | AuthEnableAPIKey)
	handler := HandleWithMiddleware(EchoHandler, []Middleware{RequireAuth()})
	handler.ServeHTTP(recorder, req)
	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
}

func TestValidMixedAuthorizationAPIKey(t *testing.T) {
	req, err := http.NewRequest("GET", "http://foo.bar/", bytes.NewReader([]byte("foo")))
	assert.Nil(t, err)
	req.Header.Set("Authorization", "APIKey valid_key2")
	recorder := httptest.NewRecorder()
	apicredentials := map[string]string{"test": "valid_key", "test2": "valid_key2"}
	hawkcredentials := map[string]string{"fxa": "foobar"}
	SetHawkCredentials(hawkcredentials)
	SetAPIKeyCredentials(apicredentials)
	SetAuthMask(AuthEnableHawk | AuthEnableAPIKey)
	handler := HandleWithMiddleware(EchoHandler, []Middleware{RequireAuth()})
	handler.ServeHTTP(recorder, req)
	assert.Equal(t, http.StatusOK, recorder.Code)
}

func TestValidMixedAuthorizationHawk(t *testing.T) {
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
	apicredentials := map[string]string{"test": "valid_key", "test2": "valid_key2"}
	hawkcredentials := map[string]string{"fxa": "foobar"}
	SetHawkCredentials(hawkcredentials)
	SetAPIKeyCredentials(apicredentials)
	SetAuthMask(AuthEnableHawk | AuthEnableAPIKey)
	handler := HandleWithMiddleware(EchoHandler, []Middleware{RequireAuth()})
	handler.ServeHTTP(recorder, req)
	assert.Equal(t, http.StatusOK, recorder.Code)
}

func TestInvalidAuthorizationAPIKeyEmpty(t *testing.T) {
	req, err := http.NewRequest("GET", "http://foo.bar/", bytes.NewReader([]byte("foo")))
	assert.Nil(t, err)
	req.Header.Set("Authorization", "APIKey ")
	recorder := httptest.NewRecorder()
	apicredentials := map[string]string{"test": "valid_key", "test2": "valid_key2"}
	hawkcredentials := map[string]string{"fxa": "foobar"}
	SetHawkCredentials(hawkcredentials)
	SetAPIKeyCredentials(apicredentials)
	SetAuthMask(AuthEnableHawk | AuthEnableAPIKey)
	handler := HandleWithMiddleware(EchoHandler, []Middleware{RequireAuth()})
	handler.ServeHTTP(recorder, req)
	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
}

func TestInvalidAuthorizationHawkEmpty(t *testing.T) {
	req, err := http.NewRequest("GET", "http://foo.bar/", bytes.NewReader([]byte("foo")))
	assert.Nil(t, err)
	req.Header.Set("Authorization", "Hawk ")
	recorder := httptest.NewRecorder()
	apicredentials := map[string]string{"test": "valid_key", "test2": "valid_key2"}
	hawkcredentials := map[string]string{"fxa": "foobar"}
	SetHawkCredentials(hawkcredentials)
	SetAPIKeyCredentials(apicredentials)
	SetAuthMask(AuthEnableHawk | AuthEnableAPIKey)
	handler := HandleWithMiddleware(EchoHandler, []Middleware{RequireAuth()})
	handler.ServeHTTP(recorder, req)
	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
}

func TestUnknownAuthorizationMethod(t *testing.T) {
	req, err := http.NewRequest("GET", "http://foo.bar/", bytes.NewReader([]byte("foo")))
	assert.Nil(t, err)
	// Valid key without correct prefix
	req.Header.Set("Authorization", "valid_key")
	recorder := httptest.NewRecorder()
	apicredentials := map[string]string{"test": "valid_key", "test2": "valid_key2"}
	hawkcredentials := map[string]string{"fxa": "foobar"}
	SetHawkCredentials(hawkcredentials)
	SetAPIKeyCredentials(apicredentials)
	SetAuthMask(AuthEnableHawk | AuthEnableAPIKey)
	handler := HandleWithMiddleware(EchoHandler, []Middleware{RequireAuth()})
	handler.ServeHTTP(recorder, req)
	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
}
