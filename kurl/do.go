package kurl

import (
	"time"
	"sync"
	"net/http"
)

type Result struct {
	RequestsCount int
	ErrorCount int
	OverallDuration time.Duration
	ResponseLatencyMin time.Duration
	ResponseLatencyAvg time.Duration
	ResponseLatencyMax time.Duration
	StatusCodesFrequency map[int]int
}

func Do(
	settings Settings,
	request http.Request,
) Result {
	// Launch one worker per thread, all blocked on workersBegin signal
	workerResults := make([]workerResult, settings.ThreadCount)	
	var workersReady sync.WaitGroup
	var workersBegin sync.WaitGroup
	var workersComplete sync.WaitGroup
	workersBegin.Add(1) 
	for i := 0; i < settings.ThreadCount; i++ {
		workerResults[i].latency = make([]time.Duration, settings.RequestCount)
		workerResults[i].statusCodesCount = make(map[int]int)

		workersReady.Add(1)
		workersComplete.Add(1)
		go worker(
			&settings,
			request, // a copy of the request on the stack, each worker can independantly modify the request inside http.Do
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
	result := Result {
		RequestsCount: settings.ThreadCount * settings.RequestCount,
		OverallDuration: elapsed,
		ResponseLatencyMin: workerResults[0].latency[0],
		ResponseLatencyMax: workerResults[0].latency[0],
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
	result.ResponseLatencyAvg = time.Duration(int64(float64(sumLatency.Milliseconds()) / float64(result.RequestsCount))) * time.Millisecond
	return result
}