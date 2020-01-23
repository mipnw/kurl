package main

import (
	"fmt"
	"github.com/mipnw/kurl/kurl"
	"net/http"
	"os"
	"time"
)

func validateCommandLine() bool {
	if url == "" {
		fmt.Printf("-url argument is required\n\n")
		return false
	}
	return true
}

func makeHTTPRequest() (*http.Request, error) {
	var method string

	if post {
		method = "POST"
	} else {
		method = "GET"
	}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header = headerValue.header

	if bodyFilename != "" {
		file, err := os.Open(bodyFilename)
		if err != nil {
			return nil, err
		}
		req.Body = file
	}
	return req, nil
}

func main() {
	parseCommandLine()
	if help || !validateCommandLine() {
		usage()
		return
	}

	request, err := makeHTTPRequest()
	if err != nil {
		fmt.Println(err)
	}

	result := kurl.Do(settings, *request)

	// Format output
	fmt.Printf("total: %d\n", result.RequestsCount)
	fmt.Printf("errors: %d\n", result.ErrorCount)
	for statusCode, freq := range result.StatusCodesFrequency {
		fmt.Printf("status code %d: %d %d%% (%s)\n",
			statusCode,
			freq,
			int(100*float32(freq)/float32(result.RequestsCount)),
			http.StatusText(statusCode))
	}
	fmt.Printf("duration: %v\n", result.OverallDuration.Round(time.Millisecond))
	fmt.Printf("latency  min: %v, avg: %v, max: %v\n",
		result.ResponseLatencyMin.Round(time.Millisecond),
		result.ResponseLatencyAvg.Round(time.Millisecond),
		result.ResponseLatencyMax.Round(time.Millisecond))
	fmt.Printf("rate: %.0f Hz\n", float64(result.RequestsCount)/result.OverallDuration.Seconds())
	fmt.Printf("200 rate: %.0f Hz\n ", float64(result.StatusCodesFrequency[200])/result.OverallDuration.Seconds())
}
