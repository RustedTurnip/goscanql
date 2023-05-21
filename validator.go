package goscanql

import (
	"database/sql"
	"fmt"
	"reflect"
)

var (
	// structVerifiers maintains all assertions that must be made on the raw input type provided
	// by the user to goscanql.
	structVerifiers = []func(t reflect.Type) error{

		// input type must be struct
		func(t reflect.Type) error {
			t = getPointerRootType(t)

			if t.Kind() == reflect.Struct {
				return nil
			}

			return fmt.Errorf("input type (%s) must be of type struct or pointer to struct", t.String())
		},
	}

	// fieldVerifiers maintians all assertions that must be made on both the raw input type and
	// any relevant type child fields for goscanql to be able to work.
	fieldVerifiers = []func(t reflect.Type) error{

		// arrays
		func(t reflect.Type) error {

			t = getPointerRootType(t)

			if t.Kind() != reflect.Array {
				return nil
			}

			return fmt.Errorf("arrays are not supported, consider using a slice or scanner implementation instead")
		},

		// maps
		func(t reflect.Type) error {

			t = getPointerRootType(t)

			if t.Kind() != reflect.Map {
				return nil
			}

			return fmt.Errorf("maps are not supported, consider using a slice or scanner implementation instead")
		},

		// multi-dimensional slice
		func(t reflect.Type) error {

			t = getPointerRootType(t)

			if t.Kind() != reflect.Slice {
				return nil
			}

			// get slice type, e.g. []*Example has a slice type of *Example
			sliceType := t.Elem()

			// get root type (e.g. ***[]string has root type of []string) and assert that it isn't a slice
			if getPointerRootType(sliceType).Kind() != reflect.Slice {
				return nil
			}

			return fmt.Errorf("multi-dimensional slices are not supported (%s), consider using a slice or scanner implementation instead", t.String())
		},
	}
)

// TODO
func verifyType[T any]() error {

	// initialise empty instance of type T so we can evaluate it's type
	var zero T
	t := reflect.TypeOf(zero)

	// run checks on input type
	for _, verifier := range structVerifiers {
		err := verifier(t)
		if err != nil {
			return err
		}
	}

	// assert no cyclic-structs
	// NOTE: this check must happen before the fieldVerifiers check as if there is a cyclic
	// struct, the fieldVerifiers check will end up in infinite recursion
	err := verifyNoCycles(t)
	if err != nil {
		return err
	}

	// run checks on all child-types of input type (and additional checks on input type)
	for _, verifier := range fieldVerifiers {
		err := traverseType(t, verifier)
		if err != nil {
			return err
		}
	}

	return nil
}

// TODO
func getPointerRootType(t reflect.Type) reflect.Type {

	if t.Kind() != reflect.Pointer {
		return t
	}

	return getPointerRootType(t.Elem())
}

// TODO
func getRootSliceType(t reflect.Type) reflect.Type {

	t = getPointerRootType(t)

	if t.Kind() != reflect.Slice {
		return t
	}

	// pass forward slice type, e.g. []*Example has a slice type of *Example
	return getRootSliceType(t.Elem())
}

// TODO
func verifyNoCycles(t reflect.Type) error {

	t = getPointerRootType(t)

	if t.Kind() != reflect.Struct {
		return nil
	}

	cyclic := hasCycle(t, map[reflect.Type]interface{}{})
	if !cyclic {
		return nil
	}

	return fmt.Errorf("goscanql does not support cyclic structs: %s", t.String())
}

// TODO mention only a struct may be provided
func hasCycle(t reflect.Type, m map[reflect.Type]interface{}) bool {

	m[t] = struct{}{}
	defer delete(m, t)

	for i := 0; i < t.NumField(); i++ {

		if !isGoscanqlField(t.Field(i)) {
			continue
		}

		fieldType := getRootSliceType(t.Field(i).Type)

		if isScanner(fieldType) {
			continue
		}

		if fieldType.Kind() != reflect.Struct {
			continue
		}

		_, ok := m[fieldType]
		if ok {
			return true
		}

		cyclic := hasCycle(fieldType, m)
		if cyclic {
			return true
		}
	}

	return false
}

// TODO
func isGoscanqlField(f reflect.StructField) bool {
	_, b := f.Tag.Lookup(scanqlTag)
	return b
}

// TODO
func isScanner(t reflect.Type) bool {
	iScanner := reflect.TypeOf((*sql.Scanner)(nil)).Elem()
	return t.Implements(iScanner)
}

// TODO
func traverseType(t reflect.Type, f func(t reflect.Type) error) error {

	t = getPointerRootType(t)

	// check input's type for compatibility
	err := f(t)
	if err != nil {
		return err
	}

	// if slice, evaluate slices sub-type
	if t.Kind() == reflect.Slice {
		return traverseType(getRootSliceType(t), f)
	}

	// if type isn't traversable (as it isn't a slice or struct) we have reached end of branch traversal
	if t.Kind() != reflect.Struct {
		return nil
	}

	// if struct, traverse each sub-field
	for i := 0; i < t.NumField(); i++ {

		// if the field isn't tagged as goscanql, ignore
		if !isGoscanqlField(t.Field(i)) {
			continue
		}

		// if type is scanner, it will have custom Scan logic that supersedes goscanql
		if isScanner(t) {
			return nil
		}

		// traverse field's subtypes
		err := traverseType(t.Field(i).Type, f)
		if err != nil {
			return err
		}
	}

	return nil
}
