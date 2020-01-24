#!/bin/sh

# -race requires CGO and some callers (i.e. Travis CI pipeline) pass that flag
env CGO_ENABLED=1 go test -mod vendor -ldflags="-w -s" -count=1 $@ ./...