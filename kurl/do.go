// Package kurl provides an API for running concurrent HTTP requests on one endpoint,
// while collecting statistics about the errors, status codes and latencies observed.
package kurl

import (
	"errors"
	"net/http"
	"sync"
	"time"
)

// Settings parameterizes the behavior the kurl.Do function.
type Settings struct {
	Timeout               time.Duration
	Verbose               bool
	WaitBetweenRequestsMs int
	ThreadCount           int
	RequestCount          int
}

// Result is the type of the return value of the Do function.
// It contains all the statiscics observed about the endpoint during the run.
type Result struct {
	RequestsCount        int
	ErrorCount           int
	OverallDuration      time.Duration
	ResponseLatencyMin   time.Duration
	ResponseLatencyAvg   time.Duration
	ResponseLatencyMax   time.Duration
	StatusCodesFrequency map[int]int
}

// Do issues a set of concurrent identical HTTP requests.
func Do(
	settings Settings,
	request http.Request,
) Result {
	// Prepare thread synchronization
	var workersReady sync.WaitGroup
	var workersBegin sync.WaitGroup
	var workersComplete sync.WaitGroup
	workersBegin.Add(1)

	// Launch one worker per thread, all blocked on workersBegin signal
	workerResults := make([]workerResult, settings.ThreadCount)
	for i := 0; i < settings.ThreadCount; i++ {
		workerResults[i].latency = make([]time.Duration, settings.RequestCount)
		workerResults[i].statusCodesCount = make(map[int]int)
		workersReady.Add(1)
		workersComplete.Add(1)

		go worker(
			&settings,
			request, // a copy of the request on the stack, each worker can independently modify the request inside http.Do
			&workersBegin,
			&workersReady,
			&workersComplete,
			&workerResults[i],
		)
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
	return result
}

func aggregateResults(
	settings Settings,
	elapsed time.Duration,
	workerResults []workerResult,
) Result {
	result := Result{
		RequestsCount:        settings.ThreadCount * settings.RequestCount,
		OverallDuration:      elapsed,
		ResponseLatencyMin:   workerResults[0].latency[0],
		ResponseLatencyMax:   workerResults[0].latency[0],
		StatusCodesFrequency: make(map[int]int),
	}

	sumLatency := time.Duration(0)
	for i := 0; i < settings.ThreadCount; i++ {
		result.ErrorCount += workerResults[i].errorCount

		for _, latency := range workerResults[i].latency {
			sumLatency += latency
			if latency < result.ResponseLatencyMin {
				result.ResponseLatencyMin = latency
			} else if latency > result.ResponseLatencyMax {
				result.ResponseLatencyMax = latency
			}
		}

		for statusCode, freq := range workerResults[i].statusCodesCount {
			result.StatusCodesFrequency[statusCode] += freq
		}
	}
	result.ResponseLatencyAvg = time.Duration(int64(float64(sumLatency.Milliseconds())/float64(result.RequestsCount))) * time.Millisecond
	return result
}

// DoMany issues a set of concurrent HTTP requests.
func DoMany(
	settings Settings,
	requests []*http.Request, // length of this array must be equal to settings.ThreadCount
) (*Result, error) {

	if settings.ThreadCount != len(requests) {
		return nil, errors.New("The length of requests must be equal to settings.ThreadCount")
	}

	// Prepare thread synchronization
	var workersReady sync.WaitGroup
	var workersBegin sync.WaitGroup
	var workersComplete sync.WaitGroup
	workersBegin.Add(1)

	// Launch one worker per thread, all blocked on workersBegin signal
	workerResults := make([]workerResult, settings.ThreadCount)
	for i := 0; i < settings.ThreadCount; i++ {
		workerResults[i].latency = make([]time.Duration, settings.RequestCount)
		workerResults[i].statusCodesCount = make(map[int]int)

		workersReady.Add(1)
		workersComplete.Add(1)

		if requests[i] == nil {
			return nil, errors.New("The requests array cannot contain nil pointers")
		}

		go worker(
			&settings,
			*requests[i], // a copy of the request on the stack, each worker can independently modify the request inside http.Do
			&workersBegin,
			&workersReady,
			&workersComplete,
			&workerResults[i],
		)
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
	return &result, nil
}
