package main

import (
	"database/sql"
	"strconv"
)

// DBConn wraps a *sql.DB or *sql.Tx.
type DBConn interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
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

// Returns the "= $N" or "IS NULL" part of an equality condition
// where the operand may be null.
func eqCond(col string, arg interface{}, args *[]interface{}) string {
	if isNil(arg) {
		return col + " IS NULL"
	}
	return col + " = " + argPlaceholder(arg, args)
}

// Returns the "= $N" or "IS NULL" part of an equality condition
// where the operand may be null, and the placeholder index is given.
func eqCondIndexed(col string, arg interface{}, index int) string {
	if isNil(arg) {
		return col + " IS NULL"
	}
	return col + " = $" + strconv.Itoa(index)
}

func createArgsList(args *[]interface{}, values ...interface{}) string {
	out := ``
	for i := 0; i < len(values); i++ {
		if i > 0 {
			out += `,`
		}
		out += argPlaceholder(values[i], args)
	}
	return out
}

func createArgsListInt64s(args *[]interface{}, values ...int64) string {
	out := ``
	for i := 0; i < len(values); i++ {
		if i > 0 {
			out += `,`
		}
		out += argPlaceholder(values[i], args)
	}
	return out
}

func createIDsValuesMap(args *[]interface{}, ids []int64, values ...[]interface{}) string {
	out := `VALUES `

	for i := 0; i < len(ids); i++ {
		if i > 0 {
			out += `,`
		}
		// include types with placeholders to be clear
		out += `(` + argPlaceholder(ids[i], args) + `::int`
		for j := 0; j < len(values); j++ {
			v := values[j][i]
			out += `,` + argPlaceholder(v, args)
			switch v.(type) {
			case int, uint, int64:
				out += `::int`
			}
		}
		out += `)`
	}

	return out
}
