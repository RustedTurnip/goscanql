package goscanql

import "reflect"

type Scanner interface {
	Scan(interface{}) error
	GetID() string
}

// TODO comment this
func implementsScanner(t reflect.Type) bool {
	return t.Implements(reflect.TypeOf((*Scanner)(nil)).Elem())
}
