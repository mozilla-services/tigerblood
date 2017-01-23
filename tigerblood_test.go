package tigerblood

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

var cases = []struct {
	path string
	ip   string
	err  bool
}{
	{
		"/192.168.0.1",
		"192.168.0.1/32",
		false,
	},
	{
		"/192.168.0.1/32",
		"192.168.0.1/32",
		false,
	},
	{
		"/300.123.345.567",
		"",
		true,
	},
	{
		"/foobar",
		"",
		true,
	},
	{
		"/....",
		"",
		true,
	},
	{
		"/2001:0db8:0123:4567:89ab:cdef:1234:5678",
		"2001:db8:123:4567:89ab:cdef:1234:5678/128",
		false,
	},
	{
		"/2001:db8::ff00:42:8329",
		"2001:db8::ff00:42:8329/128",
		false,
	},
	{
		"/127.0.0.1' or '1' = '1",
		"",
		true,
	},
	{
		"/127.0.0.1; -- SELECT(2)",
		"",
		true,
	},
}

func TestIPAddressFromHTTPPath(t *testing.T) {
	for _, c := range cases {
		ip, err := IPAddressFromHTTPPath(c.path)
		if c.err {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}
		assert.Equal(t, c.ip, ip)
	}
}

func TestReadReputationInvalidIP(t *testing.T) {
	recorder := httptest.ResponseRecorder{}
	dsn, found := os.LookupEnv("TIGERBLOOD_DSN")
	assert.True(t, found)
	db, err := NewDB(dsn)
	assert.Nil(t, err)
	err = db.CreateTables()
	assert.Nil(t, err)
	h := NewTigerbloodHandler(db, nil, nil)
	h.ReadReputation(&recorder, httptest.NewRequest("GET", "/2472814.124981275", nil))
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestReadReputationValidIP(t *testing.T) {
	recorder := httptest.ResponseRecorder{}
	dsn, found := os.LookupEnv("TIGERBLOOD_DSN")
	assert.True(t, found)
	db, err := NewDB(dsn)
	assert.Nil(t, err)
	err = db.InsertOrUpdateReputationEntry(nil, ReputationEntry{
		IP:         "127.0.0.0/8",
		Reputation: 50,
	})
	assert.Nil(t, err)
	h := NewTigerbloodHandler(db, nil, nil)
	h.ReadReputation(&recorder, httptest.NewRequest("GET", "/127.0.0.1", nil))
	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Nil(t, err)
}

func TestReadReputationNoEntry(t *testing.T) {
	recorder := httptest.ResponseRecorder{}
	dsn, found := os.LookupEnv("TIGERBLOOD_DSN")
	assert.True(t, found)
	db, err := NewDB(dsn)
	assert.Nil(t, err)
	db.emptyReputationTable()
	err = db.InsertOrUpdateReputationEntry(nil, ReputationEntry{
		IP:         "127.0.0.0/8",
		Reputation: 50,
	})
	assert.Nil(t, err)
	h := NewTigerbloodHandler(db, nil, nil)
	h.ReadReputation(&recorder, httptest.NewRequest("GET", "/255.0.0.1", nil))
	assert.Equal(t, http.StatusNotFound, recorder.Code)
	assert.Nil(t, err)
}

func TestCreateEntry(t *testing.T) {
	recorder := httptest.ResponseRecorder{}
	dsn, found := os.LookupEnv("TIGERBLOOD_DSN")
	assert.True(t, found)
	db, err := NewDB(dsn)
	assert.Nil(t, err)
	db.emptyReputationTable()
	h := NewTigerbloodHandler(db, nil, nil)
	h.CreateReputation(&recorder, httptest.NewRequest("POST", "/", strings.NewReader(`{"IP": "192.168.0.1", "reputation": 20}`)))
	assert.Equal(t, http.StatusCreated, recorder.Code)
	assert.Nil(t, err)
	entry, err := db.SelectSmallestMatchingSubnet("192.168.0.1")
	assert.Nil(t, err)
	assert.Equal(t, uint(20), entry.Reputation)

	recorder = httptest.ResponseRecorder{}
	h.CreateReputation(&recorder, httptest.NewRequest("POST", "/", strings.NewReader(`{"IP": "192.168.0.1", "reputation": 20}`)))
	assert.Equal(t, http.StatusConflict, recorder.Code)
}

func TestCreateEntryInvalidIP(t *testing.T) {
	recorder := httptest.ResponseRecorder{}
	dsn, found := os.LookupEnv("TIGERBLOOD_DSN")
	assert.True(t, found)
	db, err := NewDB(dsn)
	assert.Nil(t, err)
	db.emptyReputationTable()
	h := NewTigerbloodHandler(db, nil, nil)
	h.CreateReputation(&recorder, httptest.NewRequest("POST", "/", strings.NewReader(`{"IP": "192.168.0.1 -- SELECT(2)", "reputation": 200}`)))
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Nil(t, err)
}

func TestCreateEntryInvalidReputation(t *testing.T) {
	recorder := httptest.ResponseRecorder{}
	dsn, found := os.LookupEnv("TIGERBLOOD_DSN")
	assert.True(t, found)
	db, err := NewDB(dsn)
	assert.Nil(t, err)
	db.emptyReputationTable()
	h := NewTigerbloodHandler(db, nil, nil)
	h.CreateReputation(&recorder, httptest.NewRequest("POST", "/", strings.NewReader(`{"IP": "192.168.0.1", "reputation": 200}`)))
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Nil(t, err)
}

func TestUpdateEntry(t *testing.T) {
	recorder := httptest.ResponseRecorder{}
	dsn, found := os.LookupEnv("TIGERBLOOD_DSN")
	assert.True(t, found)
	db, err := NewDB(dsn)
	assert.Nil(t, err)
	db.emptyReputationTable()
	h := NewTigerbloodHandler(db, nil, nil)
	h.UpdateReputation(&recorder, httptest.NewRequest("PUT", "/192.168.0.1", strings.NewReader(`{"IP": "192.168.0.1", "reputation": 25}`)))
	assert.Equal(t, http.StatusNotFound, recorder.Code)
	h.CreateReputation(&recorder, httptest.NewRequest("POST", "/", strings.NewReader(`{"IP": "192.168.0.1", "reputation": 20}`)))
	recorder = httptest.ResponseRecorder{}
	h.UpdateReputation(&recorder, httptest.NewRequest("PUT", "/192.168.0.1", strings.NewReader(`{"IP": "192.168.0.1", "reputation": 25}`)))
	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Nil(t, err)
	entry, err := db.SelectSmallestMatchingSubnet("192.168.0.1")
	assert.Nil(t, err)
	assert.Equal(t, uint(25), entry.Reputation)
}

func TestDeleteEntry(t *testing.T) {
	recorder := httptest.ResponseRecorder{}
	dsn, found := os.LookupEnv("TIGERBLOOD_DSN")
	assert.True(t, found)
	db, err := NewDB(dsn)
	assert.Nil(t, err)
	db.emptyReputationTable()
	h := NewTigerbloodHandler(db, nil, nil)
	h.CreateReputation(&recorder, httptest.NewRequest("POST", "/", strings.NewReader(`{"IP": "192.168.0.1", "reputation": 20}`)))
	recorder = httptest.ResponseRecorder{}
	h.DeleteReputation(&recorder, httptest.NewRequest("DELETE", "/192.168.0.1", nil))
	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Nil(t, err)
	_, err = db.SelectSmallestMatchingSubnet("192.168.0.1")
	assert.NotNil(t, err)
}

func TestInsertReputationByViolation(t *testing.T) {
	dsn, found := os.LookupEnv("TIGERBLOOD_DSN")
	assert.True(t, found)
	db, err := NewDB(dsn)
	assert.Nil(t, err)
	err = db.EmptyTables()
	assert.Nil(t, err)

	testViolations := map[string]uint{
		"TestViolation": 90,
	}

	h := NewTigerbloodHandler(db, nil, testViolations)

	// known violation type is subtracted from default reputation
	recorder := httptest.ResponseRecorder{}
	h.UpsertReputationByViolation(&recorder, httptest.NewRequest("PUT", "/violations/192.168.0.1", strings.NewReader(`{"Violation": "TestViolation"}`)))
	assert.Equal(t, http.StatusNoContent, recorder.Code)

	entry, err := db.SelectSmallestMatchingSubnet("192.168.0.1")
	assert.Nil(t, err)
	assert.Equal(t, uint(10), entry.Reputation)

	// unknown violation type returns 400
	recorder = httptest.ResponseRecorder{}
	h.UpsertReputationByViolation(&recorder, httptest.NewRequest("PUT", "/violations/192.168.0.1", strings.NewReader(`{"Violation": "UnknownViolation"}`)))
	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	entry, err = db.SelectSmallestMatchingSubnet("192.168.0.1")
	assert.Nil(t, err)
	assert.Equal(t, uint(10), entry.Reputation)

	// test parsing invalid URL
	recorder = httptest.ResponseRecorder{}
	h.UpsertReputationByViolation(&recorder, httptest.NewRequest("PUT", "/violations", strings.NewReader(`{"Violation": "UnknownViolation"}`)))
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	recorder = httptest.ResponseRecorder{}
	h.UpsertReputationByViolation(&recorder, httptest.NewRequest("PUT", "/violations////", strings.NewReader(`{"Violation": "UnknownViolation"}`)))
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}
