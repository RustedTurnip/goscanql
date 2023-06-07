package goscanql

import (
	"fmt"
	"reflect"
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
	fmt.Println("APPENDING")
	fmt.Println(reflect.ValueOf(fs.sliceRef).Elem().Pointer())
	fParent := reflect.ValueOf(f.slice.sliceRef).Elem()

	parent.Set(reflect.Append(parent, fParent.Index(0)))
	fs.fields = append(fs.fields, f)
	fmt.Println("POST APPENDING")
	fmt.Println(reflect.ValueOf(fs.sliceRef).Elem().Pointer())
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
func (fs *fieldsSlice) empty() {

	if fs.sliceRef == nil {
		return
	}

	t := reflect.TypeOf(fs.sliceRef).Elem()
	rv := reflect.ValueOf(fs.sliceRef).Elem()

	rv.Set(reflect.MakeSlice(t, 0, 0))
}

// newFieldsSlice is a fieldsSlice constructor
func newFieldsSlice(container interface{}, f *fields) *fieldsSlice {
	fmt.Println(fmt.Sprintf("NEWFIELDSSLICE %s", reflect.TypeOf(container).String()))
	return &fieldsSlice{
		sliceRef: container,
		fields: []*fields{
			f,
		},
	}
}
