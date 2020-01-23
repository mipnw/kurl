package kurl

import (
	"net/http"
)

type Settings struct {
	Verbose bool
	Request http.Request
	WaitBetweenRequestsMs int
	ThreadCount int
	RequestCount int
}