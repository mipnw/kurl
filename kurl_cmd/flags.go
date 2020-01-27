package main

import (
	"flag"
	"fmt"
	"github.com/mipnw/kurl/kurl"
	"net/http"
	"os"
	"time"
)

var (
	settings       kurl.Settings
	help           bool
	post           bool
	endpoint       string
	headerValue    headersValue
	bodyFilename   string
	printLatencies bool
)

func usage() {
	var CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	fmt.Fprintf(CommandLine.Output(), "Kurl: load test HTTP traffic on a specified endpoint\n")
	flag.PrintDefaults()
}

func parseCommandLine() {
	flag.BoolVar(&post, "post", false, "use HTTP POST (default is GET)")
	flag.StringVar(&endpoint, "url", "", "target endpoint")
	flag.IntVar(&settings.ThreadCount, "thread", 10, "number of parallel threads")
	flag.IntVar(&settings.RequestCount, "request", 10, "number of http requests per thread")
	flag.DurationVar(&settings.WaitBetweenRequests, "wait", 0, "how long to wait between requests on each thread")
	flag.BoolVar(&help, "help", false, "print this helper")
	flag.StringVar(&bodyFilename, "body", "", "path to file containing HTTP request body")
	flag.BoolVar(&printLatencies, "pl", false, "print space-separated millisecond-rounded latencies to stdout")
	flag.BoolVar(&settings.Warm, "warm", false, "Warm up with one HTTP request (not included in the result)")

	var defaultTimeout time.Duration
	flag.DurationVar(&settings.Timeout, "timeout", defaultTimeout, "http client timeout")

	headerValue.header = make(http.Header)
	flag.Var(&headerValue, "h", "an HTTP header in the form key=value")

	flag.Parse()
}
