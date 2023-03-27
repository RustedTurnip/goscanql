package goscanql

import (
	"database/sql"
	"fmt"
)

const (
	scanqlTag = "goscanql"
)

//type aggregator[T any] struct {
//	results  []*T
//	hashRefs *hashLookup
//}
//
//type hashLookup struct {
//	// map[hash]object_reference
//	hashReferences map[string]interface{}
//
//	// map[fields_path]*hashLookup
//	hashChildren map[string]*hashLookup
//}
//
//func (h *hashLookup) addFields(f *fields) {
//
//	if _, ok := h.hashReferences[f.objHash]; !ok {
//		h.hashReferences[f.objHash] = reflect.ValueOf(f.sliceRef).Elem().Index(0)
//
//		slice := reflect.ValueOf(f.sliceRef).Elem()
//		slice.Set(reflect.Append(slice, reflect.ValueOf(f.sliceRef).Elem().Elem()))
//	}
//
//	for name, child := range f.children {
//		h.hashChildren[name] = &hashLookup{
//			hashReferences: make(map[string]interface{}),
//			hashChildren:   make(map[string]*hashLookup),
//		}
//
//		h.hashChildren[name].addFields(child)
//	}
//}

func mapFieldsToColumns[T any](cols []string, fields map[string]T) []interface{} {

	values := make([]interface{}, len(cols))

	for i, col := range cols {

		value, ok := fields[col]
		if !ok {
			continue
		}

		values[i] = value
	}

	return values
}

func scanRows[T any](rows *sql.Rows) ([]T, error) {

	result := make([]T, 0)

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	for rows.Next() {

		entry := new(T)

		fields, err := newFields(entry)
		if err != nil {
			return nil, err
		}

		err = fields.scan(cols, rows.Scan)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

// TODO this func will group the rows into the correct struct arrays and fields
func aggregateStructs[T any](rows []T) {

	// TODO pick up from here
}

// RowsToStructs will take the data in rows (*sql.Rows) as input and return a slice of
// Ts (the provided type) as the result.
func RowsToStructs[T any](rows *sql.Rows) ([]T, error) {
	return scanRows[T](rows)
}

// RowToStruct will take the data in rows (*sql.Rows) as input (similarly to RowsToStructs)
// and return a single T (the provided type) as the result and error if more or less than 1
// row is present.
func RowToStruct[T any](rows *sql.Rows) (T, error) {

	var zero T // effectively nil (as type is unknown, we can't just return nil)

	result, err := scanRows[T](rows)
	if err != nil {
		return zero, err
	}

	if len(result) != 1 {
		return zero, fmt.Errorf("rows had a non-zero length: %d", len(result))
	}

	return result[0], nil
}
