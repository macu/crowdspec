package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
)

// getUserIP returns the IP address of the remote connection.
func getUserIP(r *http.Request) string {
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}

// getLocalPort returns the server (local) port number for the given request.
func getLocalPort(r *http.Request) (string, error) {
	a, ok := r.Context().Value(http.LocalAddrContextKey).(net.Addr)
	if !ok {
		return "", fmt.Errorf("getting local address from request context")
	}
	_, port, err := net.SplitHostPort(a.String())
	if err != nil {
		return "", fmt.Errorf("extracting port: %w", err)
	}
	return port, nil
}

// buildAbsoluteURL returns an absolute URL to the given path on the currently running server.
func buildAbsoluteURL(r *http.Request, path string) (string, error) {
	if strings.HasPrefix(path, "/") {
		path = path[1:]
	}

	if isAppEngine() {
		return fmt.Sprintf("https://%s/%s", os.Getenv("DOMAIN"), path), nil
	}

	port, err := getLocalPort(r)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("http://localhost:%s/%s", port, path), nil
}
