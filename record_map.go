package goscanql

import (
	"reflect"
)

// recordList TODO map[fieldHash]record
type recordList map[string]record

type record struct {
	index int

	// map[fieldName]recordList
	otmChildren map[string]recordList
}

type recordMap[T any] struct {
	// entries represents the
	entries []T

	// hashTable (TODO rename) holds the unique value hash for each entry, and is linked to
	// the entry through the record.index.. TODO
	hashTable recordList
}

// insert TODO
func (rl recordList) insert(entry *fields, rv *reflect.Value, slice interface{}) {

	// only perform append if the provided value isn't nil (suggesting that the insert is at
	// the point in the fields where it needs to be appended) - children after this point don't
	// need to be appended because they already exist in obj
	if rv != nil { // TODO maybe need to check if obj is nil?

		srv := reflect.ValueOf(slice).Elem()
		srv.Set(reflect.Append(srv, *rv))
	}

	r := record{
		index:       len(rl), // TODO check this way of setting index will work
		otmChildren: map[string]recordList{},
	}

	for fieldName, child := range entry.oneToManys {
		rlChild := recordList{}
		rlChild.insert(child, nil, nil) // TODO
		r.otmChildren[fieldName] = rlChild
	}

	rl[entry.getHash()] = r
}

// merge TODO
func (rl recordList) merge(entry *fields, rv *reflect.Value, slice interface{}) {

	// TODO should this be in here, it feels like isNil should be more private to fields
	if entry.isNil() {
		return
	}

	f, ok := rl[entry.getHash()]
	if !ok {
		rl.insert(entry, rv, slice) // TODO
		return
	}

	match := getRootValue(reflect.ValueOf(slice).Elem().Index(f.index))

	for fieldName, child := range entry.oneToManys {

		childSlice := getRootValue(*fieldByTag(fieldName, match))
		rvChild := reflect.ValueOf(child.obj).Elem()

		f.otmChildren[fieldName].merge(child, &rvChild, childSlice.Addr().Interface())
	}
}

// merge TODO
func (rm *recordMap[T]) merge(entry *fields) {
	rv := reflect.ValueOf(entry.obj).Elem()
	rm.hashTable.merge(entry, &rv, &rm.entries)
}

// newRecordMap TODO
func newRecordMap[T any]() *recordMap[T] {
	return &recordMap[T]{
		entries:   make([]T, 0),
		hashTable: recordList{},
	}
}

func fieldByTag(tag string, v reflect.Value) *reflect.Value {

	tv := v.Type()

	for i := 0; i < v.NumField(); i++ {

		if tv.Field(i).Tag.Get("goscanql") != tag {
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
