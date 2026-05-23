package repository

import "database/sql/driver"

func driverValues(args []any) []driver.Value {
	vals := make([]driver.Value, len(args))
	for i, a := range args {
		vals[i] = a
	}
	return vals
}
