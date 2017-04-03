package tigerblood

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"math/rand"
	"io/ioutil"
)

// http://stackoverflow.com/a/22892986
var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n uint) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func randViolations(numEntries int) map[string]uint {
	violations := make(map[string]uint)
	for i := 0; i < numEntries; i++ {
		violations[randSeq(256)] = uint(rand.Intn(101))
	}
	return violations
}

func TestSetViolationPenaltiesSkipsInvalidPenalties(t *testing.T) {
	testViolations := map[string]uint{
		"": 20,
		"TestViolation:2": 120,
	}

	SetHawkCreds(nil)
	SetViolationPenalties(testViolations)

	h := NewRouter()
	req := httptest.NewRequest("GET", "/violations", nil)
	recorder := httptest.NewRecorder()
	h.ServeHTTP(recorder, req)
	res := recorder.Result()

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))

	body, err := ioutil.ReadAll(res.Body)
	assert.Nil(t, err)

	assert.Equal(t, "{}", string(body))
}

func TestSetResponseHeadersMiddleware(t *testing.T) {
	h := NewRouter()
	req := httptest.NewRequest("GET", "/__version__", nil)
	recorder := httptest.NewRecorder()
	h.ServeHTTP(recorder, req)
	res := recorder.Result()

	assert.Equal(t, http.StatusOK, res.StatusCode)

	for _, header := range DefaultResponseHeaders {
		assert.Equal(t, res.Header.Get(header.Field), header.Value)
	}
}
