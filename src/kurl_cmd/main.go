package main

import (
	"time"
	"fmt"
	"net/http"
	"os"
	"sync"
)

type workerResult struct {
	errorCount int
	statusCodesCount map[int]int
	latency []time.Duration
}

func worker(
	begin *sync.WaitGroup,
	ready *sync.WaitGroup,
	complete *sync.WaitGroup, 
	result *workerResult,
) {
	defer complete.Done()

	client := &http.Client{}
	var method string
	if post {
		method = "POST"
	} else {
		method = "GET"
	}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		fmt.Printf("Unable to create http request: %v\n", err)
		return
	}

	req.Header = headerValue.header

	if bodyFilename != "" {
		file, err := os.Open(bodyFilename)
		if err != nil {
			fmt.Printf("Unable to open -body file: %v\n", err)
			return
		}
		req.Body = file
	}

	ready.Done()

	begin.Wait()
	for i := 0; i < requestCount; i++ { 
		start := time.Now()
		resp, err := client.Do(req)
		result.latency[i] = time.Since(start)
		if err != nil {
			result.errorCount++
		} else {
			result.statusCodesCount[resp.StatusCode]++
		}

		time.Sleep(time.Duration(waitBetweenRequestsMs) * time.Millisecond)
	}
}

func validateCommandLine() bool {
	if url == "" {
		fmt.Printf("-url argument is required\n\n");
		return false
	}
	return true
}

func main() {
	parseCommandLine()
	if help || !validateCommandLine() {
		usage()
		return
	}

	//var codes []map[int]int

	workerResults := make([]workerResult, threadCount)
	
	var workersReady sync.WaitGroup
	var workersBegin sync.WaitGroup
	var workersComplete sync.WaitGroup
	
	// Launch one worker per thread, all blocked on workersBegin signal
	workersBegin.Add(1) 
	for i := 0; i < threadCount; i++ {
		workerResults[i].latency = make([]time.Duration, requestCount)
		workerResults[i].statusCodesCount = make(map[int]int)

		workersReady.Add(1)
		workersComplete.Add(1)
		go worker(&workersBegin, &workersReady, &workersComplete, &workerResults[i])
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
	minLatency := workerResults[0].latency[0]
	maxLatency := workerResults[0].latency[0]
	sumLatency := time.Duration(0)
	sumCodes := make(map[int]int)
	sumErrors := 0
	for i := 0; i < threadCount; i++ {
		sumErrors += workerResults[i].errorCount

		for _, latency := range workerResults[i].latency {
			sumLatency += latency
			if latency < minLatency {
				minLatency = latency
			} else if latency > maxLatency {
				maxLatency = latency
			}
		}

		for k,v := range workerResults[i].statusCodesCount {
			sumCodes[k] += v
		}
	}
	totalRequests := threadCount * requestCount
	avgLatency := float64(sumLatency.Milliseconds()) / float64(totalRequests)

	// Format output
	fmt.Printf("total: %d\n", totalRequests)
	fmt.Printf("errors: %d\n", sumErrors)
	for statusCode,count := range sumCodes {
		fmt.Printf("status code %d: %d %d%% (%s)\n", 
			statusCode, 
			count, 
			int(100*float32(count) / float32(totalRequests)),
			http.StatusText(statusCode))
	}
	fmt.Printf("duration: %v\n", elapsed.Round(time.Millisecond))
	fmt.Printf("latency  min: %v, avg: %.0fms, max: %v\n",
		minLatency.Round(time.Millisecond), 
		avgLatency, 
		maxLatency.Round(time.Millisecond))
	fmt.Printf("rate: %.0f Hz\n", float64(totalRequests) / elapsed.Seconds())
}
