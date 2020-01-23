package kurl_test

import (
	"github.com/mipnw/kurl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func TestStatus200(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "GET", req.Method)
		rw.Write([]byte(`OK`))
	}))
	defer server.Close()

	expectedRequest, err := http.NewRequest("GET", server.URL, nil)
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
