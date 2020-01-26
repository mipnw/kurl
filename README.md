# kurl
[![LatestRelease](https://img.shields.io/github/v/release/mipnw/kurl?sort=semver)](https://github.com/mipnw/kurl/releases/latest)
![Last Commit](https://img.shields.io/github/last-commit/mipnw/kurl)

[![Travis Build Status](https://img.shields.io/travis/com/mipnw/kurl)](https://travis-ci.com/mipnw/kurl)
[![Docker Build Status](https://img.shields.io/docker/cloud/build/mipnw/kurl)](https://hub.docker.com/r/mipnw/kurl)
[![Docker Pulls](https://img.shields.io/docker/pulls/mipnw/kurl)](https://hub.docker.com/r/mipnw/kurl)
[![Code coverage](https://img.shields.io/codecov/c/github/mipnw/kurl)](https://codecov.io/gh/mipnw/kurl)

[![Go doc](https://godoc.org/github.com/mipnw/kurl/kurl?status.svg)](http://godoc.org/github.com/mipnw/kurl/kurl)
[![Go Report Card](https://goreportcard.com/badge/github.com/mipnw/kurl)](https://goreportcard.com/report/github.com/mipnw/kurl)
[![Go Version](https://img.shields.io/github/go-mod/go-version/mipnw/kurl)](https://golang.org/)


CLI, and reusable Go package, for load testing an HTTP endpoint.

Supports HTTP GET and POST, with headers and body.

Configurable thread count, request per thread, and delays between requests. Outputs the aggregate HTTP status codes frequencies, and latencies. 

Examples
```
# > kurl -url [https://domain/path] -thread 200 -request 10
total: 2000
status code 200: 470 23% (OK)
status code 429: 1530 76% (Too Many Requests)
duration: 3.265s
latency  min: 31ms, avg: 298ms, max: 959ms (std: 153ms)
rate: 613 Hz
```

Use command line argument `-pl` to print all latencies to stdout:
```
# > kurl -url [https://domain/path] -pl -thread 1 -request 50
98 98 97 98 102 109 96 90 102 93 101 99 87 98 91 105 94 107 89 108 95 96 111 94 104 92 98 95 103 86 108 104 94 102 95 95 96 106 97 98 88 91 100 99 93 102 99 98 109 91
```

# Usage
Provided you have docker installed, you can run the Kurl CLI without having to build it.
```bash
docker pull mipnw/kurl:latest
docker run --rm mipnw/kurl:latest -help
```

You may also use Kurl inside your Go application:
```go
import "github.com/mipnw/kurl/kurl"

// Launch 100 concurrent HTTP requests
request, _ := http.NewRequest("GET", "https://domain/path", nil)
result := kurl.Do(
  kurl.Settings{ThreadCount:100, RequestCount: 1},
  *request)
```
See [Go Doc](https://godoc.org/github.com/mipnw/kurl/kurl) for API reference.

#  Build
If you have Golang installed, you can build kurl:
```bash
scripts/build.sh --release
kurl -help
```

If you do not have Golang installed, and your OS is [Mac, Linux, Windows] and your architecture x86_64, you can still build Kurl if you have Make and Docker.
```bash
make build
bin/kurl -help
```
