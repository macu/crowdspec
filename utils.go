package main

import (
	"fmt"
	"net"
	"net/http"
	"reflect"
)

// https://stackoverflow.com/a/15323988/1597274
func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func int64InSlice(i int64, list []int64) bool {
	for _, v := range list {
		if v == i {
			return true
		}
	}
	return false
}

// https://mangatmodi.medium.com/go-check-nil-interface-the-right-way-d142776edef1
func isNil(a interface{}) bool {
	if a == nil {
		return true
	}
	switch reflect.TypeOf(a).Kind() {
	case reflect.Ptr, reflect.Map, reflect.Array, reflect.Chan, reflect.Slice:
		return reflect.ValueOf(a).IsNil()
	}
	return false
}

// This function returns the port number for the given request.
func getLocalPort(r *http.Request) (string, error) {
	a, ok := r.Context().Value(http.LocalAddrContextKey).(net.Addr)
	if !ok {
		return "", fmt.Errorf("getting local address from request context")
	}
	_, port, err := net.SplitHostPort(a.String())
	if err != nil {
		return "", err
	}
	return port, nil
}
