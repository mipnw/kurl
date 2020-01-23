# kurl
[![Travis Build Status](https://img.shields.io/travis/com/mipnw/kurl)](https://travis-ci.com/mipnw/kurl)
[![Docker Build Status](https://img.shields.io/docker/cloud/build/mipnw/kurl)](https://hub.docker.com/r/mipnw/kurl)
[![Docker Pulls](https://img.shields.io/docker/pulls/mipnw/kurl)](https://hub.docker.com/r/mipnw/kurl)
[![Code coverage](https://img.shields.io/codecov/c/github/mipnw/kurl)](https://codecov.io/gh/mipnw/kurl)

[![Go doc](https://godoc.org/github.com/mipnw/kurl/kurl?status.svg)](http://godoc.org/github.com/mipnw/kurl/kurl)
[![Go Report Card](https://goreportcard.com/badge/github.com/mipnw/kurl)](https://goreportcard.com/report/github.com/mipnw/kurl)
[![Go Version](https://img.shields.io/github/go-mod/go-version/mipnw/kurl)](https://golang.org/)


CLI for load testing an HTTP endpoint, along with a reusable Go package exposing the same functionality.

Supports the HTTP GET and POST methods, with headers as command line arguments, and a body from file.

Configurable thread count, request per thread, and delays between requests. Outputs the aggregate HTTP status codes counts and latency statistics. E.g.
```
total: 2000
errors: 0
status code 200: 470 23% (OK)
status code 429: 1530 76% (Too Many Requests)
duration: 3.265s
latency  min: 31ms, avg: 298ms, max: 959ms
rate: 613 Hz
```

# Usage
If you have docker installed, you do not need to build, you can run kurl inside a container:
```bash
docker pull mipnw/kurl:latest
docker run --rm mipnw/kurl:latest -help
```
or in interactive form:
```bash
docker pull mipnw/kurl:latest
docker run --rm -it --entrypoint /bin/sh mipnw/kurl:latest
# > kurl -help
```

# Build
If you have Make and Golang installed, you can build kurl:
```bash
scripts/build.sh --release
kurl -help
```

If you have Make and Docker but not Golang, you can build for your platform inside docker:
```bash
make shell
# > scripts/build.sh --release [--mac|--linux|--windows] --binplace
# > exit
bin/kurl -help
```
