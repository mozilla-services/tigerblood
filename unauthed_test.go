package tigerblood

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestLoadBalancerHeartbeatHandler(t *testing.T) {
	h := HandleWithMiddleware(NewRouter(), []Middleware{})
	req := httptest.NewRequest("GET", "/__lbheartbeat__", nil)
	recorder := httptest.NewRecorder()
	h.ServeHTTP(recorder, req)
	res := recorder.Result()

	assert.Equal(t, http.StatusOK, res.StatusCode)
}

func TestHeartbeatHandler(t *testing.T) {
	dsn, found := os.LookupEnv("TIGERBLOOD_DSN")
	assert.True(t, found)
	db, err := NewDB(dsn)
	assert.Nil(t, err)

	SetDB(db)
	h := HandleWithMiddleware(NewRouter(), []Middleware{})
	req := httptest.NewRequest("GET", "/__heartbeat__", nil)
	recorder := httptest.NewRecorder()
	h.ServeHTTP(recorder, req)
	res := recorder.Result()

	assert.Equal(t, http.StatusOK, res.StatusCode)
}

func TestHeartbeatHandlerWithoutDB(t *testing.T) {
	SetDB(nil)

	h := HandleWithMiddleware(NewRouter(), []Middleware{})
	req := httptest.NewRequest("GET", "/__heartbeat__", nil)
	recorder := httptest.NewRecorder()
	h.ServeHTTP(recorder, req)
	res := recorder.Result()

	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
}

func TestVersionHandler(t *testing.T) {
	h := HandleWithMiddleware(NewRouter(), []Middleware{})
	req := httptest.NewRequest("GET", "/__version__", nil)
	recorder := httptest.NewRecorder()
	h.ServeHTTP(recorder, req)
	res := recorder.Result()

	assert.Equal(t, http.StatusOK, res.StatusCode)
}

func TestDebugRoutesWhenProfileHandlersEnabled(t *testing.T) {
	SetProfileHandlers(true)

	h := HandleWithMiddleware(NewRouter(), []Middleware{})
	req := httptest.NewRequest("GET", "/debug/pprof/", nil)
	recorder := httptest.NewRecorder()
	h.ServeHTTP(recorder, req)
	res := recorder.Result()

	assert.Equal(t, http.StatusOK, res.StatusCode)
}

func TestDebugRoutesWhenProfileHandlersDisabled(t *testing.T) {
	SetProfileHandlers(false)

	h := HandleWithMiddleware(NewRouter(), []Middleware{})
	req := httptest.NewRequest("GET", "/debug/pprof/", nil)
	recorder := httptest.NewRecorder()
	h.ServeHTTP(recorder, req)
	res := recorder.Result()

	assert.Equal(t, http.StatusMovedPermanently, res.StatusCode)
}
