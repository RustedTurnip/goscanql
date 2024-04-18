package goscanql

import (
	"reflect"
)

// recordList is a type of map[string]record, where the key is the unique hash of the entity it
// represents (the hash is taken from the entity's initialised fields) and the value is of type
// record (used to locate a particular entry in a one-to-many relationship).
//
// In short, the recordList type is used to bind an entity hash to a record for later look-ups
// when performing entity merges.
type recordList map[string]record

// record represents a single child entity of a one-to-many relationship. This is used to look up
// a value (from the parent slice) that a fields needs to be merged with (by matching the stored
// hash of the record, with the hash of a fields).
type record struct {

	// index is the position/index in the slice that the entity the record represents is located
	// at.
	index int

	// otmChildren is the list of child one-to--many relationships that the entity this record
	// represents has. The type can be thought of as: map[fieldName]recordList.
	otmChildren map[string]recordList
}

// recordMap maintains a slice of the root type that goscanql is called with, and then the entry
// recordList for this (the recordList representing root entities rather than child one-to-manys).
// recordMap can be used to merge a fields into the existing entities.
type recordMap[T any] struct {

	// entries represents the list of currently parsed entities of type T.
	entries []T

	// hashTable represents the hashes and record information of the entities stored in the entries slice
	// of recordMap. This is used for entity matching during a merge to ensure that new data is added in
	// the right place rather than adding duplicate values.
	hashTable recordList
}

// insert will add the provided value of rv to the provided slice as a new value.
func (rl recordList) insert(entry *fields, rv *reflect.Value, slice interface{}) {
	// only perform append if the provided value isn't nil (suggesting that the insert is at
	// the point in the fields where it needs to be appended). Children after this point don't
	// need to be appended because they already exist in obj.
	if rv != nil {
		srv := reflect.ValueOf(slice).Elem()
		srv.Set(reflect.Append(srv, *rv))
	}

	r := record{
		index:       len(rl),
		otmChildren: map[string]recordList{},
	}

	for fieldName, child := range entry.oneToManys {
		rlChild := recordList{}
		rlChild.insert(child, nil, nil)
		r.otmChildren[fieldName] = rlChild
	}

	rl[entry.getHash()] = r
}

// merge will recursively search the provided fields against the stored records to determine
// how the value represented by fields should be combined into the existing entries. Where a
// one-to-many relationship is found where no child matches the hash of the fields, this will
// be added as a new value in the one-to-many slice.
func (rl recordList) merge(entry *fields, rv *reflect.Value, slice interface{}) {
	if entry.isNil() {
		return
	}

	f, ok := rl[entry.getHash()]
	if !ok {
		rl.insert(entry, rv, slice)
		return
	}

	match := getRootValue(reflect.ValueOf(slice).Elem().Index(f.index))

	for fieldName, child := range entry.oneToManys {
		childSlice := getRootValue(*fieldByTag(fieldName, match))
		rvChild := reflect.ValueOf(child.obj).Elem()

		f.otmChildren[fieldName].merge(child, &rvChild, childSlice.Addr().Interface())
	}
}

// merge will apply the provided fields to the existing entities maintained by recordMap, using
// fields hash values to determine where the data already exists, or where it should be added
// as new.
func (rm *recordMap[T]) merge(entry *fields) {
	rv := reflect.ValueOf(entry.obj).Elem()
	rm.hashTable.merge(entry, &rv, &rm.entries)
}

// newRecordMap is the constructor for record map, and will return an instantiated recordMap
// based on the provided type T.
func newRecordMap[T any]() *recordMap[T] {
	return &recordMap[T]{
		entries:   make([]T, 0),
		hashTable: recordList{},
	}
}

// fieldByTag will look up a field of the provided value (v) by the field's tag value (where
// the field is tagged with goscanql). If no field matches the provided tag, then nil is
// returned.
func fieldByTag(tag string, v reflect.Value) *reflect.Value {
	tv := v.Type()

	for i := 0; i < v.NumField(); i++ {
		if tv.Field(i).Tag.Get(scanqlTag) != tag {
			continue
		}

		f := v.Field(i)
		return &f
	}

	return nil
}

// getRootValue will traverse the provided reflect.Value (v) until a non-pointer type
// is reached and return that.
//
// Note: if the provided value isn't fully instantiated, i.e. a pointer to a nil value, then
// this will cause problems when trying to call functions like .Type() on the returned value.
func getRootValue(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Pointer {
		v = v.Elem()
	}

	return v
}
