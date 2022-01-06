package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

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

// AtoTimeNilIfEmpty returns a pointer to a time value represented by the given string,
// or nil if the given string is empty.
func AtoTimeNilIfEmpty(s string) (*time.Time, error) {
	if s == "" {
		return nil, nil
	}
	// Parse in default format output by JSON encoding
	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// AtoStringArray is for JSON string arrays.
func AtoStringArray(s string) ([]string, error) {
	var out []string
	if len(s) == 0 {
		return out, nil
	}
	err := json.Unmarshal([]byte(s), &out)
	if err != nil {
		return []string{}, fmt.Errorf("parsing string array: %w", err)
	}
	return out, nil
}

// AtoInt64Array is for parsing ID lists where the input is a JSON encoded array of ints.
func AtoInt64Array(s string) ([]int64, error) {
	var out []int64
	if len(s) == 0 {
		return out, nil
	}
	err := json.Unmarshal([]byte(s), &out)
	if err != nil {
		return []int64{}, fmt.Errorf("parsing int array: %w", err)
	}
	return out, nil
}

// IntToA converts an int to a base 10 string.
func IntToA(i int) string {
	return strconv.FormatInt(int64(i), 10)
}

// UintToA converts a uint to a base 10 string.
func UintToA(i uint) string {
	return strconv.FormatUint(uint64(i), 10)
}

// Uint64ToA converts a uint64 to a base 10 string.
func Uint64ToA(i uint64) string {
	return strconv.FormatUint(i, 10)
}

// Substr returns the first n characters from the given string.
func Substr(s string, n int) string {
	if n > len(s) {
		return s
	}
	return s[:n]
}
