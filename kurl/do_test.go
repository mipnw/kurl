package kurl_test

import (
	"github.com/mipnw/kurl/kurl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestStatus200(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "GET", req.Method)
		assert.Equal(t, "value", req.Header.Get("key"))
		rw.Write([]byte(`OK`))
	}))
	defer server.Close()

	expectedRequest, err := http.NewRequest("GET", server.URL, nil)
	require.Nil(t, err)
	expectedRequest.Header.Add("key", "value")

	settings := kurl.Settings{
		ThreadCount:  10,
		RequestCount: 10,
	}

	result, err := kurl.Do(
		settings,
		*expectedRequest,
	)
	assert.Nil(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 0, result.ErrorCount)
	assert.Equal(t, settings.ThreadCount*settings.RequestCount, result.CompletedCount)
	assert.Equal(t, result.CompletedCount, result.StatusCodesFrequency[http.StatusOK])
}

func TestStatusCode429(t *testing.T) {
	var requestCount uint64
	requestCount = 0
	mux := sync.Mutex{}

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "POST", req.Method)

		var ok bool
		mux.Lock()
		ok = (requestCount%2 == 0)
		requestCount++
		mux.Unlock()

		if ok {
			rw.Write([]byte(`OK`))
		} else {
			http.Error(rw, `TOO MANY REQUESTS`, http.StatusTooManyRequests)
		}
	}))
	defer server.Close()

	expectedRequest, err := http.NewRequest("POST", server.URL, nil)
	require.Nil(t, err)

	settings := kurl.Settings{
		ThreadCount:  10,
		RequestCount: 10,
	}
	result, err := kurl.Do(
		settings,
		*expectedRequest,
	)
	assert.Nil(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 0, result.ErrorCount)
	assert.Equal(t, settings.ThreadCount*settings.RequestCount, result.CompletedCount)
	assert.Equal(t, result.CompletedCount/2, result.StatusCodesFrequency[http.StatusOK])
	assert.Equal(t, result.CompletedCount/2, result.StatusCodesFrequency[http.StatusTooManyRequests])
}

func TestUnreachableServer(t *testing.T) {
	expectedRequest, err := http.NewRequest("POST", "localhost:9999", nil)
	require.Nil(t, err)

	result, err := kurl.Do(
		kurl.Settings{
			ThreadCount:  5,
			RequestCount: 10,
		},
		*expectedRequest,
	)
	assert.Nil(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 0, result.CompletedCount)
	assert.Equal(t, 50, result.ErrorCount)
	assert.Equal(t, 0, len(result.StatusCodesFrequency))

}

func TestMany(t *testing.T) {

	settings := kurl.Settings{
		ThreadCount:  10,
		RequestCount: 10,
	}

	lock := sync.Mutex{}
	seen := make([]int, settings.ThreadCount)

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		require.Equal(t, "GET", req.Method)
		i, err := strconv.Atoi(req.Header.Get("test"))
		require.Nil(t, err)

		lock.Lock()
		seen[i]++
		lock.Unlock()

		rw.Write([]byte(`OK`))
	}))
	defer server.Close()

	requests := make([]*http.Request, settings.ThreadCount)
	for i := 0; i < settings.ThreadCount; i++ {
		var err error
		requests[i], err = http.NewRequest("GET", server.URL, nil)
		require.Nil(t, err)
		requests[i].Header.Add("test", strconv.Itoa(i))
	}

	result, err := kurl.DoMany(
		settings,
		requests,
	)

	for i := 0; i < settings.ThreadCount; i++ {
		require.Equal(t, settings.RequestCount, seen[i])
	}

	assert.Nil(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 0, result.ErrorCount)
	assert.Equal(t, settings.ThreadCount*settings.RequestCount, result.CompletedCount)
	assert.Equal(t, result.CompletedCount, result.StatusCodesFrequency[http.StatusOK])
}

func TestDoManyRequestsLengthMismatch(t *testing.T) {
	settings := kurl.Settings{
		ThreadCount:  7,
		RequestCount: 3,
	}

	requests := make([]*http.Request, 3)

	result, err := kurl.DoMany(
		settings,
		requests,
	)

	require.NotNil(t, err)
	assert.Equal(t, "The length of requests must be equal to settings.ThreadCount", err.Error())
	assert.Nil(t, result)
}

func TestDoManyTestsLengthMismatch(t *testing.T) {
	settings := kurl.Settings{
		ThreadCount:  2,
		RequestCount: 3,
	}

	requests := make([]*http.Request, 2)
	var err error
	requests[0], err = http.NewRequest("GET", "localhost", nil)
	require.Nil(t, err)
	requests[1], err = http.NewRequest("GET", "localhost", nil)
	require.Nil(t, err)

	tests := make([]kurl.Test, 3)

	result, err := kurl.DoManyTest(
		settings,
		requests,
		tests,
	)

	require.NotNil(t, err)
	assert.Equal(t, "The length of tests must be equal to settings.ThreadCount", err.Error())
	assert.Nil(t, result)
}

func TestDoManyNilRequest(t *testing.T) {
	settings := kurl.Settings{
		ThreadCount:  1,
		RequestCount: 1,
	}

	requests := make([]*http.Request, 1)

	result, err := kurl.DoMany(
		settings,
		requests,
	)

	require.NotNil(t, err)
	assert.Equal(t, "The requests array cannot contain nil pointers", err.Error())
	assert.Nil(t, result)
}

func TestDoManyTest(t *testing.T) {
	settings := kurl.Settings{
		ThreadCount:  5,
		RequestCount: 2,
	}

	lock := sync.Mutex{}
	seen := make([]int, settings.ThreadCount)

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		require.Equal(t, "GET", req.Method)
		i, err := strconv.Atoi(req.Header.Get("test"))
		require.Nil(t, err)

		lock.Lock()
		seen[i]++
		lock.Unlock()

		rw.Write([]byte(`OK`))
	}))
	defer server.Close()

	requests := make([]*http.Request, settings.ThreadCount)
	tests := make([]kurl.Test, settings.ThreadCount)
	for i := 0; i < settings.ThreadCount; i++ {
		var err error
		requests[i], err = http.NewRequest("GET", server.URL, nil)
		require.Nil(t, err)
		requests[i].Header.Add("test", strconv.Itoa(i))

		tests[i] = kurl.Test(func(resp *http.Response, latency time.Duration) {
			require.Equal(t, 200, resp.StatusCode)
		})
	}

	result, err := kurl.DoManyTest(
		settings,
		requests,
		tests,
	)

	for i := 0; i < settings.ThreadCount; i++ {
		require.Equal(t, settings.RequestCount, seen[i])
	}

	assert.Nil(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 0, result.ErrorCount)
	assert.Equal(t, settings.ThreadCount*settings.RequestCount, result.CompletedCount)
	assert.Equal(t, result.CompletedCount, result.StatusCodesFrequency[http.StatusOK])
}

func TestWaitBetweenRequests(t *testing.T) {

	hasSeenFirstRequest := false
	lastRequestTime := time.Now()

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		now := time.Now()
		if hasSeenFirstRequest {
			require.LessOrEqual(t, int64(150), now.Sub(lastRequestTime).Milliseconds())
		}
		hasSeenFirstRequest = true
		rw.Write([]byte(`OK`))
	}))
	defer server.Close()

	expectedRequest, err := http.NewRequest("GET", server.URL, nil)
	require.Nil(t, err)

	settings := kurl.Settings{
		ThreadCount:         1,
		RequestCount:        5,
		WaitBetweenRequests: 150 * time.Millisecond,
	}

	result, err := kurl.Do(
		settings,
		*expectedRequest,
	)
	assert.Nil(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 0, result.ErrorCount)
	assert.Equal(t, settings.ThreadCount*settings.RequestCount, result.CompletedCount)
	assert.Equal(t, result.CompletedCount, result.StatusCodesFrequency[http.StatusOK])
}

func TestWarm(t *testing.T) {
	lock := sync.Mutex{}
	seenWarm := false

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {

		lock.Lock()
		if !seenWarm {
			assert.Equal(t, "GET", req.Method, "Server is not receiving test requests before warmup!")
			seenWarm = true
		} else {
			assert.Equal(t, "POST", req.Method, "Server did not receive a warmup!")
		}
		lock.Unlock()

		rw.Write([]byte(`OK`))
	}))
	defer server.Close()

	request, err := http.NewRequest("POST", server.URL, nil)
	require.Nil(t, err)

	settings := kurl.Settings{
		Warm:         true,
		ThreadCount:  5,
		RequestCount: 5,
	}

	result, err := kurl.Do(
		settings,
		*request,
	)
	assert.Nil(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 0, result.ErrorCount)
	assert.Equal(t, settings.ThreadCount*settings.RequestCount, result.CompletedCount)
	assert.Equal(t, result.CompletedCount, result.StatusCodesFrequency[http.StatusOK])
}

func TestWarmFailed(t *testing.T) {
	request, err := http.NewRequest("POST", "http://localhost:9999", nil)
	require.Nil(t, err)

	settings := kurl.Settings{
		Warm:         true,
		ThreadCount:  5,
		RequestCount: 5,
	}

	result, err := kurl.Do(
		settings,
		*request,
	)
	require.NotNil(t, err)
	assert.Equal(t, "Warm failed: ", err.Error()[0:13])
	assert.Nil(t, result)
}
