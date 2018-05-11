package tigerblood

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSetResponseHeadersMiddleware(t *testing.T) {
	h := HandleWithMiddleware(NewRouter(), []Middleware{SetResponseHeaders()})
	req := httptest.NewRequest("GET", "/__version__", nil)
	recorder := httptest.NewRecorder()
	h.ServeHTTP(recorder, req)
	res := recorder.Result()

	assert.Equal(t, http.StatusOK, res.StatusCode)

	for _, header := range DefaultResponseHeaders {
		assert.Equal(t, res.Header.Get(header.Field), header.Value)
	}
}
