package main

import (
	"database/sql"
	"fmt"
	"net/url"
	"strconv"
)

// DBConn wraps a *sql.DB or *sql.Tx.
type DBConn interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}

func validateURL(s string) error {
	u, err := url.ParseRequestURI(s)
	if err != nil {
		return fmt.Errorf("parsing URL: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("invalid scheme: %s", u.Scheme)
	}
	return nil
}

// AtoInt converts base 10 string to int.
func AtoInt(s string) (int, error) {
	r, err := strconv.ParseInt(s, 10, 64)
	return int(r), err
}

// AtoUint converts base 10 string to uint.
func AtoUint(s string) (uint, error) {
	r, err := strconv.ParseUint(s, 10, 64)
	return uint(r), err
}

// AtoInt64 converts base 10 string to int64.
func AtoInt64(s string) (int64, error) {
	r, err := strconv.ParseInt(s, 10, 64)
	return r, err
}

// AtoInt64NilIfEmpty ("ne: nil if empty") converts base 10 string to int64,
// and returns nil on err or empty.
func AtoInt64NilIfEmpty(s string) (*int64, error) {
	r, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		e := err.(*strconv.NumError)
		if e.Num == "" {
			// Input was blank; return no error
			return nil, nil
		}
		return nil, err
	}
	return &r, nil
}

// AtoPointerNilIfEmpty returns a pointer to the given string, or nil if given an empty string.
func AtoPointerNilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// AtoBool converts the given string to boolean.
func AtoBool(a string) bool {
	return a == "1" || a == "true"
}

// https://stackoverflow.com/a/15323988/1597274
func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
