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

	SetViolationPenalties(testViolations)

	h := HandleWithMiddleware(NewRouter(), []Middleware{})
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
	SetViolationPenalties(nil)

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

	SetDB(db)
	SetMaxEntries(100)
	SetViolationPenalties(testViolations)

	h := HandleWithMiddleware(NewRouter(), []Middleware{})

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
		assert.Equal(t, http.StatusMovedPermanently, recorder.Code) // gorilla/mux redirect

		recorder = httptest.ResponseRecorder{}
		h.ServeHTTP(&recorder, httptest.NewRequest("PUT", "/violations////", strings.NewReader(`{"Violation": "UnknownViolation"}`)))
		assert.Equal(t, http.StatusMovedPermanently, recorder.Code) // gorilla/mux redirect
	})
}

func TestInsertReputationByViolationRequiresDB(t *testing.T) {
	testViolations := map[string]uint{
		"TestViolation": 90,
	}
	SetViolationPenalties(testViolations)

	SetDB(nil)

	h := HandleWithMiddleware(NewRouter(), []Middleware{})

	recorder := httptest.ResponseRecorder{}
	h.ServeHTTP(&recorder, httptest.NewRequest("PUT", "/violations/192.168.0.1", strings.NewReader(`{"Violation": "TestViolation"}`)))
	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
}


func TestMultiInsertReputationByViolation(t *testing.T) {
	dsn, found := os.LookupEnv("TIGERBLOOD_DSN")
	assert.True(t, found)
	db, err := NewDB(dsn)
	assert.Nil(t, err)
	err = db.EmptyTables()
	assert.Nil(t, err)

	testViolations := map[string]uint{
		"Test:Violation": 90,
	}

	SetDB(db)
	SetMaxEntries(3)
	SetViolationPenalties(testViolations)

	h := HandleWithMiddleware(NewRouter(), []Middleware{})

	t.Run("invalid json", func (t *testing.T) {
		recorder := httptest.ResponseRecorder{}
		h.ServeHTTP(&recorder, httptest.NewRequest("PUT", "/violations/", strings.NewReader(`[`)))
		assert.Equal(t, http.StatusBadRequest, recorder.Code)
	})

	t.Run("no ips", func (t *testing.T) {
		recorder := httptest.ResponseRecorder{}
		h.ServeHTTP(&recorder, httptest.NewRequest("PUT", "/violations/", strings.NewReader(`[]`)))
		assert.Equal(t, http.StatusBadRequest, recorder.Code)
	})

	t.Run("too many ips", func (t *testing.T) {
		recorder := httptest.ResponseRecorder{}
		h.ServeHTTP(&recorder, httptest.NewRequest("PUT", "/violations/", strings.NewReader(`[{"ip": "192.168.0.1", "Violation": "Test:Violation"}, {"ip": "192.168.1.1", "Violation": "Test:Violation"}, {"ip": "192.168.2.1", "Violation": "Test:Violation"}, {"ip": "192.168.3.1", "Violation": "Test:Violation"}]`)))
		assert.Equal(t, http.StatusBadRequest, recorder.Code)

		err = db.EmptyTables()
		assert.Nil(t, err)
	})

	t.Run("known violation single ip", func (t *testing.T) {
		// known violation type is subtracted from default reputation
		recorder := httptest.ResponseRecorder{}
		h.ServeHTTP(&recorder, httptest.NewRequest("PUT", "/violations/", strings.NewReader(`[{"ip": "192.168.0.1", "Violation": "Test:Violation"}]`)))
		assert.Equal(t, http.StatusNoContent, recorder.Code)

		entry, err := db.SelectSmallestMatchingSubnet("192.168.0.1")
		assert.Nil(t, err)
		assert.Equal(t, uint(10), entry.Reputation)

		err = db.EmptyTables()
		assert.Nil(t, err)
	})

	t.Run("known violation type many ips", func (t *testing.T) {
		recorder := httptest.ResponseRecorder{}
		h.ServeHTTP(&recorder, httptest.NewRequest("PUT", "/violations/", strings.NewReader(`[{"ip": "192.168.0.1", "Violation": "Test:Violation"}, {"ip": "192.168.0.100", "Violation": "Test:Violation"}]`)))
		assert.Equal(t, http.StatusNoContent, recorder.Code)

		entry, err := db.SelectSmallestMatchingSubnet("192.168.0.1")
		assert.Nil(t, err)
		assert.Equal(t, uint(10), entry.Reputation)

		entry, err = db.SelectSmallestMatchingSubnet("192.168.0.100")
		assert.Nil(t, err)
		assert.Equal(t, uint(10), entry.Reputation)

		err = db.EmptyTables()
		assert.Nil(t, err)
	})

	t.Run("known violation duplicate ip counted twice", func (t *testing.T) {
		// known violation type is subtracted from default reputation
		recorder := httptest.ResponseRecorder{}
		h.ServeHTTP(&recorder, httptest.NewRequest("PUT", "/violations/", strings.NewReader(`[{"ip": "192.168.0.1", "Violation": "Test:Violation"}, {"ip": "192.168.0.1", "Violation": "Test:Violation"}]`)))
		assert.Equal(t, http.StatusNoContent, recorder.Code)

		entry, err := db.SelectSmallestMatchingSubnet("192.168.0.1")
		assert.Nil(t, err)
		assert.Equal(t, uint(0), entry.Reputation)

		err = db.EmptyTables()
		assert.Nil(t, err)
	})

	t.Run("unknown violation type returns ip and index of first failure", func (t *testing.T) {
		recorder := httptest.NewRecorder()

		h := HandleWithMiddleware(NewRouter(), []Middleware{})
		req := httptest.NewRequest("PUT", "/violations/", strings.NewReader(`[{"ip": "192.168.0.1", "Violation": ""}]`))
		h.ServeHTTP(recorder, req)
		assert.Equal(t, http.StatusBadRequest, recorder.Code)

		res := recorder.Result()
		assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
		body, err := ioutil.ReadAll(res.Body)
		assert.Nil(t, err)

		assert.Equal(t, "{\"Errno\":19,\"EntryIndex\":0,\"Entry\":{\"Ip\":\"192.168.0.1\",\"Violation\":\"\"},\"Msg\":\"\"}", string(body))
	})

	t.Run("invalid violation type", func (t *testing.T) {
		// invalid violation type returns 400
		recorder := httptest.ResponseRecorder{}
		h.ServeHTTP(&recorder, httptest.NewRequest("PUT", "/violations/", strings.NewReader(`[{"ip": "192.168.0.1", "Violation": "InvalidViolation!"}]`)))
		assert.Equal(t, http.StatusBadRequest, recorder.Code)
	})
}


func TestMultiInsertReputationByViolationRequiresDB(t *testing.T) {
	testViolations := map[string]uint{
		"TestViolation": 90,
	}
	SetViolationPenalties(testViolations)

	SetDB(nil)
	SetMaxEntries(100)

	h := HandleWithMiddleware(NewRouter(), []Middleware{})

	recorder := httptest.ResponseRecorder{}
	h.ServeHTTP(&recorder, httptest.NewRequest("PUT", "/violations/", strings.NewReader(`[{"ip": "192.168.0.1", "Violation": "TestViolation"}]`)))
	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
}
