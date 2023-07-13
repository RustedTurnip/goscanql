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

	// GetID ia used to identify a value's uniqueness compared to other values of
	// the same type during a scan. This can be returned as nil, but should
	// otherwise consistently return a value that uniquely represents the types
	// value.
	GetID() []byte
}

// implementsScanner evaluates the provided type and returns true if it implements
// the Scanner interface, or false otherwise.
func implementsScanner(t reflect.Type) bool {
	return t.Implements(reflect.TypeOf((*Scanner)(nil)).Elem())
}

// NullString represents a string that can be null. If null, then the attribute
// Valid will be set to false, otherwise the value stored in String represents the
// string value. This type implements the goscanql Scanner interface and can be
// used when scanning potentially null strings in from a database.
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

// NullInt64 represents a string that can be null. If null, then the attribute
// Valid will be set to false, otherwise the value stored in Int64 represents the
// int64 value. This type implements the goscanql Scanner interface and can be
// used when scanning potentially null int64s in from a database.
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

// NullInt32 represents a string that can be null. If null, then the attribute
// Valid will be set to false, otherwise the value stored in Int32 represents the
// int32 value. This type implements the goscanql Scanner interface and can be
// used when scanning potentially null int32s in from a database.
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

// NullInt16 represents a string that can be null. If null, then the attribute
// Valid will be set to false, otherwise the value stored in Int16 represents the
// int16 value. This type implements the goscanql Scanner interface and can be
// used when scanning potentially null int16s in from a database.
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

// NullByte represents a string that can be null. If null, then the attribute
// Valid will be set to false, otherwise the value stored in Byte represents the
// byte value. This type implements the goscanql Scanner interface and can be
// used when scanning potentially null bytes in from a database.
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

// NullFloat64 represents a string that can be null. If null, then the attribute
// Valid will be set to false, otherwise the value stored in Float64 represents the
// float64 value. This type implements the goscanql Scanner interface and can be
// used when scanning potentially null float64s in from a database.
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

// NullBool represents a string that can be null. If null, then the attribute
// Valid will be set to false, otherwise the value stored in Bool represents the
// bool value. This type implements the goscanql Scanner interface and can be
// used when scanning potentially null bools in from a database.
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

// NullTime represents a string that can be null. If null, then the attribute
// Valid will be set to false, otherwise the value stored in Time represents the
// time value. This type implements the goscanql Scanner interface and can be
// used when scanning potentially null time in from a database.
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
