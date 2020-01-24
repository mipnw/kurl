#!/bin/bash

fmt=`go fmt ./...`
[[ -n $fmt ]] && echo "go fmt error:\n$fmt" && exit 1

vet=`go vet ./...`
[[ -n $vet ]] && echo "go vet error:\n$vet" && exit 1
