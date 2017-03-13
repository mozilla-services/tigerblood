package tigerblood

import (
	"github.com/DataDog/datadog-go/statsd"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"math/rand"
	"io/ioutil"
)



func TestDurationTrackingMiddlewareWithStatsd(t *testing.T) {
	statsdClient, err := statsd.New("127.0.0.1:39209")
	assert.Nil(t, err)
	statsdClient.Namespace = "TestDurationTrackingMiddlewareWithStatsd."

	h := HandleWithMiddleware(NewRouter(), []Middleware{RecordStartTime(), AddStatsdClient(statsdClient), LogRequestDuration(1e7)})
	req := httptest.NewRequest("GET", "/__version__", nil)
	recorder := httptest.NewRecorder()
	h.ServeHTTP(recorder, req)
	res := recorder.Result()

	assert.Equal(t, http.StatusOK, res.StatusCode)

	// TODO: assert statsd receives data
}


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


func TestDurationTrackingMiddlewareLogsSlowRequest(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	testViolations := randViolations(9001)
	slowRequestToleranceNs := 100
	h := HandleWithMiddleware(NewRouter(),
		[]Middleware{RecordStartTime(),
			AddViolations(testViolations),
			LogRequestDuration(slowRequestToleranceNs)})
	start := time.Now()

	req := httptest.NewRequest("GET", "/violations", nil)
	recorder := httptest.NewRecorder()
	h.ServeHTTP(recorder, req)
	res := recorder.Result()

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.True(t, time.Since(start).Nanoseconds() > int64(slowRequestToleranceNs))
}

func TestAddViolationsMiddlewareSkipsInvalidPenalties(t *testing.T) {
	testViolations := map[string]uint{
		"": 20,
		"TestViolation:2": 120,
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

	assert.Equal(t, "{}", string(body))
}

func TestDurationTrackingMiddlewareWithoutStatsd(t *testing.T) {
	h := HandleWithMiddleware(NewRouter(), []Middleware{RecordStartTime(), LogRequestDuration(1e7)})
	req := httptest.NewRequest("GET", "/__version__", nil)
	recorder := httptest.NewRecorder()
	h.ServeHTTP(recorder, req)
	res := recorder.Result()

	assert.Equal(t, http.StatusOK, res.StatusCode)
}

func TestDurationTrackingMiddlewareWithoutStatsdAndStartTime(t *testing.T) {
	// should never be configured this way, but also shouldn't explode

	h := HandleWithMiddleware(NewRouter(), []Middleware{LogRequestDuration(1e7)})
	req := httptest.NewRequest("GET", "/__version__", nil)
	recorder := httptest.NewRecorder()
	h.ServeHTTP(recorder, req)
	res := recorder.Result()

	assert.Equal(t, http.StatusOK, res.StatusCode)
}

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
