# kurl

Command line tool to load test an HTTP endpoint, written in Go.

Configurable thread count, request per thread, and delays between requests. Outputs the aggregate HTTP status codes counts. E.g.
```
total: 2500
errors: 0
status code 200: 1144
status code 429: 1352
status code 504: 1
status code 503: 3
duration: 5.2719082s
rate: 474.211596 requests/sec
```

# Usage
If you have docker installed, you do not need to build, you can run kurl inside a container:
```bash
docker pull mipnw/kurl:latest
docker run --rm -it mipnw/kurl:latest
# > kurl -help
```

If you have Make and Golang installed, you can build kurl for your workstation's architecture:
```bash
make shell
scripts/build.sh --release [--mac|--windows|--linux]
bin/kurl -help
```
