#!/bin/sh

fmt=`go fmt ./...`
echo $fmt
[[ -n $fmt ]] && exit 1

vet=`go vet ./...`
echo $vet
[[ -n $vet ]] && exit 1
