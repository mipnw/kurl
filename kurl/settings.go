package kurl

import (
	"net/http"
)

// Settings parameterizes the behavior the kurl.Do function.
type Settings struct {
	Verbose               bool
	Request               http.Request
	WaitBetweenRequestsMs int
	ThreadCount           int
	RequestCount          int
}
