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

func evaluate(obj interface{}) (*fields, error) {

	expectedFields := newFields()

	err := evaluateStruct(obj, expectedFields)
	if err != nil {
		return nil, err
	}

	return expectedFields, nil
}

func evaluateStruct(obj interface{}, fields *fields) error {

	rv := reflect.ValueOf(obj)

	// if not pointer error as cannot affect the original obj
	if rv.Kind() != reflect.Pointer {
		return errors.New("provided obj must be a pointer")
	}

	// if type is slice, add 1 element to it to store values
	// TODO walk through pointers until a value is found to write to
	// TODO (see: https://stackoverflow.com/questions/35604356/json-unmarshal-accepts-a-pointer-to-a-pointer)
	if rv.Elem().Kind() == reflect.Slice {

		// get slice type, e.g. []*Example has a slice type of *Example
		sliceType := reflect.TypeOf(rv.Elem().Interface()).Elem()

		// create new element of sliceType
		element := reflect.New(sliceType).Elem()

		// if underlying type is a pointer, instantiate
		if element.Kind() == reflect.Pointer {
			element.Set(reflect.New(sliceType.Elem()))
		}

		// append new element to slice
		rv.Elem().Set(reflect.Append(rv.Elem(), element))

		// if element is pointer, pass it through directly
		if element.Kind() == reflect.Pointer {
			return evaluateStruct(rv.Elem().Index(0).Interface(), fields)
		}
		// else create pointer to it to pass through
		return evaluateStruct(rv.Elem().Index(0).Addr().Interface(), fields)
	}

	// unwrap pointer
	rv = rv.Elem()
	t := rv.Type()

	// extract expected fields
	for i := 0; i < t.NumField(); i++ {

		fieldType := t.Field(i)
		fieldValue := rv.Field(i)

		fieldName, ok := fieldType.Tag.Lookup(scanqlTag)
		if !ok {
			// skip if field doesn't have scanql tag
			continue
		}

		// if pointer
		if fieldValue.Kind() == reflect.Pointer {
			rv.Field(i).Set(reflect.New(fieldType.Type.Elem()))

			// evaluate with pointer to new instance
			err := evaluateStruct(rv.Field(i).Interface(), fields.addChild(fieldName))
			if err != nil {
				return err
			}

			continue
		}

		// if nested struct or slice
		if fieldValue.Kind() == reflect.Struct || fieldValue.Kind() == reflect.Slice {

			// set current field to new instance of field type
			rv.Field(i).Set(reflect.New(fieldValue.Type()).Elem())

			// evaluate with pointer to new instance
			err := evaluateStruct(fieldValue.Addr().Interface(), fields.addChild(fieldName))
			if err != nil {
				return err
			}

			continue
		}

		// add field to map
		fmt.Println(fieldValue.Addr().Pointer())
		fields.addField(fieldName, rv.Field(i).Addr().Interface())
	}

	return nil
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

		fields, err := evaluate(entry)
		if err != nil {
			return nil, err
		}

		err = rows.Scan(mapFieldsToColumns(cols, fields.getFieldReferences())...)
		if err != nil {
			return nil, err
		}

		result = append(result, *entry)
	}

	return result, nil
}

// instantiateAndReturn will take any value and instantiate it with the equivalent Zero
// value for that type, e.g. 0 for int or an empty struct for a struct. It will then return
// that value as a reflect.Value.
//
// If the type is a pointer (at any level, e.g. *int or ****int) the function will traverse
// to the very root of the pointers (in this case to the int) and instantiate and return
// that. The original pointers will be set to point to this new value also.
//
// Note, if the pointer is uninitialised, to keep a reference to it you will need to pass
// it in as a pointer, for example:
//
//	var i *int
//
// would need to be passed in as
//
//	instantiateAndReturn(&i)
//
// as the default value would be nil, and therefore is not addressable. However if the pointer
// is initialised e.g. ip in this example:
//
//	var i int
//	ip := &i
//
// then that can be passed in directly:
//
//	instantiateAndReturn(ip)
func instantiateAndReturn[T any](t T) reflect.Value {
	return instantiateValue(reflect.ValueOf(t).Elem())
}

func instantiateValue(val reflect.Value) reflect.Value {

	// get value of i (must pass in as pointer), see:
	// https://stackoverflow.com/questions/34145072/can-you-initialise-a-pointer-variable-with-golang-reflect

	// if we are not at root
	if val.Kind() == reflect.Pointer {

		// instantiate current value
		val.Set(reflect.New(val.Type().Elem()))

		// crawl further
		return instantiateValue(val.Elem())
	}

	return val
}

// TODO this func will group the rows into the correct struct arrays and fields
func aggregateStructs[T any](rows []T) {

	// TODO pick up from here
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
