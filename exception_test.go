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

func TestFileExceptionSource(t *testing.T) {
	dsn, found := os.LookupEnv("TIGERBLOOD_DSN")
	assert.True(t, found)
	db, err := NewDB(dsn)
	assert.Nil(t, err)
	SetDB(db)
	err = db.EmptyTables()
	assert.Nil(t, err)
	assert.Nil(t, AddFileException("testdata/exceptions.txt"))
	assert.Nil(t, InitializeExceptions())
	ret, err := testDB.SelectExceptionsContaining("10.20.0.50")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(ret))
	assert.Equal(t, ret[0].Creator, "file:testdata/exceptions.txt")
	ret, err = testDB.SelectExceptionsContaining("172.16.0.4")
	assert.Nil(t, err)
	assert.Equal(t, 0, len(ret))
	ret, err = testDB.SelectExceptionsContaining("192.168.51.200")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(ret))
	assert.Equal(t, ret[0].Creator, "file:testdata/exceptions.txt")

	assert.Nil(t, db.Close())
}

func TestListExceptions(t *testing.T) {
	dsn, found := os.LookupEnv("TIGERBLOOD_DSN")
	assert.True(t, found)
	db, err := NewDB(dsn)
	assert.Nil(t, err)
	db.EmptyTables()

	var entries []ExceptionEntry

	SetDB(db)
	h := HandleWithMiddleware(NewRouter(), []Middleware{})

	recorder := httptest.NewRecorder()
	h.ServeHTTP(recorder, httptest.NewRequest("GET", "/exceptions", nil))
	res := recorder.Result()
	assert.Equal(t, http.StatusNotFound, res.StatusCode)

	assert.Nil(t, db.InsertOrUpdateExceptionEntry(nil, ExceptionEntry{
		IP:      "10.0.5.0/24",
		Creator: "file:/test",
	}))

	recorder = httptest.NewRecorder()
	h.ServeHTTP(recorder, httptest.NewRequest("GET", "/exceptions", nil))
	res = recorder.Result()
	assert.Equal(t, http.StatusOK, res.StatusCode)
	body, err := ioutil.ReadAll(res.Body)
	assert.Nil(t, err)
	err = json.Unmarshal(body, &entries)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(entries))

	assert.Nil(t, db.InsertOrUpdateExceptionEntry(nil, ExceptionEntry{
		IP:      "10.0.6.0/24",
		Creator: "file:/test",
	}))

	recorder = httptest.NewRecorder()
	h.ServeHTTP(recorder, httptest.NewRequest("GET", "/exceptions", nil))
	res = recorder.Result()
	assert.Equal(t, http.StatusOK, res.StatusCode)
	body, err = ioutil.ReadAll(res.Body)
	assert.Nil(t, err)
	err = json.Unmarshal(body, &entries)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(entries))
}

func TestExceptionApplyOnWriteSingle(t *testing.T) {
	recorder := httptest.ResponseRecorder{}
	dsn, found := os.LookupEnv("TIGERBLOOD_DSN")
	assert.True(t, found)
	db, err := NewDB(dsn)
	assert.Nil(t, err)
	db.EmptyTables()

	SetDB(db)
	h := HandleWithMiddleware(NewRouter(), []Middleware{})

	h.ServeHTTP(&recorder, httptest.NewRequest("PUT", "/192.168.0.1", strings.NewReader(`{"IP": "192.168.0.1", "reputation": 20}`)))
	assert.Equal(t, http.StatusOK, recorder.Code)
	recorder = httptest.ResponseRecorder{}
	h.ServeHTTP(&recorder, httptest.NewRequest("PUT", "/192.168.1.1", strings.NewReader(`{"IP": "192.168.1.1", "reputation": 20}`)))
	assert.Equal(t, http.StatusOK, recorder.Code)
	entry, err := testDB.SelectSmallestMatchingSubnet("192.168.0.1")
	assert.Nil(t, err)
	assert.Equal(t, uint(20), entry.Reputation)
	entry, err = testDB.SelectSmallestMatchingSubnet("192.168.1.1")
	assert.Nil(t, err)
	assert.Equal(t, uint(20), entry.Reputation)

	db.EmptyTables()

	assert.Nil(t, db.InsertOrUpdateExceptionEntry(nil, ExceptionEntry{
		IP:      "192.168.0.0/24",
		Creator: "file:/test",
	}))
	recorder = httptest.ResponseRecorder{}
	h.ServeHTTP(&recorder, httptest.NewRequest("PUT", "/192.168.0.1", strings.NewReader(`{"IP": "192.168.0.1", "reputation": 20}`)))
	assert.Equal(t, http.StatusOK, recorder.Code)
	recorder = httptest.ResponseRecorder{}
	h.ServeHTTP(&recorder, httptest.NewRequest("PUT", "/192.168.1.1", strings.NewReader(`{"IP": "192.168.1.1", "reputation": 20}`)))
	assert.Equal(t, http.StatusOK, recorder.Code)
	entry, err = testDB.SelectSmallestMatchingSubnet("192.168.0.1")
	assert.NotNil(t, err)
	entry, err = testDB.SelectSmallestMatchingSubnet("192.168.1.1")
	assert.Nil(t, err)
	assert.Equal(t, uint(20), entry.Reputation)

	assert.Nil(t, db.Close())
}

func TestExceptionApplyOnWriteViolationMulti(t *testing.T) {
	recorder := httptest.ResponseRecorder{}
	dsn, found := os.LookupEnv("TIGERBLOOD_DSN")
	assert.True(t, found)
	db, err := NewDB(dsn)
	assert.Nil(t, err)
	db.EmptyTables()

	testViolations := map[string]uint{
		"Test:Violation":  90,
		"Test:Violation2": 10,
	}

	SetDB(db)
	SetViolationPenalties(testViolations)
	h := HandleWithMiddleware(NewRouter(), []Middleware{})

	h.ServeHTTP(&recorder, httptest.NewRequest("PUT", "/violations/", strings.NewReader(`[{"ip": "192.168.0.1", "violation": "Test:Violation"}, {"ip": "10.20.20.20", "violation": "Test:Violation2"}]`)))
	assert.Equal(t, http.StatusNoContent, recorder.Code)
	entry, err := testDB.SelectSmallestMatchingSubnet("192.168.0.1")
	assert.Nil(t, err)
	assert.Equal(t, uint(10), entry.Reputation)
	entry, err = testDB.SelectSmallestMatchingSubnet("10.20.20.20")
	assert.Nil(t, err)
	assert.Equal(t, uint(90), entry.Reputation)

	db.EmptyTables()

	assert.Nil(t, db.InsertOrUpdateExceptionEntry(nil, ExceptionEntry{
		IP:      "10.20.0.0/16",
		Creator: "file:/test",
	}))
	recorder = httptest.ResponseRecorder{}
	h.ServeHTTP(&recorder, httptest.NewRequest("PUT", "/violations/", strings.NewReader(`[{"ip": "192.168.0.1", "violation": "Test:Violation"}, {"ip": "10.20.20.20", "violation": "Test:Violation2"}]`)))
	assert.Equal(t, http.StatusNoContent, recorder.Code)
	entry, err = testDB.SelectSmallestMatchingSubnet("192.168.0.1")
	assert.Nil(t, err)
	assert.Equal(t, uint(10), entry.Reputation)
	entry, err = testDB.SelectSmallestMatchingSubnet("10.20.20.20")
	assert.NotNil(t, err)

	assert.Nil(t, db.Close())
}

func TestExceptionApplyOnReadSingle(t *testing.T) {
	recorder := httptest.ResponseRecorder{}
	dsn, found := os.LookupEnv("TIGERBLOOD_DSN")
	assert.True(t, found)
	db, err := NewDB(dsn)
	assert.Nil(t, err)
	db.EmptyTables()

	SetDB(db)
	h := HandleWithMiddleware(NewRouter(), []Middleware{})

	h.ServeHTTP(&recorder, httptest.NewRequest("PUT", "/192.168.0.1", strings.NewReader(`{"IP": "192.168.0.1", "reputation": 20}`)))
	assert.Equal(t, http.StatusOK, recorder.Code)
	entry, err := db.SelectSmallestMatchingSubnet("192.168.0.1")
	assert.Nil(t, err)
	assert.Equal(t, uint(20), entry.Reputation)

	recorder = httptest.ResponseRecorder{}
	h.ServeHTTP(&recorder, httptest.NewRequest("GET", "/192.168.0.1", nil))
	assert.Equal(t, http.StatusOK, recorder.Code)

	assert.Nil(t, db.InsertOrUpdateExceptionEntry(nil, ExceptionEntry{
		IP:      "192.168.0.0/29",
		Creator: "file:/test",
	}))
	recorder = httptest.ResponseRecorder{}
	h.ServeHTTP(&recorder, httptest.NewRequest("GET", "/192.168.0.1", nil))
	assert.Equal(t, http.StatusNotFound, recorder.Code)

	assert.Nil(t, db.Close())
}
