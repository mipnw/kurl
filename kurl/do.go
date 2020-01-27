// Package kurl provides an API for running concurrent HTTP requests on one endpoint,
// while collecting statistics about the errors, status codes and latencies observed.
package kurl

import (
	"errors"
	"net/http"
	"sync"
	"time"
)

// Test is a function to run on the response of every http request
type Test func(response *http.Response, latency time.Duration)

// Settings parameterizes the behavior the kurl.Do function.
type Settings struct {
	Timeout             time.Duration // http client timeout
	Verbose             bool          // increase kurl's verbosity
	WaitBetweenRequests time.Duration // delay between requests on each thread
	ThreadCount         int           // number of threads
	RequestCount        int           // number of identical and consecutive requests per thread
	Warm                bool          // warm up with 1 http request request
}

// Result is the type of the return value of the Do function.
// It contains all the statiscics observed about the endpoint during the run.
type Result struct {
	CompletedCount       int
	ErrorCount           int
	OverallDuration      time.Duration
	Latencies            []time.Duration
	StatusCodesFrequency map[int]int
}

// Do issues a set of concurrent and identical HTTP requests.
func Do(
	settings Settings,
	request http.Request,
) (*Result, error) {
	requests := make([]*http.Request, settings.ThreadCount)
	for i := 0; i < settings.ThreadCount; i++ {
		requests[i] = &request
	}
	return DoMany(settings, requests)
}

func aggregateResults(
	settings Settings,
	elapsed time.Duration,
	workerResults []workerResult,
) Result {
	result := Result{
		OverallDuration:      elapsed,
		StatusCodesFrequency: make(map[int]int),
	}

	for i := 0; i < settings.ThreadCount; i++ {
		result.ErrorCount += workerResults[i].errorCount
		result.CompletedCount += len(workerResults[i].latency) - workerResults[i].errorCount
		for statusCode, freq := range workerResults[i].statusCodesCount {
			result.StatusCodesFrequency[statusCode] += freq
		}
	}

	return result
}

// DoMany issues a set of concurrent HTTP requests, where each thread issues a sequence of requests that
// can be different from other threads. Use Do if all requests are the same.
func DoMany(
	settings Settings,
	requests []*http.Request, // length of this array must be equal to settings.ThreadCount
) (*Result, error) {
	tests := make([]Test, settings.ThreadCount)
	return DoManyTest(settings, requests, tests)
}

// DoManyTest issues a set of concurrent HTTP requests, where each thread issues a sequence of requests that
// can be different from other threads, and tests each HTTP response.
func DoManyTest(
	settings Settings,
	requests []*http.Request, // length of this array must be equal to settings.ThreadCount
	tests []Test, // length of this array must be equal to settings.ThreadCount
) (*Result, error) {
	if settings.ThreadCount != len(requests) {
		return nil, errors.New("The length of requests must be equal to settings.ThreadCount")
	}
	if settings.ThreadCount != len(tests) {
		return nil, errors.New("The length of tests must be equal to settings.ThreadCount")
	}

	// Prepare thread synchronization
	var workersReady sync.WaitGroup
	var workersBegin sync.WaitGroup
	var workersComplete sync.WaitGroup
	workersBegin.Add(1)

	// Launch one worker per thread, all blocked on workersBegin signal
	workerResults := make([]workerResult, settings.ThreadCount)
	latencies := make([]time.Duration, settings.RequestCount*settings.ThreadCount)
	for i := 0; i < settings.ThreadCount; i++ {
		workerResults[i].latency = latencies[i*settings.RequestCount : ((i + 1) * settings.RequestCount)]
		workerResults[i].statusCodesCount = make(map[int]int)

		workersReady.Add(1)
		workersComplete.Add(1)

		if requests[i] == nil {
			return nil, errors.New("The requests array cannot contain nil pointers")
		}

		go worker(
			&settings,
			*requests[i], // a copy of the request on the stack, each worker can independently modify the request inside http.Do
			tests[i],
			&workersBegin,
			&workersReady,
			&workersComplete,
			&workerResults[i],
		)
	}

	// Warm
	if settings.Warm {
		_, err := http.Get(requests[0].URL.String())
		if err != nil {
			return nil, errors.New("Warm failed: " + err.Error())
		}
	}

	// Wait until all workers are ready
	workersReady.Wait()

	// Release all the workers
	start := time.Now()
	workersBegin.Done()

	// Wait until all workers are done
	workersComplete.Wait()
	elapsed := time.Since(start)

	// Aggregate statistics
	result := aggregateResults(settings, elapsed, workerResults)
	result.Latencies = latencies
	return &result, nil
}
