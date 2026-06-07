package repository

import "database/sql/driver"

type listQueries struct {
	withQ       string
	noFilter    string
	withStatus  string
	withQStatus string
}

func driverValues(args []any) []driver.Value {
	vals := make([]driver.Value, len(args))
	for i, a := range args {
		vals[i] = a
	}
	return vals
}

func buildFilteredListQuery(queries listQueries, limit int, offset int, q string, status string) (string, []any) {
	searchPattern := "%" + q + "%"

	switch {
	case q != "" && status != "":
		return queries.withQStatus, []any{searchPattern, status, limit, offset}
	case q != "":
		return queries.withQ, []any{searchPattern, limit, offset}
	case status != "":
		return queries.withStatus, []any{status, limit, offset}
	default:
		return queries.noFilter, []any{limit, offset}
	}
}
