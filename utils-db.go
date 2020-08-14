package main

import (
	"database/sql"
	"strconv"
)

// DBConn wraps a *sql.DB or *sql.Tx.
type DBConn interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}

// Returns a placeholder representing the given arg in args,
// adding the arg to args if not already present.
func argPlaceholder(arg interface{}, args *[]interface{}) string {
	for i := 0; i < len(*args); i++ {
		if (*args)[i] == arg {
			return "$" + strconv.Itoa(i+1)
		}
	}
	*args = append(*args, arg)
	return "$" + strconv.Itoa(len(*args))
}

// Returns a condition matching the given subspecID,
// adding the given subspecID to args if not already present.
func subspecCond(subspecID *int64, args *[]interface{}) string {
	if subspecID == nil {
		return "subspec_id IS NULL"
	}
	return "subspec_id = " + argPlaceholder(*subspecID, args)
}

// Returns a condition matching the given subspecID,
// using a fixed placeholder index when pointer is non-nil.
func subspecCondIndexed(subspecID *int64, index int) string {
	if subspecID == nil {
		return "subspec_id IS NULL"
	}
	return "subspec_id = $" + strconv.Itoa(index)
}

// Returns a condition matching the given parentID,
// adding the given parentID to args if not already present.
func parentCond(parentID *int64, args *[]interface{}) string {
	if parentID == nil {
		return "parent_id IS NULL"
	}
	return "parent_id = " + argPlaceholder(*parentID, args)
}

// Returns a condition matching the given parentID,
// using a fixed placeholder index when pointer is non-nil.
func parentCondIndexed(parentID *int64, index int) string {
	if parentID == nil {
		return "parent_id IS NULL"
	}
	return "parent_id = $" + strconv.Itoa(index)
}
