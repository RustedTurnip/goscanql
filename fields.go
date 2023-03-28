package goscanql

import (
	"crypto/sha1"
	"fmt"
	"reflect"
)

// fieldsContainer maintains a record of a slice of entities along with the
// fields representations of those entities to facilitate fields merging.
type fieldsContainer struct {

	// sliceRef is the reference of the slice containing the entities as their
	// own type.
	sliceRef interface{}

	// fields represents the entities contained in sliceRef but as fields (to
	// facilitate hash lookups).
	fields []*fields
}

// append will add the provided fields (and the entity it represents) into
// the fieldsContainer.
func (fc *fieldsContainer) append(f *fields) {

	parent := reflect.ValueOf(fc.sliceRef).Elem()
	fParent := reflect.ValueOf(f.sliceRef.sliceRef).Elem()

	parent.Set(reflect.Append(parent, fParent.Index(0)))
	fc.fields = append(fc.fields, f)
}

// getExisting returns the existing *fields entity that is contained within
// the fieldsContainer by looking up the provided fields hash.
//
// nil is returned when the fields entity doesn't already exist.
func (fc *fieldsContainer) getExisting(f *fields) *fields {

	fHash := f.getHash()

	for _, existing := range fc.fields {
		if existing.getHash() == fHash {
			return existing
		}
	}

	return nil
}

// newFieldsContainer is a fieldsContainer constructor
func newFieldsContainer(slice interface{}, f *fields) *fieldsContainer {
	return &fieldsContainer{
		sliceRef: slice,
		fields: []*fields{
			f,
		},
	}
}

type fields struct {
	sliceRef          *fieldsContainer
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

func (f *fields) scan(columns []string, scan func(...interface{}) error) error {

	refs := mapFieldsToColumns(columns, f.getFieldReferences())

	err := scan(refs...)
	if err != nil {
		return err
	}

	return nil
}

func newFields(obj interface{}) (*fields, error) {

	fields := &fields{
		orderedFieldNames: make([]string, 0),
		references:        make(map[string]interface{}),
		children:          make(map[string]*fields),
	}

	fields.sliceRef = newFieldsContainer(obj, fields)

	err := initialiseFields("", obj, fields)
	if err != nil {
		return nil, err
	}

	return fields, nil
}

func initialiseFields(prefix string, obj interface{}, fields *fields) error {

	rv := instantiateAndReturn(obj)
	t := rv.Type()

	// if type is slice, add 1 element to it to store values
	if rv.Kind() == reflect.Slice {

		fields.sliceRef.sliceRef = rv.Addr().Interface()

		// get slice type, e.g. []*Example has a slice type of *Example
		sliceType := reflect.TypeOf(rv.Interface()).Elem()

		// create new element of sliceType
		element := reflect.New(sliceType).Elem()

		// instantiate element's root value
		instantiateAndReturn(element.Addr().Interface())

		// append new element to slice
		rv.Set(reflect.Append(rv, element))

		// add child and evaluate it
		return initialiseFields("", rv.Index(0).Addr().Interface(), fields)
	}

	// if type is primitive (not slice or struct) use arbitrary field name of the attributes
	// type and return e.g. for arry of strings (rather than array of structs):
	//
	//  type User struct {
	//      Aliases []string `goscanql:"aliases"`
	//  }
	//
	// would end up as:
	//
	//  "aliases_string"
	//
	// in the field references so that it is accessible.
	if rv.Kind() != reflect.Struct {
		fields.addField(rv.Type().String(), rv.Addr().Interface())
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

		fieldValueRoot := instantiateAndReturn(fieldValue.Addr().Interface())

		// if nested struct
		if fieldValueRoot.Kind() == reflect.Struct {

			// evaluate as part of this struct (as one-to-one relationship)
			initialiseFields(fieldName, fieldValueRoot.Addr().Interface(), fields)
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
// If the parent element is the same, the merge will apply to children of the parent
// and leave the parent untouched.
//
// If the parent and provided fields are different and do not belong to a slice in which
// they can coexist, an error will be returned.
func (f *fields) merge(m *fields) error {

	var existing *fields

	// if element doesn't belong to a slice
	if f.sliceRef.sliceRef == nil {

		// and the provided element doesn't match the current element
		// then fail merge as they are different so children cannot be merged
		if f.getHash() != m.getHash() {
			return fmt.Errorf("cannot merge fields as their data differs and they do not belong to a slice.")
		}

		// if the provided fields matches the current fields, set existing to be
		// current
		existing = f

	} else {

		// else, if sliceRef isn't nil, set existing to be any existing entity with
		// same hash
		existing = f.sliceRef.getExisting(m)

		// if *fields doesn't already exist, add it as new
		if existing == nil {
			f.sliceRef.append(m)
			return nil
		}
	}

	// for each of existing fields children, merge with incoming fields
	for name, child := range existing.children {

		mChild, ok := m.children[name]
		if !ok {
			return fmt.Errorf("provided fields is missing expected child \"%s\"", name)
		}

		if err := child.merge(mChild); err != nil {
			return err
		}
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
