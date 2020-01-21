#!/bin/sh

# setup the development environment
echo
echo Installing bash, git, libc6-compat, build-base
apk --no-cache add \
    bash \
    git \
    libc6-compat \
    build-base 

echo
echo Installing Delve
go get -u github.com/go-delve/delve/cmd/dlv
