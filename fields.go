package goscanql

import (
	"crypto/sha1"
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"time"
)

// fieldsContainer maintains a record of a slice of entities along with the
// fields representations of those entities to facilitate fields merging.
type fieldsContainer struct {

	// containerRef is the reference of the slice containing the entities as their
	// own type.
	containerRef interface{}

	// fields represents the entities contained in containerRef but as fields (to
	// facilitate hash lookups).
	fields []*fields
}

func (fc *fieldsContainer) isSlice() bool {
	return reflect.TypeOf(fc.containerRef).Elem().Kind() == reflect.Slice
}

// append will add the provided fields (and the entity it represents) into
// the fieldsContainer.
func (fc *fieldsContainer) append(f *fields) {

	if !fc.isSlice() {
		panic(fmt.Errorf("cannot append to fieldsContainer that doesn't hold slice"))
	}

	parent := reflect.ValueOf(fc.containerRef).Elem()
	fParent := reflect.ValueOf(f.container.containerRef).Elem()

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

func (fc *fieldsContainer) empty() {

	if fc.containerRef == nil {
		return
	}

	t := reflect.TypeOf(fc.containerRef).Elem()
	rv := reflect.ValueOf(fc.containerRef).Elem()

	switch t.Kind() {

	// if slice, replace slice with an empty one
	case reflect.Slice:
		rv.Set(reflect.New(t).Elem())
		return

	default:
		rv.Set(reflect.New(t).Elem())
	}
}

// newFieldsContainer is a fieldsContainer constructor
func newFieldsContainer(container interface{}, f *fields) *fieldsContainer {
	return &fieldsContainer{
		containerRef: container,
		fields: []*fields{
			f,
		},
	}
}

type fields struct {
	container            *fieldsContainer
	orderedFieldNames    []string
	orderedOneToOneNames []string
	references           map[string]interface{}
	byteReferences       map[string]*[]byte
	oneToOnes            map[string]*fields
	oneToManys           map[string]*fields
}

func (f *fields) addNewOneToMany(name string, obj interface{}) error {
	return addNewChild(name, obj, f.oneToManys)
}

func (f *fields) addNewOneToOne(name string, obj interface{}) error {
	return addNewChild(name, obj, f.oneToOnes)
}

func addNewChild(name string, obj interface{}, m map[string]*fields) error {

	if _, ok := m[name]; ok {
		panic(fmt.Errorf("child with same name (\"%s\") already exists", name))
	}

	child, err := newFields(obj)
	if err != nil {
		return err
	}

	m[name] = child
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
	f.byteReferences[name] = &[]byte{}
}

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

func (f *fields) crawlFields(fn func(string, *fields) bool) {
	f.crawlFieldsWithPrefix("", fn)
}

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

func (f *fields) isNil() bool {

	for _, b := range f.byteReferences {
		if len(*b) > 0 {
			return false
		}
	}

	return true
}

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

func (f *fields) emptyNilFieldsFromSlice() {

	if f.isNil() {
		f.container.empty()
		return
	}

	for _, child := range f.oneToOnes {
		child.emptyNilFieldsFromSlice()
	}

	for _, child := range f.oneToManys {
		child.emptyNilFieldsFromSlice()
	}
}

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

	f.emptyNilFieldsFromSlice()

	return nil
}

func newFields(obj interface{}) (*fields, error) {

	fields := &fields{
		orderedFieldNames: make([]string, 0),
		references:        make(map[string]interface{}),
		byteReferences:    make(map[string]*[]byte),
		oneToOnes:         make(map[string]*fields),
		oneToManys:        make(map[string]*fields),
	}

	fields.container = newFieldsContainer(obj, fields)

	err := initialiseFields("", obj, fields)
	if err != nil {
		return nil, err
	}

	return fields, nil
}

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

		fields.container.containerRef = rv.Addr().Interface()

		// get slice type, e.g. []*Example has a slice type of *Example
		sliceType := reflect.TypeOf(rv.Interface()).Elem()

		// create new element of sliceType
		element := reflect.New(sliceType).Elem()

		// instantiate element's root value
		instantiateAndReturnRoot(element.Addr().Interface())

		// append new element to slice
		rv.Set(reflect.Append(rv, element))

		// add child and evaluate it
		return initialiseFields("", rv.Index(0).Addr().Interface(), fields)
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
				fields.addNewOneToOne(fieldName, fieldValueAll[len(fieldValueAll)-1].Addr().Interface())
				continue
			}
		}

		// if nested slice
		if fieldValueRoot.Kind() == reflect.Slice {

			// evaluate with pointer to new instance (as child because one-to-many relationship)
			err := fields.addNewOneToMany(fieldName, fieldValueRoot.Addr().Interface())
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
	if !f.container.isSlice() {

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
		existing = f.container.getExisting(m)

		// if *fields doesn't already exist, add it as new
		if existing == nil {
			f.container.append(m)
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
