package kurl

import (
	"time"
	"sync"
	"net/http"
)

type workerResult struct {
	errorCount int
	statusCodesCount map[int]int
	latency []time.Duration
}

func worker(
	settings *Settings,
	request http.Request,
	begin *sync.WaitGroup,
	ready *sync.WaitGroup,
	complete *sync.WaitGroup, 
	result *workerResult,
) {
	defer complete.Done()

	client := &http.Client{}
	
	ready.Done()

	begin.Wait()
	for i := 0; i < settings.RequestCount; i++ { 
		start := time.Now()
		resp, err := client.Do(&request)
		result.latency[i] = time.Since(start)
		if err != nil {
			result.errorCount++
		} else {
			result.statusCodesCount[resp.StatusCode]++
		}

		time.Sleep(time.Duration(settings.WaitBetweenRequestsMs) * time.Millisecond)
	}
}
