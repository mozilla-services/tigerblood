package tigerblood

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
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

	assert.Nil(t, db.Close())
}

func TestReadReputationValidIP(t *testing.T) {
	recorder := httptest.ResponseRecorder{}
	dsn, found := os.LookupEnv("TIGERBLOOD_DSN")
	assert.True(t, found)
	db, err := NewDB(dsn)
	assert.Nil(t, err)
	_, err = db.InsertOrUpdateReputationEntry(nil, ReputationEntry{
		IP:         "127.0.0.0/8",
		Reputation: 50,
	})
	assert.Nil(t, err)

	SetDB(db)
	h := HandleWithMiddleware(NewRouter(), []Middleware{})
	h.ServeHTTP(&recorder, httptest.NewRequest("GET", "/127.0.0.1", nil))
	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Nil(t, err)

	assert.Nil(t, db.Close())
}

func TestReadReputationNoEntry(t *testing.T) {
	recorder := httptest.ResponseRecorder{}
	dsn, found := os.LookupEnv("TIGERBLOOD_DSN")
	assert.True(t, found)
	db, err := NewDB(dsn)
	assert.Nil(t, err)
	db.EmptyTables()

	_, err = db.InsertOrUpdateReputationEntry(nil, ReputationEntry{
		IP:         "127.0.0.0/8",
		Reputation: 50,
	})
	assert.Nil(t, err)

	SetDB(db)
	h := HandleWithMiddleware(NewRouter(), []Middleware{})
	h.ServeHTTP(&recorder, httptest.NewRequest("GET", "/255.0.0.1", nil))
	assert.Equal(t, http.StatusNotFound, recorder.Code)
	assert.Nil(t, err)

	assert.Nil(t, db.Close())
}

func TestUpdateEntry(t *testing.T) {
	recorder := httptest.ResponseRecorder{}
	dsn, found := os.LookupEnv("TIGERBLOOD_DSN")
	assert.True(t, found)
	db, err := NewDB(dsn)
	assert.Nil(t, err)
	db.EmptyTables()

	SetDB(db)
	h := HandleWithMiddleware(NewRouter(), []Middleware{})
	h.ServeHTTP(&recorder, httptest.NewRequest("PUT", "/192.168.0.1",
		strings.NewReader(`{"IP": "192.168.0.1", "reputation": 25}`)))
	assert.Equal(t, http.StatusOK, recorder.Code)
	entry, err := db.SelectSmallestMatchingSubnet("192.168.0.1")
	assert.Nil(t, err)
	assert.Equal(t, uint(25), entry.Reputation)

	recorder = httptest.ResponseRecorder{}
	h.ServeHTTP(&recorder, httptest.NewRequest("PUT", "/192.168.0.1", strings.NewReader(`{"IP": "192.168.0.1", "reputation": 50}`)))
	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Nil(t, err)
	entry, err = db.SelectSmallestMatchingSubnet("192.168.0.1")
	assert.Nil(t, err)
	assert.Equal(t, uint(50), entry.Reputation)

	assert.Nil(t, db.Close())
}

func TestUpdateEntryInvalidJson(t *testing.T) {
	recorder := httptest.ResponseRecorder{}
	dsn, found := os.LookupEnv("TIGERBLOOD_DSN")
	assert.True(t, found)
	db, err := NewDB(dsn)
	assert.Nil(t, err)
	db.EmptyTables()

	SetDB(db)
	h := HandleWithMiddleware(NewRouter(), []Middleware{})
	h.ServeHTTP(&recorder, httptest.NewRequest("PUT", "/192.168.0.1", strings.NewReader(`{"IP": `)))
	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	assert.Nil(t, db.Close())
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
	db.EmptyTables()

	SetDB(db)
	h := HandleWithMiddleware(NewRouter(), []Middleware{})
	h.ServeHTTP(&recorder, httptest.NewRequest("POST", "/", strings.NewReader(`{"IP": "192.168.0.1", "reputation": 20}`)))
	recorder = httptest.ResponseRecorder{}
	h.ServeHTTP(&recorder, httptest.NewRequest("DELETE", "/192.168.0.1", nil))
	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Nil(t, err)
	_, err = db.SelectSmallestMatchingSubnet("192.168.0.1")
	assert.NotNil(t, err)

	assert.Nil(t, db.Close())
}

func TestDeleteEntryNoDB(t *testing.T) {
	recorder := httptest.ResponseRecorder{}

	h := HandleWithMiddleware(NewRouter(), []Middleware{})

	SetDB(nil)
	h.ServeHTTP(&recorder, httptest.NewRequest("DELETE", "/192.168.0.1", nil))
	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func TestReadReputationReviewed(t *testing.T) {
	dsn, found := os.LookupEnv("TIGERBLOOD_DSN")
	assert.True(t, found)
	db, err := NewDB(dsn)
	assert.Nil(t, err)
	db.EmptyTables()

	SetDB(db)
	h := HandleWithMiddleware(NewRouter(), []Middleware{})

	recorder := httptest.NewRecorder()
	h.ServeHTTP(recorder, httptest.NewRequest("PUT", "/192.168.0.1", strings.NewReader(`{"IP": "192.168.0.1", "reputation": 20}`)))
	assert.Equal(t, http.StatusOK, recorder.Code)
	entry, err := db.SelectSmallestMatchingSubnet("192.168.0.1")
	assert.Nil(t, err)
	assert.Equal(t, uint(20), entry.Reputation)
	assert.Equal(t, false, entry.Reviewed)

	var ent ReputationEntry
	recorder = httptest.NewRecorder()
	h.ServeHTTP(recorder, httptest.NewRequest("GET", "/192.168.0.1", nil))
	res := recorder.Result()
	assert.Equal(t, http.StatusOK, recorder.Code)
	body, err := ioutil.ReadAll(res.Body)
	assert.Nil(t, err)
	err = json.Unmarshal(body, &ent)
	assert.Nil(t, err)
	assert.Equal(t, uint(20), ent.Reputation)
	assert.Equal(t, false, ent.Reviewed)

	assert.Nil(t, testDB.SetReviewedFlag(nil, ent, true))

	recorder = httptest.NewRecorder()
	h.ServeHTTP(recorder, httptest.NewRequest("GET", "/192.168.0.1", nil))
	res = recorder.Result()
	assert.Equal(t, http.StatusOK, recorder.Code)
	body, err = ioutil.ReadAll(res.Body)
	assert.Nil(t, err)
	err = json.Unmarshal(body, &ent)
	assert.Nil(t, err)
	assert.Equal(t, uint(20), ent.Reputation)
	assert.Equal(t, true, ent.Reviewed)

	recorder = httptest.NewRecorder()
	h.ServeHTTP(recorder, httptest.NewRequest("PUT", "/192.168.0.1", strings.NewReader(`{"IP": "192.168.0.1", "reputation": 100}`)))
	assert.Equal(t, http.StatusOK, recorder.Code)

	recorder = httptest.NewRecorder()
	h.ServeHTTP(recorder, httptest.NewRequest("GET", "/192.168.0.1", nil))
	res = recorder.Result()
	assert.Equal(t, http.StatusOK, recorder.Code)
	body, err = ioutil.ReadAll(res.Body)
	assert.Nil(t, err)
	err = json.Unmarshal(body, &ent)
	assert.Nil(t, err)
	assert.Equal(t, uint(100), ent.Reputation)
	assert.Equal(t, false, ent.Reviewed)

	assert.Nil(t, db.Close())
}
