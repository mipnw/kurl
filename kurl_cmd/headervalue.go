package kurlcmd

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

type headersValue struct {
	header http.Header
}

func (hv *headersValue) String() string {
	return fmt.Sprintf("%v", hv.header)
}

func (hv *headersValue) Set(value string) error {
	arr := strings.Split(value, "=")
	if len(arr) != 2 {
		return errors.New("Bad header argument")
	}
	hv.header.Add(arr[0], arr[1])
	return nil
}
