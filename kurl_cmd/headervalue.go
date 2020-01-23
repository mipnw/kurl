package main

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
) 

type headersValue struct {
	header http.Header
}

func (this *headersValue) String() string {
    return fmt.Sprintf("%v", this.header)
}

func (this *headersValue) Set(value string) error {
	arr := strings.Split(value, "=")
	if len(arr) != 2 {
		return errors.New("Bad header argument")
	}
	this.header.Add(arr[0], arr[1])
    return nil
}
