package goscanql

import (
	"database/sql"
	"reflect"
)

// Scanner represents an interface to provide custom scanning logic to a field for
// goscanql. This can be used to parse data from an sql column in a non-default
// way, for example parsing a string into a struct instead of a string, or to
// provide a way to scan data into a type that is otherwise unsupported like an
// array or a multi-dimensional slice.
type Scanner interface {
	sql.Scanner
	GetID() string
}

// implementsScanner evaluates the provided type and returns true if it implements
// the Scanner interface, or false otherwise.
func implementsScanner(t reflect.Type) bool {
	return t.Implements(reflect.TypeOf((*Scanner)(nil)).Elem())
}
