# kurl

Command line tool to load test an HTTP endpoint, written in Go.

Configurable thread count, request per thread, and delays between requests. Outputs the aggregate HTTP status codes counts. E.g.
```
Statistics:
http status code 503: 18
http status code 429:  7
http status code 200: 75
Duration: 3.20s
Rate: 31.15 requests/sec
```

# Build
`go build -o bin/kurl src/*"`

# Usage
See `bin/kurl -help`
