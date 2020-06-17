package main

import "strconv"

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
