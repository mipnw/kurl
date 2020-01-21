package main

import (
	"time"
	"fmt"
	"net/http"
	"sync"
)

func worker(wg *sync.WaitGroup, codes *map[int]int) {
	defer wg.Done()

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("Unable to create http request: %v\n", err)
		return
	}

	req.Header = headerValue.header
	
	for i := 0; i < requestCount; i++ { 
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Http error: %v\n", err)
			break
		}
		(*codes)[resp.StatusCode]++

		time.Sleep(time.Duration(waitBetweenRequestsMs) * time.Millisecond)
	}
}

func main() {
	parseCommandLine()
	if help {
		usage()
		return
	}

	fmt.Printf("Using %d threads and %d req/thread\n\n", threadCount, requestCount)

	var codes []map[int]int
	var wg sync.WaitGroup
	
	// Launch one worker per thread
	start := time.Now()
	for i := 0; i < threadCount; i++ {
		c := make(map[int]int)
		codes = append(codes, c)
		wg.Add(1)
		go worker(&wg, &c)
	}
	wg.Wait()
	elapsed := time.Since(start)

	// Aggregate statistics
	sumCodes := make(map[int]int)
	for i := 0; i < threadCount; i++ {
		for k,v := range codes[i] {
			sumCodes[k] += v
		}
	}
	fmt.Printf("\nStatistics:\n")
	for k,v := range sumCodes {
		fmt.Printf("http status code %d: %d\n", k, v)
	}
	fmt.Printf("Duration: %v\n", elapsed)
	fmt.Printf("Rate: %f requests/sec\n", float64(threadCount * requestCount) / elapsed.Seconds())
}
