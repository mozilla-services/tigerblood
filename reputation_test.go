package tigerblood

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestReadReputationInvalidIP(t *testing.T) {
	recorder := httptest.ResponseRecorder{}
	dsn, found := os.LookupEnv("TIGERBLOOD_DSN")
	assert.True(t, found)
	db, err := NewDB(dsn)
	assert.Nil(t, err)
	err = db.CreateTables()
	assert.Nil(t, err)

	SetDB(db)
	h := HandleWithMiddleware(NewRouter(), []Middleware{})
	h.ServeHTTP(&recorder, httptest.NewRequest("GET", "/2472814.124981275", nil))
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

	SetDB(db)
	h := HandleWithMiddleware(NewRouter(), []Middleware{})
	h.ServeHTTP(&recorder, httptest.NewRequest("GET", "/127.0.0.1", nil))
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

	SetDB(db)
	h := HandleWithMiddleware(NewRouter(), []Middleware{})
	h.ServeHTTP(&recorder, httptest.NewRequest("GET", "/255.0.0.1", nil))
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

	SetDB(db)
	h := HandleWithMiddleware(NewRouter(), []Middleware{})

	h.ServeHTTP(&recorder, httptest.NewRequest("POST", "/", strings.NewReader(`{"IP": "192.168.0.1", "reputation": 20}`)))
	assert.Equal(t, http.StatusCreated, recorder.Code)
	assert.Nil(t, err)
	entry, err := db.SelectSmallestMatchingSubnet("192.168.0.1")
	assert.Nil(t, err)
	assert.Equal(t, uint(20), entry.Reputation)

	recorder = httptest.ResponseRecorder{}
	h.ServeHTTP(&recorder, httptest.NewRequest("POST", "/", strings.NewReader(`{"IP": "192.168.0.1", "reputation": 20}`)))
	assert.Equal(t, http.StatusConflict, recorder.Code)
}

func TestCreateEntryNoDB(t *testing.T) {
	recorder := httptest.ResponseRecorder{}
	SetDB(nil)
	h := HandleWithMiddleware(NewRouter(), []Middleware{})

	h.ServeHTTP(&recorder, httptest.NewRequest("POST", "/", strings.NewReader(`{"IP": "192.168.0.1", "reputation": 20}`)))
	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func TestCreateEntryInvalidJson(t *testing.T) {
	recorder := httptest.ResponseRecorder{}
	dsn, found := os.LookupEnv("TIGERBLOOD_DSN")
	assert.True(t, found)
	db, err := NewDB(dsn)
	assert.Nil(t, err)
	db.emptyReputationTable()

	SetDB(db)
	h := HandleWithMiddleware(NewRouter(), []Middleware{})
	h.ServeHTTP(&recorder, httptest.NewRequest("POST", "/", strings.NewReader(`{"IP": "192.168.0.1`)))
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Nil(t, err)
}

func TestCreateEntryInvalidIP(t *testing.T) {
	recorder := httptest.ResponseRecorder{}
	dsn, found := os.LookupEnv("TIGERBLOOD_DSN")
	assert.True(t, found)
	db, err := NewDB(dsn)
	assert.Nil(t, err)
	db.emptyReputationTable()

	SetDB(db)
	h := HandleWithMiddleware(NewRouter(), []Middleware{})
	h.ServeHTTP(&recorder, httptest.NewRequest("POST", "/", strings.NewReader(`{"IP": "192.168.0.1 -- SELECT(2)", "reputation": 200}`)))
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

	SetDB(db)
	h := HandleWithMiddleware(NewRouter(), []Middleware{})
	h.ServeHTTP(&recorder, httptest.NewRequest("POST", "/", strings.NewReader(`{"IP": "192.168.0.1", "reputation": 200}`)))
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

	SetDB(db)
	h := HandleWithMiddleware(NewRouter(), []Middleware{})
	h.ServeHTTP(&recorder, httptest.NewRequest("PUT", "/192.168.0.1", strings.NewReader(`{"IP": "192.168.0.1", "reputation": 25}`)))
	assert.Equal(t, http.StatusNotFound, recorder.Code)
	h.ServeHTTP(&recorder, httptest.NewRequest("POST", "/", strings.NewReader(`{"IP": "192.168.0.1", "reputation": 20}`)))
	recorder = httptest.ResponseRecorder{}
	h.ServeHTTP(&recorder, httptest.NewRequest("PUT", "/192.168.0.1", strings.NewReader(`{"IP": "192.168.0.1", "reputation": 25}`)))
	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Nil(t, err)
	entry, err := db.SelectSmallestMatchingSubnet("192.168.0.1")
	assert.Nil(t, err)
	assert.Equal(t, uint(25), entry.Reputation)
}

func TestUpdateEntryInvalidJson(t *testing.T) {
	recorder := httptest.ResponseRecorder{}
	dsn, found := os.LookupEnv("TIGERBLOOD_DSN")
	assert.True(t, found)
	db, err := NewDB(dsn)
	assert.Nil(t, err)
	db.emptyReputationTable()

	SetDB(db)
	h := HandleWithMiddleware(NewRouter(), []Middleware{})
	h.ServeHTTP(&recorder, httptest.NewRequest("PUT", "/192.168.0.1", strings.NewReader(`{"IP": `)))
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestUpdateEntryNoDB(t *testing.T) {
	recorder := httptest.ResponseRecorder{}

	SetDB(nil)
	h := HandleWithMiddleware(NewRouter(), []Middleware{})
	h.ServeHTTP(&recorder, httptest.NewRequest("PUT", "/192.168.0.1", strings.NewReader(`{"IP": "192.168.0.1", "reputation": 20}`)))
	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func TestDeleteEntry(t *testing.T) {
	recorder := httptest.ResponseRecorder{}
	dsn, found := os.LookupEnv("TIGERBLOOD_DSN")
	assert.True(t, found)
	db, err := NewDB(dsn)
	assert.Nil(t, err)
	db.emptyReputationTable()

	SetDB(db)
	h := HandleWithMiddleware(NewRouter(), []Middleware{})
	h.ServeHTTP(&recorder, httptest.NewRequest("POST", "/", strings.NewReader(`{"IP": "192.168.0.1", "reputation": 20}`)))
	recorder = httptest.ResponseRecorder{}
	h.ServeHTTP(&recorder, httptest.NewRequest("DELETE", "/192.168.0.1", nil))
	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Nil(t, err)
	_, err = db.SelectSmallestMatchingSubnet("192.168.0.1")
	assert.NotNil(t, err)
}
