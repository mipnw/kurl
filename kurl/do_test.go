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

	result := kurl.Do(
		kurl.Settings{
			ThreadCount:  10,
			RequestCount: 10,
		},
		*expectedRequest,
	)

	assert.Equal(t, 100, result.RequestsCount)
	assert.Equal(t, 0, result.ErrorCount)
	assert.Equal(t, 100, result.StatusCodesFrequency[http.StatusOK])
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

	result := kurl.Do(
		kurl.Settings{
			ThreadCount:  10,
			RequestCount: 10,
		},
		*expectedRequest,
	)

	assert.Equal(t, 100, result.RequestsCount)
	assert.Equal(t, 0, result.ErrorCount)
	assert.Equal(t, 50, result.StatusCodesFrequency[http.StatusOK])
	assert.Equal(t, 50, result.StatusCodesFrequency[http.StatusTooManyRequests])
}

func TestUnreachableServer(t *testing.T) {
	expectedRequest, err := http.NewRequest("POST", "localhost:9999", nil)
	require.Nil(t, err)

	result := kurl.Do(
		kurl.Settings{
			ThreadCount:  5,
			RequestCount: 10,
		},
		*expectedRequest,
	)

	assert.Equal(t, 50, result.RequestsCount)
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

	require.Nil(t, err)
	require.Equal(t, 100, result.RequestsCount)
	require.Equal(t, 0, result.ErrorCount)
	require.Equal(t, 100, result.StatusCodesFrequency[http.StatusOK])
}

func TestDoManyLengthMismatch(t *testing.T) {
	settings := kurl.Settings{
		ThreadCount:  7,
		RequestCount: 3,
	}

	requests := make([]*http.Request, 3)

	result, err := kurl.DoMany(
		settings,
		requests,
	)

	assert.NotNil(t, err)
	assert.Equal(t, "The length of requests must be equal to settings.ThreadCount", err.Error())
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

	assert.NotNil(t, err)
	assert.Equal(t, "The requests array cannot contain nil pointers", err.Error())
	assert.Nil(t, result)
}
