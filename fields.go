package goscanql

import (
	"crypto/sha1"
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"time"
)

// fieldsSlice maintains a record of a slice of entities along with the
// fields representations of those entities to facilitate fields merging.
type fieldsSlice struct {

	// sliceRef is the reference of the slice containing the entities as their
	// own type.
	sliceRef interface{}

	// fields represents the entities contained in sliceRef but as fields (to
	// facilitate hash lookups).
	fields []*fields
}

// append will add the provided fields (and the entity it represents) into
// the fieldsSlice.
func (fs *fieldsSlice) append(f *fields) {

	parent := reflect.ValueOf(fs.sliceRef).Elem()
	fParent := reflect.ValueOf(f.slice.sliceRef).Elem()

	parent.Set(reflect.Append(parent, fParent.Index(0)))
	fs.fields = append(fs.fields, f)
}

// getExisting returns the existing *fields entity that is contained within
// the fieldsSlice by looking up the provided fields hash.
//
// nil is returned when the fields entity doesn't already exist.
func (fs *fieldsSlice) getExisting(f *fields) *fields {

	fHash := f.getHash()

	for _, existing := range fs.fields {
		if existing.getHash() == fHash {
			return existing
		}
	}

	return nil
}

// empty will set the slice to be an empty slice (removing any contained elements)
func (fc *fieldsSlice) empty() {

	if fc.sliceRef == nil {
		return
	}

	t := reflect.TypeOf(fc.sliceRef).Elem()
	rv := reflect.ValueOf(fc.sliceRef).Elem()

	rv.Set(reflect.MakeSlice(t, 0, 0))
}

// newFieldsSlice is a fieldsSlice constructor
func newFieldsSlice(container interface{}, f *fields) *fieldsSlice {
	return &fieldsSlice{
		sliceRef: container,
		fields: []*fields{
			f,
		},
	}
}

// fields holds a goscanql parsed struct, maintaining references to the fields
// of the struct and any sub-structs (children).
type fields struct {

	// obj is a reference (pointer) to the struct that this fields fields belong to.
	obj interface{}

	// slice represents the slice that the struct is a part of. If the struct isn't a part
	// of a slice, this will be nil.
	slice *fieldsSlice

	// orderedFieldNames maintains the field names in the order of which they were added
	// to facilitate reliable hashing when comparing fields entities.
	orderedFieldNames []string

	// orderedOneToOneNames maintains the names of the one-to-one relationship children so
	// that they can reliably be hashed for comparison.
	orderedOneToOneNames []string

	// references holds a reference to each field belonging to a fields entity so they can
	// be set.
	references map[string]interface{}

	// byteReferences maintains a reference of a byte slice for each field which is used for
	// determining nil fields.
	byteReferences map[string]*[]byte

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

	for childName, _ := range f.oneToOnes {
		if childName == name {
			return collisionErr
		}
	}

	for childName, _ := range f.oneToManys {
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
	return nil
}

// addField will add a single field to the current fields (e.g. a string or int).
func (f *fields) addField(name string, value interface{}) {

	// assert that field hasn't already been added
	if _, ok := f.references[name]; ok {
		panic(fmt.Errorf("field with name \"%s\" already added", name))
	}

	// add field to this instance
	f.orderedFieldNames = append(f.orderedFieldNames, name)
	f.references[name] = value
	f.byteReferences[name] = &[]byte{}
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

		return false
	})

	return m
}

// getByteReferences returns a map of all of the fields byte references (including any child
// field references).
func (f *fields) getByteReferences() map[string]*[]byte {

	m := make(map[string]*[]byte)

	f.crawlFields(func(prefix string, fi *fields) bool {

		for name, reference := range fi.byteReferences {
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

	raw := make([]byte, 0)

	for _, key := range f.orderedFieldNames {

		value := f.references[key]

		strValue := fmt.Sprintf("%s%v", key, reflect.ValueOf(value).Elem().Interface())

		raw = append(raw, []byte(strValue)...)
	}

	// hash fields to create unique id for struct
	h := sha1.New()
	h.Write(raw)

	return string(h.Sum(nil))
}

// isNil will the incoming data to a fields (once it has been written to the byteReferences)
// to see if the object that the fields represents will be nil.
func (f *fields) isNil() bool {

	for _, b := range f.byteReferences {
		if len(*b) > 0 {
			return false
		}
	}

	return true
}

// isMatch will compare the provided fields (m) to the current fields to see if they are equal
// in value, returning true if they are, and false otherwise.
func (f *fields) isMatch(m *fields) bool {

	if f.getHash() != m.getHash() {
		return false
	}

	for name, child := range f.oneToOnes {

		mChild, ok := m.oneToOnes[name]
		if !ok {
			return false
		}

		if !child.isMatch(mChild) {
			return false
		}
	}

	return true
}

// emptyNilFields will nullify where possible any nil objects that are represented by the fields.
func (f *fields) emptyNilFields() {

	// if the fields values are nil
	if f.isNil() {

		// if the object belongs to a slice, empty that
		if f.slice != nil {
			f.slice.empty()
		}

		// empty the object represented by the fields, e.g. *int would be set to nil,
		// or int would be set to 0.
		rv := reflect.ValueOf(f.obj).Elem()
		rv.Set(reflect.New(rv.Type()).Elem())

		return
	}

	// repeat for all children
	for _, child := range f.oneToOnes {
		child.emptyNilFields()
	}

	for _, child := range f.oneToManys {
		child.emptyNilFields()
	}
}

// scan will attempt to apply the provided scan function to the fields object
// by providing it with all the field references so that values can be written.
func (f *fields) scan(columns []string, scan func(...interface{}) error) error {

	byteRefs := mapFieldsToColumns(columns, f.getByteReferences())

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
		obj:               obj,
		orderedFieldNames: make([]string, 0),
		references:        make(map[string]interface{}),
		byteReferences:    make(map[string]*[]byte),
		oneToOnes:         make(map[string]*fields),
		oneToManys:        make(map[string]*fields),
	}

	// if slice, set the fields slice to be the slice so we can append to it during
	// fields.merge
	if rv.Kind() == reflect.Slice {
		fields.slice = newFieldsSlice(rv.Addr().Interface(), fields)
	}

	// initialise the newly created fields around the obj being pointed to
	err := initialiseFields("", obj, fields)
	if err != nil {
		return nil, err
	}

	return fields, nil
}

// initialiseFields uses reflection to map it out and maintain references to the object's
// fields.
func initialiseFields(prefix string, obj interface{}, fields *fields) error {

	rva := instantiateAndReturnAll(obj)

	rv := rva[0]
	t := rv.Type()

	// if type implements the Scanner interface, add it as is
	iScanner := reflect.TypeOf((*sql.Scanner)(nil)).Elem()
	if rv.Type().Implements(iScanner) {
		fields.addField(prefix, rv.Addr().Interface())
		return nil
	}

	// if type is slice, add 1 element to it to store values
	if rv.Kind() == reflect.Slice {
		panic("multi-dimensional slices are not supported")
	}

	// if time.Time
	if _, ok := rv.Interface().(time.Time); ok {
		fields.addField(prefix, rv.Addr().Interface())
		return nil
	}

	// if primitive
	if rv.Kind() != reflect.Struct {
		fields.addField(prefix, rv.Addr().Interface())
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

		// if nested struct
		if fieldValueRoot.Kind() == reflect.Struct {

			// and if struct is not time
			if _, ok := fieldValueRoot.Interface().(time.Time); !ok {

				// evaluate as part of this struct (as one-to-one relationship)
				fields.addNewChild(fieldName, fieldValueAll[len(fieldValueAll)-1].Addr().Interface())
				continue
			}
		}

		// if nested slice
		if fieldValueRoot.Kind() == reflect.Slice {

			// evaluate with pointer to new instance (as child because one-to-many relationship)
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

// merge will attempt to merge the provided fields (m) to the fields being called upon.
// This merge will result in any differing elements being added along side the current
// element if they are different but belong to a slice.
//
// If the parent element is the same, the merge will apply to oneToManys of the parent
// and leave the parent untouched.
//
// If the parent and provided fields are different and do not belong to a slice in which
// they can coexist, an error will be returned.
func (f *fields) merge(m *fields) error {

	// if nothing to merge, return
	if m.isNil() {
		return nil
	}

	var existing *fields

	// if element doesn't belong to a slice
	if f.slice == nil {

		// and the provided element doesn't match the current element
		// then fail merge as they are different so oneToManys cannot be merged
		if !f.isMatch(m) {
			return fmt.Errorf("cannot merge fields as their data differs and they do not belong to a slice.")
		}

		// if the provided fields matches the current fields, set existing to be
		// current
		existing = f

	} else {

		// else, if container isn't nil, set existing to be any existing entity with
		// same hash
		existing = f.slice.getExisting(m)

		// if *fields doesn't already exist, add it as new
		if existing == nil {
			f.slice.append(m)
			return nil
		}
	}

	// for each of existing fields oneToManys, merge with incoming fields
	for name, child := range existing.oneToManys {

		mChild, ok := m.oneToManys[name]
		if !ok {
			return fmt.Errorf("provided fields is missing expected child \"%s\"", name)
		}

		if mChild.isNil() {
			continue
		}

		if err := child.merge(mChild); err != nil {
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
