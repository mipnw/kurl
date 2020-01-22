# kurl

Command line tool to load test an HTTP endpoint, written in Go.

Supports the HTTP GET and POST methods, with headers, but currently no body.

Configurable thread count, request per thread, and delays between requests. Outputs the aggregate HTTP status codes counts. E.g.
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
docker run --rm -it mipnw/kurl:latest
# > kurl -help
```

If you have Make and Golang installed, you can build kurl and have it binplaced at /usr/local/bin:
```bash
scripts/build.sh --release
kurl -help
```
