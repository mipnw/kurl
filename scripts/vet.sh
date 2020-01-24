#!/bin/bash

fmt=`go fmt ./...`
[[ -n $fmt ]] && printf "go fmt error:\n$fmt\n" && exit 1

vet=`env CGO_ENABLED=0 go vet -mod=vendor ./...`
[[ -n $vet ]] && printf "go vet error:\n$vet\n" && exit 1

exit 0
