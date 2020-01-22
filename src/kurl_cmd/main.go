package main

import (
	"time"
	"fmt"
	"net/http"
	"os"
	"sync"
)

func worker(
	begin *sync.WaitGroup,
	ready *sync.WaitGroup,
	complete *sync.WaitGroup, 
	codes *map[int]int,
	errorCount *int,
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

	errorCnt := 0
	ready.Done()

	begin.Wait()
	for i := 0; i < requestCount; i++ { 
		resp, err := client.Do(req)
		if err != nil {
			errorCnt++
		} else {
			(*codes)[resp.StatusCode]++
		}

		time.Sleep(time.Duration(waitBetweenRequestsMs) * time.Millisecond)
	}
	*errorCount = errorCnt
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

	var codes []map[int]int
	errorCount := make([]int,threadCount) // tracks http errors, i.e. we couldn't even get a status code

	var workersReady sync.WaitGroup
	var workersBegin sync.WaitGroup
	var workersComplete sync.WaitGroup
	
	// Launch one worker per thread, all blocked on workersBegin signal
	workersBegin.Add(1) 
	for i := 0; i < threadCount; i++ {
		c := make(map[int]int)
		codes = append(codes, c)
		workersReady.Add(1)
		workersComplete.Add(1)
		go worker(&workersBegin, &workersReady, &workersComplete, &c, &errorCount[i])
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
	sumCodes := make(map[int]int)
	sumErrors := 0
	for i := 0; i < threadCount; i++ {
		sumErrors += errorCount[i]

		for k,v := range codes[i] {
			sumCodes[k] += v
		}
	}

	// Format output
	fmt.Printf("total: %d\n", threadCount * requestCount)
	fmt.Printf("errors: %d\n", sumErrors)
	for statusCode,count := range sumCodes {
		fmt.Printf("status code %d: %d (%s)\n", statusCode, count, http.StatusText(statusCode))
	}
	fmt.Printf("duration: %v\n", elapsed)
	fmt.Printf("rate: %f requests/sec\n", float64(threadCount * requestCount) / elapsed.Seconds())
}
