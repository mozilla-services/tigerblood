package tigerblood

import (
	"github.com/stretchr/testify/assert"
	"net/http/httptest"
	"os"
	"testing"
)

var cases = []struct {
	path string
	ip   string
	err  bool
}{
	{
		"/192.168.0.1",
		"192.168.0.1",
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
	ReadReputation(&recorder, httptest.NewRequest("GET", "/2472814.124981275", nil), db)
	assert.Equal(t, 400, recorder.Code)
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
	ReadReputation(&recorder, httptest.NewRequest("GET", "/127.0.0.1", nil), db)
	assert.Equal(t, 200, recorder.Code)
	assert.Nil(t, err)
}

func TestReadReputationNoEntry(t *testing.T) {
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
	ReadReputation(&recorder, httptest.NewRequest("GET", "/255.0.0.1", nil), db)
	assert.Equal(t, 404, recorder.Code)
	assert.Nil(t, err)
}
