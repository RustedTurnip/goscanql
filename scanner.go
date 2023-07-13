package goscanql

import (
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"time"
)

// Scanner represents an interface to provide custom scanning logic to a field for
// goscanql. This can be used to parse data from an sql column in a non-default
// way, for example parsing a string into a struct instead of a string, or to
// provide a way to scan data into a type that is otherwise unsupported like an
// array or a multi-dimensional slice.
type Scanner interface {
	sql.Scanner
	GetID() []byte
}

// implementsScanner evaluates the provided type and returns true if it implements
// the Scanner interface, or false otherwise.
func implementsScanner(t reflect.Type) bool {
	return t.Implements(reflect.TypeOf((*Scanner)(nil)).Elem())
}

// TODO comment
type NullString struct {
	String string
	Valid  bool
}

func (ns *NullString) Scan(value interface{}) error {
	if value == nil {
		ns.String, ns.Valid = "", false
		return nil
	}

	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("NullString received non-string type (%s) during Scan", reflect.TypeOf(value).String())
	}

	ns.String, ns.Valid = str, true
	return nil
}

func (ns *NullString) GetID() []byte {

	if !ns.Valid {
		return nil
	}

	return []byte(ns.String)
}

// TODO comment
type NullInt64 struct {
	Int64 int64
	Valid bool
}

func (ni *NullInt64) Scan(value interface{}) error {
	if value == nil {
		ni.Int64, ni.Valid = 0, false
		return nil
	}

	i, ok := value.(int64)
	if !ok {
		return fmt.Errorf("NullInt64 received non-int64 type (%s) during Scan", reflect.TypeOf(value).String())
	}

	ni.Int64, ni.Valid = i, true
	return nil
}

func (ni *NullInt64) GetID() []byte {

	if !ni.Valid {
		return nil
	}

	return []byte(strconv.FormatInt(ni.Int64, 10))
}

// TODO comment
type NullInt32 struct {
	Int32 int32
	Valid bool
}

func (ni *NullInt32) Scan(value interface{}) error {
	if value == nil {
		ni.Int32, ni.Valid = 0, false
		return nil
	}

	i, ok := value.(int32)
	if !ok {
		return fmt.Errorf("NullInt32 received non-int32 type (%s) during Scan", reflect.TypeOf(value).String())
	}

	ni.Int32, ni.Valid = i, true
	return nil
}

func (ni *NullInt32) GetID() []byte {

	if !ni.Valid {
		return nil
	}

	return []byte(strconv.FormatInt(int64(ni.Int32), 10))
}

// TODO comment
type NullInt16 struct {
	Int16 int16
	Valid bool
}

func (ni *NullInt16) Scan(value interface{}) error {
	if value == nil {
		ni.Int16, ni.Valid = 0, false
		return nil
	}

	i, ok := value.(int16)
	if !ok {
		return fmt.Errorf("NullInt16 received non-int16 type (%s) during Scan", reflect.TypeOf(value).String())
	}

	ni.Int16, ni.Valid = i, true
	return nil
}

func (ni *NullInt16) GetID() []byte {

	if !ni.Valid {
		return nil
	}

	return []byte(strconv.FormatInt(int64(ni.Int16), 10))
}

// TODO comment
type NullByte struct {
	Byte  byte
	Valid bool
}

func (ni *NullByte) Scan(value interface{}) error {
	if value == nil {
		ni.Byte, ni.Valid = 0, false
		return nil
	}

	i, ok := value.(byte)
	if !ok {
		return fmt.Errorf("NullByte received non-byte type (%s) during Scan", reflect.TypeOf(value).String())
	}

	ni.Byte, ni.Valid = i, true
	return nil
}

func (ni *NullByte) GetID() []byte {

	if !ni.Valid {
		return nil
	}

	return []byte{ni.Byte}
}

// TODO comment
type NullFloat64 struct {
	Float64 float64
	Valid   bool
}

func (ni *NullFloat64) Scan(value interface{}) error {
	if value == nil {
		ni.Float64, ni.Valid = 0, false
		return nil
	}

	i, ok := value.(float64)
	if !ok {
		return fmt.Errorf("NullFloat64 received non-float64 type (%s) during Scan", reflect.TypeOf(value).String())
	}

	ni.Float64, ni.Valid = i, true
	return nil
}

func (ni *NullFloat64) GetID() []byte {

	if !ni.Valid {
		return nil
	}

	return []byte(strconv.FormatFloat(ni.Float64, 'f', -1, 64))
}

// TODO comment
type NullBool struct {
	Bool  bool
	Valid bool
}

func (ni *NullBool) Scan(value interface{}) error {
	if value == nil {
		ni.Bool, ni.Valid = false, false
		return nil
	}

	i, ok := value.(bool)
	if !ok {
		return fmt.Errorf("NullBool received non-bool type (%s) during Scan", reflect.TypeOf(value).String())
	}

	ni.Bool, ni.Valid = i, true
	return nil
}

func (ni *NullBool) GetID() []byte {

	if !ni.Valid {
		return nil
	}

	return []byte(strconv.FormatBool(ni.Bool))
}

// TODO comment
type NullTime struct {
	Time  time.Time
	Valid bool
}

func (ni *NullTime) Scan(value interface{}) error {
	if value == nil {
		ni.Time, ni.Valid = time.Time{}, false
		return nil
	}

	i, ok := value.(time.Time)
	if !ok {
		return fmt.Errorf("NullTime received non-time.Time type (%s) during Scan", reflect.TypeOf(value).String())
	}

	ni.Time, ni.Valid = i, true
	return nil
}

func (ni *NullTime) GetID() []byte {

	if !ni.Valid {
		return nil
	}

	return []byte(ni.Time.Format(time.RFC3339Nano))
}
