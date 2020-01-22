package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
)

var (
	help bool
	post bool
	url string
	waitBetweenRequestsMs int
	threadCount int
	requestCount int
	headerValue headersValue
)

func usage() {
	var CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	fmt.Fprintf(CommandLine.Output(), "Kurl: load test HTTP traffic on a specified endpoint\n")
	flag.PrintDefaults()
}

func parseCommandLine() {
	flag.BoolVar(&post, "post", false, "use HTTP POST (default is GET)")
	flag.StringVar(&url, "url", "", "target endpoint")
	flag.IntVar(&threadCount, "thread", 10, "number of parallel threads")
	flag.IntVar(&requestCount, "request", 10, "number of http requests per thread")
	flag.IntVar(&waitBetweenRequestsMs, "wait", 0, "number of milliseconds to wait between requests")
	flag.BoolVar(&help, "help", false, "print this helper")

	headerValue.header = make(http.Header)
	flag.Var(&headerValue, "h", "an HTTP header in the form key=value")

	flag.Parse()
}