package goscanql

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
)

const (
	scanqlTag = "goscanql"
)

func evaluateStruct(obj interface{}) (map[string]interface{}, error) {

	rv := reflect.ValueOf(obj)

	// if not pointer error as cannot affect the original obj
	if rv.Kind() != reflect.Pointer {
		return nil, errors.New("provided obj must be a pointer")
	}

	// unwrap pointer
	rv = rv.Elem()
	t := rv.Type()

	expectedFields := make(map[string]interface{})

	// extract expected fields
	for i := 0; i < t.NumField(); i++ {

		field, ok := t.Field(i).Tag.Lookup(scanqlTag)
		if !ok {
			// skip if field doesn't have scanql tag
			continue
		}

		expectedFields[field] = rv.Field(i).Addr().Interface()
	}

	return expectedFields, nil
}

func mapFieldsToColumns(cols []string, fields map[string]interface{}) []interface{} {

	values := make([]interface{}, len(cols))

	for i, col := range cols {

		value, ok := fields[col]
		if !ok {
			continue
		}

		values[i] = value
	}

	return values
}

func scanRows[T any](rows *sql.Rows) ([]T, error) {

	result := make([]T, 0)

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	for rows.Next() {

		entry := new(T)

		fields, err := evaluateStruct(entry)
		if err != nil {
			return nil, err
		}

		err = rows.Scan(mapFieldsToColumns(cols, fields)...)
		if err != nil {
			return nil, err
		}

		result = append(result, *entry)
	}

	return result, nil
}

// RowsToStructs will take the data in rows (*sql.Rows) as input and return a slice of
// Ts (the provided type) as the result.
func RowsToStructs[T any](rows *sql.Rows) ([]T, error) {
	return scanRows[T](rows)
}

// RowToStruct will take the data in rows (*sql.Rows) as input (similarly to RowsToStructs)
// and return a single T (the provided type) as the result and error if more or less than 1
// row is present.
func RowToStruct[T any](rows *sql.Rows) (T, error) {

	var zero T // effectively nil (as type is unknown, we can't just return nil)

	result, err := scanRows[T](rows)
	if err != nil {
		return zero, err
	}

	if len(result) != 1 {
		return zero, fmt.Errorf("rows had a non-zero length: %d", len(result))
	}

	return result[0], nil
}
