package goscanql

import (
	"fmt"
	"reflect"
)

type typeValidator func(t reflect.Type) error

var (
	// structValidators maintains all assertions that must be made on the raw input type provided
	// by the user to goscanql.
	structValidators = []typeValidator{
		isStruct,
	}

	// fieldValidators maintains all assertions that must be made on both the raw input type and
	// any relevant type child fields for goscanql to be able to work.
	fieldValidators = []typeValidator{
		isNotArray,
		isNotMap,
		isNotMultidimensionalSlice,
		isNotFunc,
		isNotChan,
		isNotCustomInterface,
	}
)

var (
	// genericInterfaceType is the type of interface {} and is used for interface type
	// comparisons.
	//
	// Note: a pointer to interface {} (*interface {}) is used because: "As interface types
	// are only used for static typing, a common idiom to find the reflection Type for an
	// interface type Foo is to use a *Foo value."
	// Ref: https://stackoverflow.com/a/34698753
	genericInterfaceType = reflect.TypeOf((*interface{})(nil)).Elem()
)

// isStruct takes a reflect.Type (t) and returns an error if it is not a struct (or nil
// otherwise).
func isStruct(t reflect.Type) error {
	t = getPointerRootType(t)

	if t.Kind() == reflect.Struct {
		return nil
	}

	return fmt.Errorf("input type (%s) must be of type struct or pointer to struct", t.String())
}

// isNotArray takes a reflect.Type (t) and returns an error if it is an array (or nil
// otherwise).
func isNotArray(t reflect.Type) error {
	t = getPointerRootType(t)

	// if type (or pointer to type) implements Scanner, then it is exempt
	if asScanner(reflect.New(t)) != nil {
		return nil
	}

	// recursively search slice types until base type found
	if t.Kind() == reflect.Slice {
		// before assessing nested type to see if it is an array, we must make sure that the nested
		// type isn't exempt from the validations by implementing Scanner.
		if implementsScanner(t) {
			return nil
		}

		return isNotArray(t.Elem())
	}

	if t.Kind() != reflect.Array {
		return nil
	}

	return fmt.Errorf("arrays are not supported (%s), consider using a slice instead", t.String())
}

// isNotMap takes a reflect.Type (t) and returns an error if it is a map (or nil
// otherwise).
func isNotMap(t reflect.Type) error {
	t = getPointerRootType(t)

	// if type (or pointer to type) implements Scanner, then it is exempt
	if asScanner(reflect.New(t)) != nil {
		return nil
	}

	// recursively search array/slice types until base type found
	if t.Kind() == reflect.Array || t.Kind() == reflect.Slice {
		return isNotMap(t.Elem())
	}

	if t.Kind() != reflect.Map {
		return nil
	}

	return fmt.Errorf("maps are not supported (%s), consider using a slice instead", t.String())
}

// isNotMultidimensionalSlice takes a reflect.Type (t) and returns an error if
// it is a multi-dimensional slice (or nil otherwise).
func isNotMultidimensionalSlice(t reflect.Type) error {
	t = getPointerRootType(t)

	// if type (or pointer to type) implements Scanner, then it is exempt
	if asScanner(reflect.New(t)) != nil {
		return nil
	}

	if t.Kind() != reflect.Slice {
		return nil
	}

	// get slice type, e.g. []*Example has a slice type of *Example
	sliceType := t.Elem()

	// get root type (e.g. ***[]string has root type of []string) and assert that it isn't a slice
	if getPointerRootType(sliceType).Kind() != reflect.Slice {
		return nil
	}

	// if sliceType (or pointer to type) implements Scanner, then it is exempt
	if asScanner(reflect.New(sliceType)) != nil {
		return nil
	}

	return fmt.Errorf("multi-dimensional slices are not supported (%s), consider using a slice instead", t.String())
}

// isNotFunc takes a reflect.Type (t) and returns an error if it is a function (or
// the nested type if it is a slice/array) or nil otherwise.
func isNotFunc(t reflect.Type) error {
	t = getPointerRootType(t)

	// if type (or pointer to type) implements Scanner, then it is exempt
	if asScanner(reflect.New(t)) != nil {
		return nil
	}

	// recursively search array/slice types until base type found
	if t.Kind() == reflect.Array || t.Kind() == reflect.Slice {
		return isNotFunc(t.Elem())
	}

	if t.Kind() != reflect.Func {
		return nil
	}

	return fmt.Errorf("functions are not supported (%s)", t.String())
}

// isNotChan takes a reflect.Type (t) and returns an error if it is a chan (or
// the nested type if it is a slice/array) or nil otherwise.
func isNotChan(t reflect.Type) error {
	t = getPointerRootType(t)

	// if type (or pointer to type) implements Scanner, then it is exempt
	if asScanner(reflect.New(t)) != nil {
		return nil
	}

	// recursively search array/slice types until base type found
	if t.Kind() == reflect.Array || t.Kind() == reflect.Slice {
		return isNotChan(t.Elem())
	}

	if t.Kind() != reflect.Chan {
		return nil
	}

	return fmt.Errorf("channels are not supported (%s)", t.String())
}

// isNotCustomInterface takes a type and returns an error when the type is an interface other than
// interface {}.
func isNotCustomInterface(t reflect.Type) error {
	t = getPointerRootType(t)

	// recursively search array/slice types until base type found
	if t.Kind() == reflect.Array || t.Kind() == reflect.Slice {
		return isNotCustomInterface(t.Elem())
	}

	if t.Kind() != reflect.Interface {
		return nil
	}

	// goscanql supports general interfaces, just not custom ones
	if t == genericInterfaceType {
		return nil
	}

	return fmt.Errorf("interface types other than interface{} are not supported (%s)", t.String())
}

// validateType analyses the provided input type and ensures that it will is valid based on
// goscanql's input rules (including no cyclic structs).
func validateType(it interface{}) error {
	t := reflect.TypeOf(it)

	// run checks on input type
	for _, validator := range structValidators {
		err := validator(t)
		if err != nil {
			return err
		}
	}

	// assert no cyclic-structs
	// NOTE: this check must happen before the fieldValidators check as if there is a cyclic
	// struct, the fieldValidators check will end up in infinite recursion
	err := verifyNoCycles(t)
	if err != nil {
		return err
	}

	// run checks on all child-types of input type (and additional checks on input type)
	for _, validator := range fieldValidators {
		err := traverseType(t, validator)
		if err != nil {
			return err
		}
	}

	return nil
}

// getPointerRootType takes a reflect.Type (t) as input and returns the innermost type
// that is not a pointer.
//
// For example, ****[]string returns []string, **[]*string returns []*string and so on.
func getPointerRootType(t reflect.Type) reflect.Type {
	if t.Kind() != reflect.Pointer {
		return t
	}

	return getPointerRootType(t.Elem())
}

// getSliceRootType takes a reflect.Type (t) as input and returns the first non-slice
// type.
//
// NOTE: pointers to slices are treated as slices, but slices to pointers of
// non-slices, are left as pointers.
//
// For example, **[]*[]string would return string, but **[]*[]*string would return
// *string as the type (leaving the pointer on the string type even though the
// pointers to slices have been treated as slices).
func getSliceRootType(t reflect.Type) reflect.Type {
	raw := getPointerRootType(t)

	if raw.Kind() != reflect.Slice {
		return t
	}

	// pass forward slice type, e.g. []*Example has a slice type of *Example
	return getSliceRootType(raw.Elem())
}

// verifyNoCycles takes a reflect.Type (t) and analyses it for cycles (where a struct
// maintains an internal reference to the same struct).
//
// NOTE: this function assumes that t is a struct type, any other type will result in
// a panic.
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

// hasCycle implements a recursive crawl that traverses the children of the provided
// reflect.Type (t) and looks for any struct cycles (where a struct type has a field
// of its own type - this could be a field of a field).
//
// NOTE: this function assumes that t is a struct type, any other type will result in
// a panic.
func hasCycle(t reflect.Type, m map[reflect.Type]interface{}) bool {
	m[t] = struct{}{}
	defer delete(m, t)

	for i := 0; i < t.NumField(); i++ {
		if !isGoscanqlField(t.Field(i)) {
			continue
		}

		fieldType := getSliceRootType(t.Field(i).Type) // strip away slices
		fieldType = getPointerRootType(fieldType)      // strip away pointers

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

// isGoscanqlField takes a reflect.Field (f) and evaluates whether it is a field
// designated for goscanql or not (meaning the parent struct has it tagged with
// `sql:"tag_name"`). If so, true is returned, otherwise false.
func isGoscanqlField(f reflect.StructField) bool {
	_, b := f.Tag.Lookup(scanqlTag)
	return b
}

// traverseType will recursively traverse the children of the provided type and
// run the provided func (f) on each child field. This function provides a generic
// way to traverse the fields of a struct.
//
// If a non-struct type is provided, the function will be run on the provided type
// and return immediately (as there are now more fields to traverse).
func traverseType(t reflect.Type, f func(t reflect.Type) error) error {
	t = getPointerRootType(t)

	// check input's type for compatibility
	err := f(t)
	if err != nil {
		return err
	}

	// if slice, evaluate slices sub-type
	if t.Kind() == reflect.Slice {
		return traverseType(getSliceRootType(t), f)
	}

	// if type isn't traversable (as it isn't a slice or struct) we have reached end of branch traversal
	if t.Kind() != reflect.Struct {
		return nil
	}

	// if struct, traverse each sub-field
	for i := 0; i < t.NumField(); i++ {
		// if the field isn't tagged for goscanql, ignore
		if !isGoscanqlField(t.Field(i)) {
			continue
		}

		// traverse field's subtypes
		err := traverseType(t.Field(i).Type, f)
		if err != nil {
			return err
		}
	}

	return nil
}
