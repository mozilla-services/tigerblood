package tigerblood

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func TestListViolations(t *testing.T) {
	testViolations := map[string]uint{
		"TestViolation": 90,
		"TestViolation:2": 20,
	}

	h := HandleWithMiddleware(NewRouter(), []Middleware{AddViolations(testViolations)})
	req := httptest.NewRequest("GET", "/violations", nil)
	recorder := httptest.NewRecorder()
	h.ServeHTTP(recorder, req)
	res := recorder.Result()

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))

	body, err := ioutil.ReadAll(res.Body)
	assert.Nil(t, err)

	assert.Equal(t, "{\"TestViolation\":90,\"TestViolation:2\":20}", string(body))
}

func TestListViolationsMissingViolationsMiddleware(t *testing.T) {
	h := HandleWithMiddleware(NewRouter(), []Middleware{})
	req := httptest.NewRequest("GET", "/violations", nil)
	recorder := httptest.NewRecorder()
	h.ServeHTTP(recorder, req)
	res := recorder.Result()

	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
}

func TestInsertReputationByViolation(t *testing.T) {
	dsn, found := os.LookupEnv("TIGERBLOOD_DSN")
	assert.True(t, found)
	db, err := NewDB(dsn)
	assert.Nil(t, err)
	err = db.EmptyTables()
	assert.Nil(t, err)

	testViolations := map[string]uint{
		"Test:Violation": 90,
	}

	h := HandleWithMiddleware(NewRouter(), []Middleware{AddDB(db), AddViolations(testViolations)})

	t.Run("known", func (t *testing.T) {
		// known violation type is subtracted from default reputation
		recorder := httptest.ResponseRecorder{}
		h.ServeHTTP(&recorder, httptest.NewRequest("PUT", "/violations/192.168.0.1", strings.NewReader(`{"Violation": "Test:Violation"}`)))
		assert.Equal(t, http.StatusNoContent, recorder.Code)

		entry, err := db.SelectSmallestMatchingSubnet("192.168.0.1")
		assert.Nil(t, err)
		assert.Equal(t, uint(10), entry.Reputation)
	})

	t.Run("unknown", func (t *testing.T) {
		// unknown violation type returns 400
		recorder := httptest.ResponseRecorder{}
		h.ServeHTTP(&recorder, httptest.NewRequest("PUT", "/violations/192.168.0.1", strings.NewReader(`{"Violation": "UnknownViolation"}`)))
		assert.Equal(t, http.StatusBadRequest, recorder.Code)
	})

	t.Run("invalid", func (t *testing.T) {
		// invalid violation type returns 400
		recorder := httptest.ResponseRecorder{}
		h.ServeHTTP(&recorder, httptest.NewRequest("PUT", "/violations/192.168.0.1", strings.NewReader(`{"Violation": "UnknownViolation!"}`)))
		assert.Equal(t, http.StatusBadRequest, recorder.Code)
	})

	t.Run("invalid-urls", func (t *testing.T) {
		// test parsing invalid URL
		recorder := httptest.ResponseRecorder{}
		h.ServeHTTP(&recorder, httptest.NewRequest("PUT", "/violations", strings.NewReader(`{"Violation": "UnknownViolation"}`)))
		assert.Equal(t, http.StatusBadRequest, recorder.Code)
		recorder = httptest.ResponseRecorder{}
		h.ServeHTTP(&recorder, httptest.NewRequest("PUT", "/violations////", strings.NewReader(`{"Violation": "UnknownViolation"}`)))
		assert.Equal(t, http.StatusMovedPermanently, recorder.Code) // gorilla/mux redirect
	})
}
