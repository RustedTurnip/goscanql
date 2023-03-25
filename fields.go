package goscanql

import (
	"fmt"
	"reflect"
)

type fields struct {
	objRef            interface{}
	orderedFieldNames []string
	references        map[string]interface{}
	children          map[string]*fields
}

func (f *fields) addNewChild(name string, obj interface{}) error {
	if _, ok := f.children[name]; ok {
		panic(fmt.Errorf("child with same name (\"%s\") already exists", name))
	}

	child, err := newFields(obj)
	if err != nil {
		return err
	}

	f.children[name] = child
	return nil
}

func (f *fields) addField(name string, value interface{}) {

	// assert that field hasn't already been added
	if _, ok := f.references[name]; ok {
		panic(fmt.Errorf("field with name \"%s\" already added", name))
	}

	// add field to this instance
	f.orderedFieldNames = append(f.orderedFieldNames, name)
	f.references[name] = value
}

func (f *fields) getFieldReferences() map[string]interface{} {

	m := make(map[string]interface{})

	f.crawlReferences(func(key string, value interface{}) {
		m[key] = value
	})

	return m
}

func (f *fields) getFieldByteReferences() map[string]*[]byte {

	m := make(map[string]*[]byte)

	f.crawlReferences(func(key string, value interface{}) {
		m[key] = &[]byte{}
	})

	return m
}

func (f *fields) crawlReferences(fn func(key string, value interface{})) {
	f.crawlReferencesWithPrefix("", fn)
}

func (f *fields) crawlReferencesWithPrefix(prefix string, fn func(key string, value interface{})) {

	// if there is a prefix, format it accordingly
	if prefix != "" {
		prefix = fmt.Sprintf("%s_", prefix)
	}

	// for each field, run callback (fn)
	for name, reference := range f.references {
		fn(fmt.Sprintf("%s%s", prefix, name), reference)
	}

	// crawl through children and repeat
	for name, child := range f.children {
		childPrefix := fmt.Sprintf("%s%s", prefix, name)
		child.crawlReferencesWithPrefix(childPrefix, fn)
	}
}

func newFields(obj interface{}) (*fields, error) {

	fields := &fields{
		orderedFieldNames: make([]string, 0),
		references:        make(map[string]interface{}),
		children:          make(map[string]*fields),
	}

	err := initialiseFields(obj, fields)
	if err != nil {
		return nil, err
	}

	return fields, nil
}

func initialiseFields(obj interface{}, fields *fields) error {

	fields.objRef = obj

	rv := instantiateAndReturn(obj)
	t := rv.Type()

	// if type is slice, add 1 element to it to store values
	if rv.Kind() == reflect.Slice {

		// get slice type, e.g. []*Example has a slice type of *Example
		sliceType := reflect.TypeOf(rv.Interface()).Elem()

		// create new element of sliceType
		element := reflect.New(sliceType).Elem()

		// instantiate element's root value
		elementValue := instantiateAndReturn(element.Addr().Interface())

		// append new element to slice
		rv.Set(reflect.Append(rv, element))

		// add child and evaluate it
		return initialiseFields(elementValue.Addr().Interface(), fields)
	}

	// extract expected fields
	for i := 0; i < t.NumField(); i++ {

		fieldType := t.Field(i)
		fieldValue := rv.Field(i)

		fieldName, ok := fieldType.Tag.Lookup(scanqlTag)

		// skip if field doesn't have scanql tag
		if !ok {
			continue
		}

		fieldValueRoot := instantiateAndReturn(fieldValue.Addr().Interface())

		// if nested struct or slice
		if fieldValueRoot.Kind() == reflect.Struct || fieldValueRoot.Kind() == reflect.Slice {

			// evaluate with pointer to new instance
			err := fields.addNewChild(fieldName, fieldValueRoot.Addr().Interface())
			if err != nil {
				return err
			}

			continue
		}

		// add field to map
		fields.addField(fieldName, rv.Field(i).Addr().Interface())
	}

	return nil
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
	fmt.Println(reflect.TypeOf(t))
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
