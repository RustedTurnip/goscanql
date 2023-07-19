package goscanql

import (
	"crypto/sha1"
	"fmt"
	"reflect"
	"strings"
	"time"
)

// nullBytes represents a scannable entity that can be used to determine if the incoming value is
// nil (but does not store the value).
type nullBytes struct {
	isNil bool
}

// Scan is required to implement the Scan interface for reading SQL rows into fields. This function
// will assess whether the inbound value is nil or not, but doesn't store the value itself.
func (n *nullBytes) Scan(value interface{}) error {
	n.isNil = value == nil
	return nil
}

func newNullBytes() *nullBytes {
	return &nullBytes{
		isNil: true,
	}
}

// fields holds a goscanql parsed struct, maintaining references to the fields
// of the struct and any sub-structs (children).
type fields struct {

	// obj is a reference (pointer) to the struct that this fields fields belong to.
	obj interface{}

	// orderedFieldNames maintains the field names in the order of which they were added
	// to facilitate reliable hashing when comparing fields entities.
	orderedFieldNames []string

	// orderedScannerNames maintains the scanner field names in the order of which they were
	// added to facilitate reliable hashing when comparing fields entities.
	orderedScannerNames []string

	// orderedOneToOneNames maintains the names of the one-to-one relationship children so
	// that they can reliably be hashed for comparison.
	orderedOneToOneNames []string

	// references holds a reference to each field belonging to a fields entity so they can
	// be set.
	references map[string]interface{}

	// scannerReferences holds a reference to each scanner belonging to a fields entity so
	// they can be set.
	scannerReferences map[string]Scanner

	// nullFields holds a nullBytes entity for each field and is used to determine whether a
	// field is nil or not.
	nullFields map[string]*nullBytes

	// oneToOnes holds all child structs of the fields entity that are maintained as a
	// one-to-one relationship.
	oneToOnes map[string]*fields

	// oneToManys holds all child structs of the fields entity that are maintained as a
	// one-to-many relationship (meaning the sub-struct is contained within a slice).
	oneToManys map[string]*fields
}

// addNewChild will create a new fields entity and add it to the current fields as a child
// in either a one-to-one relationship, or a one-to-many relationship based on the type
// of obj.
//
// Note: obj must be a reference to the object, e.g. of type *Struct, or *[]Struct.
func (f *fields) addNewChild(name string, obj interface{}) error {

	rv := reflect.ValueOf(obj)

	// create new fields instance
	child, err := newFields(obj)
	if err != nil {
		return err
	}

	// ensure that child with name doesn't already exist
	collisionErr := fmt.Errorf("child already exists with name \"%s\"", name)

	for childName := range f.oneToOnes {
		if childName == name {
			return collisionErr
		}
	}

	for childName := range f.oneToManys {
		if childName == name {
			return collisionErr
		}
	}

	// add child to appropriate relationship map of fields
	if rv.Elem().Kind() == reflect.Slice {
		f.oneToManys[name] = child
		return nil
	}

	f.oneToOnes[name] = child
	f.orderedOneToOneNames = append(f.orderedOneToOneNames, name)
	return nil
}

func newFieldCollisionError(fieldName string) error {
	return fmt.Errorf("field with name \"%s\" already added", fieldName)
}

// addField will add a single field to the current fields (e.g. a string or int).
func (f *fields) addField(name string, value interface{}) error {

	// assert that field hasn't already been added
	if _, ok := f.references[name]; ok {
		return newFieldCollisionError(name)
	}

	// assert that field hasn't already been added as scanner
	if _, ok := f.scannerReferences[name]; ok {
		return newFieldCollisionError(name)
	}

	// add field to this instance
	f.orderedFieldNames = append(f.orderedFieldNames, name)
	f.references[name] = value
	f.nullFields[name] = newNullBytes()

	return nil
}

// addScanner will add a single scanner to the current scanners.
func (f *fields) addScanner(name string, value Scanner) error {

	// assert that scanner hasn't already been added as field
	if _, ok := f.references[name]; ok {
		return newFieldCollisionError(name)
	}

	// assert that scanner hasn't already been added
	if _, ok := f.scannerReferences[name]; ok {
		return newFieldCollisionError(name)
	}

	// add field to this instance
	f.orderedScannerNames = append(f.orderedScannerNames, name)
	f.scannerReferences[name] = value
	f.nullFields[name] = newNullBytes()

	return nil
}

// getFieldReferences returns a map of all of the fields references (including any child
// field references).
func (f *fields) getFieldReferences() map[string]interface{} {

	m := make(map[string]interface{})

	f.crawlFields(func(prefix string, fi *fields) bool {

		if fi.isNil() {
			return true
		}

		for name, reference := range fi.references {
			m[buildReferenceName(prefix, name)] = reference
		}

		for name, scanner := range fi.scannerReferences {
			m[buildReferenceName(prefix, name)] = scanner
		}

		return false
	})

	return m
}

// getNullFieldReferences returns a map of all of the null fieldreferences (including any child
// references).
func (f *fields) getNullFieldReferences() map[string]*nullBytes {

	m := make(map[string]*nullBytes)

	f.crawlFields(func(prefix string, fi *fields) bool {

		for name, reference := range fi.nullFields {
			m[buildReferenceName(prefix, name)] = reference
		}

		return false
	})

	return m
}

// crawlFields will recursively iterate of each field of each fields and its children.
func (f *fields) crawlFields(fn func(string, *fields) bool) {
	f.crawlFieldsWithPrefix("", fn)
}

// crawlFields will recursively iterate of each field of each fields and its children
// with the added context of the prefix field which is used to reference child fields.
func (f *fields) crawlFieldsWithPrefix(prefix string, fn func(string, *fields) bool) bool {

	// if cancel signalled, return and don't bother processing this field's children
	if fn(prefix, f) {
		return true
	}

	// crawl each one-to-one child
	for name, child := range f.oneToOnes {
		child.crawlFieldsWithPrefix(buildReferenceName(prefix, name), fn)
	}

	// crawl each one-to-many child
	for name, child := range f.oneToManys {
		child.crawlFieldsWithPrefix(buildReferenceName(prefix, name), fn)
	}

	return false
}

// buildReferenceName will put together a field reference name based on the provided
// prefix, and the field's name, e.g.
//
// Prefix: pet, Name: animal := pet_animal
func buildReferenceName(prefix, name string) string {

	strs := make([]string, 0)

	if prefix != "" {
		strs = append(strs, prefix)
	}

	if name != "" {
		strs = append(strs, name)
	}

	return strings.Join(strs, "_")
}

// getHash will hash a fields entity so that it can be easily compared to another fields.
func (f *fields) getHash() string {

	raw := f.getBytePrint("")

	// hash fields to create unique id for struct
	h := sha1.New()
	h.Write(raw)

	return string(h.Sum(nil))
}

// getBytePrint will return a "fingerprint" of the current fields entity and it's one-to-one
// children as an array of bytes.
func (f *fields) getBytePrint(prefix string) []byte {

	print := make([]byte, 0)

	for _, key := range f.orderedFieldNames {

		value := f.references[key]
		strValue := fmt.Sprintf("{%s:%#v}", buildReferenceName(prefix, key), reflect.ValueOf(value).Elem().Interface())
		print = append(print, []byte(strValue)...)
	}

	for _, key := range f.orderedScannerNames {

		value := f.scannerReferences[key]
		strValue := fmt.Sprintf("{%s:%s}", buildReferenceName(prefix, key), value.GetID())
		print = append(print, []byte(strValue)...)
	}

	for _, key := range f.orderedOneToOneNames {

		child := f.oneToOnes[key]
		print = append(print, child.getBytePrint(key)...)
	}

	return print
}

// isNil will the incoming data to a fields (once it has been written to the nullFields)
// to see if the object that the fields represents will be nil.
func (f *fields) isNil() bool {

	for _, b := range f.nullFields {
		if !b.isNil {
			return false
		}
	}

	return true
}

// isMatch will compare the provided fields (m) to the current fields to see if they are equal
// in value, returning true if they are, and false otherwise.
func (f *fields) isMatch(m *fields) bool {
	return f.getHash() == m.getHash()
}

func (f *fields) emptyNilFields() {

	if f.isNil() {
		// empty the object represented by the fields, e.g. *int would be set to nil,
		// or int would be set to 0.
		rv := reflect.ValueOf(f.obj).Elem()
		rv.Set(reflect.New(rv.Type()).Elem())
		return
	}

	for _, child := range f.oneToOnes {
		child.emptyNilFields()
	}

	for tag, child := range f.oneToManys {
		if !child.isNil() {
			child.emptyNilFields()
			continue
		}

		slice := getRootValue(*fieldByTag(tag, getRootValue(reflect.ValueOf(f.obj))))
		slice.Set(reflect.New(slice.Type()).Elem()) // set to empty slice
	}

}

// scan will attempt to apply the provided scan function to the fields object
// by providing it with all the field references so that values can be written.
func (f *fields) scan(columns []string, scan func(...interface{}) error) error {

	byteRefs := mapFieldsToColumns(columns, f.getNullFieldReferences())

	err := scan(byteRefs...)
	if err != nil {
		return err
	}

	refs := mapFieldsToColumns(columns, f.getFieldReferences())

	err = scan(refs...)
	if err != nil {
		return err
	}

	f.emptyNilFields()
	return nil
}

// newFields is the fields constructor that will process the provided object, and use
// reflection to map it out and maintain references to the object's fields.
func newFields(obj interface{}) (*fields, error) {

	// instantiate root of obj to create fields around
	rva := instantiateAndReturnAll(obj)
	rv := rva[0]

	// if the obj is a slice, we must make obj represent an element of the slice instead of
	// the slice itself as slices are the basis for one-to-many relationships
	if rv.Kind() == reflect.Slice {

		// get slice type, e.g. []*Example has a slice type of *Example
		sliceType := reflect.TypeOf(rv.Interface()).Elem()

		// create new element of sliceType
		element := reflect.New(sliceType).Elem()

		// instantiate element's root value
		instantiateAndReturnRoot(element.Addr().Interface())

		// append new element to slice
		rv.Set(reflect.Append(rv, element))

		// point object to newly created 0th element of slice
		obj = rv.Index(0).Addr().Interface()
	}

	// create new fields
	fields := &fields{
		obj:                  obj,
		orderedFieldNames:    make([]string, 0),
		orderedScannerNames:  make([]string, 0),
		orderedOneToOneNames: make([]string, 0),
		references:           make(map[string]interface{}),
		scannerReferences:    make(map[string]Scanner),
		nullFields:           make(map[string]*nullBytes),
		oneToOnes:            make(map[string]*fields),
		oneToManys:           make(map[string]*fields),
	}

	// initialise the newly created fields around the obj being pointed to
	err := fields.initialise("")
	if err != nil {
		return nil, err
	}

	return fields, nil
}

// initialise uses reflection to map it out and maintain references to the object's
// fields.
func (f *fields) initialise(prefix string) error {

	rva := instantiateAndReturnAll(f.obj)

	rv := rva[0]
	t := rv.Type()

	// if type implements the Scanner interface (this triggers when initialise is called for a slice value)
	if implementsScanner(reflect.TypeOf(f.obj)) {
		err := f.addScanner(prefix, f.obj.(Scanner))
		if err != nil {
			return err
		}
	}

	// if time.Time (this triggers when initialise is called for a slice value)
	if _, ok := rv.Interface().(time.Time); ok {

		err := f.addField(prefix, rv.Addr().Interface())
		if err != nil {
			return err
		}

		return nil
	}

	// if type doesn't have nested fields (this triggers when initialise is called for a slice value)
	if rv.Kind() != reflect.Struct {

		err := f.addField(prefix, rv.Addr().Interface())
		if err != nil {
			return err
		}

		return nil
	}

	// extract expected fields
	for i := 0; i < t.NumField(); i++ {

		fieldType := t.Field(i)
		fieldValue := rv.Field(i)

		fieldName, ok := fieldType.Tag.Lookup(scanqlTag)

		if prefix != "" {
			fieldName = fmt.Sprintf("%s_%s", prefix, fieldName)
		}

		// skip if field doesn't have scanql tag
		if !ok {
			continue
		}

		fieldValueAll := instantiateAndReturnAll(fieldValue.Addr().Interface())
		fieldValueRoot := fieldValueAll[0]

		var action func() error
		scanner := asScanner(fieldValueRoot)

		switch {

		// if field implements Scanner
		case scanner != nil:
			action = func() error {
				return f.addScanner(fieldName, scanner)
			}

		// if nested struct
		case fieldValueRoot.Kind() == reflect.Struct:

			// if struct is not time
			if _, ok := fieldValueRoot.Interface().(time.Time); ok {
				action = func() error {
					return f.addField(fieldName, rv.Field(i).Addr().Interface())
				}
				break // break out of switch case
			}

			// evaluate as part of this struct (as one-to-one relationship)
			action = func() error {
				return f.addNewChild(fieldName, fieldValueAll[len(fieldValueAll)-1].Addr().Interface())
			}

		// if nested slice
		case fieldValueRoot.Kind() == reflect.Slice:
			action = func() error {
				return f.addNewChild(fieldName, fieldValueRoot.Addr().Interface())
			}

		default:
			action = func() error {
				return f.addField(fieldName, rv.Field(i).Addr().Interface())
			}
		}

		err := action()
		if err != nil {
			return err
		}
	}

	return nil
}

// instantiateAndReturnRoot will take any value and instantiate it with the equivalent Zero
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
//	instantiateAndReturnRoot(&i)
//
// as the default value would be nil, and therefore is not addressable. However if the pointer
// is initialised e.g. ip in this example:
//
//	var i int
//	ip := &i
//
// then that can be passed in directly:
//
//	instantiateAndReturnRoot(ip)
func instantiateAndReturnRoot[T any](t T) reflect.Value {
	return instantiateValue(reflect.ValueOf(t).Elem())[0]
}

func instantiateAndReturnAll[T any](t T) []reflect.Value {
	return instantiateValue(reflect.ValueOf(t).Elem())
}

func instantiateValue(val reflect.Value) []reflect.Value {

	// get value of i (must pass in as pointer), see:
	// https://stackoverflow.com/questions/34145072/can-you-initialise-a-pointer-variable-with-golang-reflect

	// if we are not at root
	if val.Kind() == reflect.Pointer {

		// instantiate current value
		val.Set(reflect.New(val.Type().Elem()))

		// crawl further
		return append(instantiateValue(val.Elem()), val)
	}

	return []reflect.Value{val}
}

// asScanner will return a Scanner established from the provided value. If a non-pointer is returned
// e.g. ExampleStruct, then the returned Scanner will be of the underlying type *ExampleStruct as a
// pointer is required for the type to modify its attributes and affect the original.
func asScanner(value reflect.Value) Scanner {

	if value.Type().Kind() != reflect.Pointer {
		value = value.Addr()
	}

	if implementsScanner(value.Type()) {
		return value.Interface().(Scanner)
	}

	return nil
}
