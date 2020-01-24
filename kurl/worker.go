package kurl

import (
	"net/http"
	"sync"
	"time"
)

type workerResult struct {
	errorCount       int
	statusCodesCount map[int]int
	latency          []time.Duration
}

func worker(
	settings *Settings,
	request http.Request,
	test Test,
	begin *sync.WaitGroup,
	ready *sync.WaitGroup,
	complete *sync.WaitGroup,
	result *workerResult,
) {
	defer complete.Done()

	client := &http.Client{
		Timeout: settings.Timeout,
	}

	ready.Done()

	begin.Wait()
	for i := 0; i < settings.RequestCount; i++ {

		start := time.Now()
		resp, err := client.Do(&request)
		result.latency[i] = time.Since(start)

		start = time.Now()
		if err != nil {
			result.errorCount++
		} else {
			result.statusCodesCount[resp.StatusCode]++
		}

		// Run the test if we have one
		if test != nil {
			test(resp, result.latency[i])
		}

		// Delay this thread if we need to wait between requests
		elapsedSinceLastRequest := time.Since(start)
		if elapsedSinceLastRequest < settings.WaitBetweenRequests {
			time.Sleep(settings.WaitBetweenRequests - elapsedSinceLastRequest)
		}
	}
}
