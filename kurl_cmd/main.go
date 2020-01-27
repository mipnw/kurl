package main

import (
	"fmt"
	"github.com/mipnw/kurl/kurl"
	"math"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

func validateCommandLine() bool {
	_, err := url.ParseRequestURI(endpoint)
	if err != nil {
		fmt.Printf("-url argument is required and must be a valid URL\n\n")
		return false
	}
	if bodyFilename != "" {
		info, err := os.Stat(bodyFilename)
		if os.IsNotExist(err) || info.IsDir() {
			fmt.Printf("file %s does not exist\n\n", bodyFilename)
			return false
		}
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
	req, err := http.NewRequest(method, endpoint, nil)
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

	result, err := kurl.Do(settings, *request)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	// Error count to stderr
	if result.ErrorCount != 0 {
		fmt.Fprintf(os.Stderr, "http errors: %d\n", result.ErrorCount)
	}

	// Formatted output to stdout
	if printLatencies {
		// space-separated, millisecond rounted latencies to stdout, for easy loading in your favorite math IDE
		outputStr := ""
		for i := 0; i < len(result.Latencies); i++ {
			outputStr += fmt.Sprintf("%d ", result.Latencies[i].Round(time.Millisecond).Milliseconds())
		}
		outputStr = strings.TrimRight(outputStr, " ")
		fmt.Println(outputStr)
	} else {
		// Default formatted output
		fmt.Printf("completed: %d %.0fHz\n", result.CompletedCount, float64(result.CompletedCount)/result.OverallDuration.Seconds())

		if result.CompletedCount > 0 {
			for statusCode, freq := range result.StatusCodesFrequency {
				fmt.Printf("http %d (%s): %d %d%% %.0fHz\n",
					statusCode, http.StatusText(statusCode),
					freq,
					int(100*float32(freq)/float32(result.CompletedCount)),                      // percentage
					float64(result.StatusCodesFrequency[200])/result.OverallDuration.Seconds()) // rate in Hz
			}
		}

		fmt.Printf("duration: %v\n", result.OverallDuration.Round(time.Millisecond))
		printLatencyStats(result)
	}
}

func printLatencyStats(result *kurl.Result) {
	minLatency := result.Latencies[0]
	var avgLatency time.Duration
	maxLatency := result.Latencies[0]

	completed := 0
	for i := 0; i < len(result.Latencies); i++ {
		// Skip the flagged latencies which correspond to HTTP errors
		if result.Latencies[i] == 0 {
			continue
		}

		completed++
		avgLatency += result.Latencies[i]
		if result.Latencies[i] < minLatency {
			minLatency = result.Latencies[i]
		} else if result.Latencies[i] > maxLatency {
			maxLatency = result.Latencies[i]
		}
	}
	avgLatency = time.Duration(float64(avgLatency.Milliseconds())/float64(completed)) * time.Millisecond
	if completed != result.CompletedCount {
		panic("Kurl has a bug in the handling of latency statistics when there are HTTP errors")
	}

	var stdLatency time.Duration
	if result.CompletedCount > 1 {
		var agg int64
		for i := 0; i < result.CompletedCount; i++ {
			d := result.Latencies[i] - avgLatency
			agg += (d.Microseconds() * d.Microseconds())
		}
		avg := float64(agg) / float64(result.CompletedCount-1)
		stdLatency = time.Duration(math.Sqrt(avg)) * time.Microsecond
	}

	if result.CompletedCount > 0 {
		fmt.Printf("latency  min: %v, avg: %v, max: %v (std:%v)\n",
			minLatency.Round(time.Millisecond),
			avgLatency.Round(time.Millisecond),
			maxLatency.Round(time.Millisecond),
			stdLatency.Round(time.Millisecond))
	}
}
