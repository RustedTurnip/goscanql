package goscanql

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNullString_Scan(t *testing.T) {

	tests := []struct {
		name            string
		scanInput       interface{}
		nullStringInput *NullString
		expected        *NullString
		expectedErr     error
	}{
		{
			name:            "Valid String Empty NullString",
			scanInput:       "valid_string",
			nullStringInput: &NullString{},
			expected: &NullString{
				String: "valid_string",
				Valid:  true,
			},
			expectedErr: nil,
		},
		{
			name:      "Valid String Non-Empty NullString",
			scanInput: "valid_string",
			nullStringInput: &NullString{
				String: "existing_string",
				Valid:  false,
			},
			expected: &NullString{
				String: "valid_string",
				Valid:  true,
			},
			expectedErr: nil,
		},
		{
			name:            "Invalid Input Empty NullString",
			scanInput:       0,
			nullStringInput: &NullString{},
			expected: &NullString{
				String: "",
				Valid:  false,
			},
			expectedErr: fmt.Errorf("NullString received non-string type (int) during Scan"),
		},
		{
			name:      "Invalid Input Non-Empty NullString",
			scanInput: 0,
			nullStringInput: &NullString{
				String: "existing_string",
				Valid:  true,
			},
			expected: &NullString{
				String: "existing_string",
				Valid:  true,
			},
			expectedErr: fmt.Errorf("NullString received non-string type (int) during Scan"),
		},
		{
			name:            "Nil Input Empty NullString",
			scanInput:       nil,
			nullStringInput: &NullString{},
			expected: &NullString{
				String: "",
				Valid:  false,
			},
			expectedErr: nil,
		},
		{
			name:      "Nil Input Non-Empty NullString",
			scanInput: nil,
			nullStringInput: &NullString{
				String: "existing_string",
				Valid:  true,
			},
			expected: &NullString{
				String: "",
				Valid:  false,
			},
			expectedErr: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			// Act
			err := test.nullStringInput.Scan(test.scanInput)

			// Assert
			assert.Equal(t, test.expected, test.nullStringInput)
			assert.Equal(t, test.expectedErr, err)
		})
	}
}

func TestNullInt64_Scan(t *testing.T) {

	tests := []struct {
		name           string
		scanInput      interface{}
		nullInt64Input *NullInt64
		expected       *NullInt64
		expectedErr    error
	}{
		{
			name:           "Valid In64 Empty NullInt64",
			scanInput:      int64(64),
			nullInt64Input: &NullInt64{},
			expected: &NullInt64{
				Int64: 64,
				Valid: true,
			},
			expectedErr: nil,
		},
		{
			name:      "Valid Int64 Non-Empty NullInt64",
			scanInput: int64(64),
			nullInt64Input: &NullInt64{
				Int64: 32,
				Valid: false,
			},
			expected: &NullInt64{
				Int64: 64,
				Valid: true,
			},
			expectedErr: nil,
		},
		{
			name:           "Invalid Input Empty NullInt64",
			scanInput:      "non_int64",
			nullInt64Input: &NullInt64{},
			expected: &NullInt64{
				Int64: 0,
				Valid: false,
			},
			expectedErr: fmt.Errorf("NullInt64 received non-int64 type (string) during Scan"),
		},
		{
			name:      "Invalid Input Non-Empty NullInt64",
			scanInput: "non_int64",
			nullInt64Input: &NullInt64{
				Int64: 64,
				Valid: true,
			},
			expected: &NullInt64{
				Int64: 64,
				Valid: true,
			},
			expectedErr: fmt.Errorf("NullInt64 received non-int64 type (string) during Scan"),
		},
		{
			name:           "Nil Input Empty NullInt64",
			scanInput:      nil,
			nullInt64Input: &NullInt64{},
			expected: &NullInt64{
				Int64: 0,
				Valid: false,
			},
			expectedErr: nil,
		},
		{
			name:      "Nil Input Non-Empty NullInt64",
			scanInput: nil,
			nullInt64Input: &NullInt64{
				Int64: 32,
				Valid: true,
			},
			expected: &NullInt64{
				Int64: 0,
				Valid: false,
			},
			expectedErr: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			// Act
			err := test.nullInt64Input.Scan(test.scanInput)

			// Assert
			assert.Equal(t, test.expected, test.nullInt64Input)
			assert.Equal(t, test.expectedErr, err)
		})
	}
}

func TestNullInt32_Scan(t *testing.T) {

	tests := []struct {
		name           string
		scanInput      interface{}
		nullInt32Input *NullInt32
		expected       *NullInt32
		expectedErr    error
	}{
		{
			name:           "Valid In32 Empty NullInt32",
			scanInput:      int32(64),
			nullInt32Input: &NullInt32{},
			expected: &NullInt32{
				Int32: 64,
				Valid: true,
			},
			expectedErr: nil,
		},
		{
			name:      "Valid Int32 Non-Empty NullInt32",
			scanInput: int32(64),
			nullInt32Input: &NullInt32{
				Int32: 32,
				Valid: false,
			},
			expected: &NullInt32{
				Int32: 64,
				Valid: true,
			},
			expectedErr: nil,
		},
		{
			name:           "Invalid Input Empty NullInt32",
			scanInput:      "non_int32",
			nullInt32Input: &NullInt32{},
			expected: &NullInt32{
				Int32: 0,
				Valid: false,
			},
			expectedErr: fmt.Errorf("NullInt32 received non-int32 type (string) during Scan"),
		},
		{
			name:      "Invalid Input Non-Empty NullInt32",
			scanInput: int64(64),
			nullInt32Input: &NullInt32{
				Int32: 64,
				Valid: true,
			},
			expected: &NullInt32{
				Int32: 64,
				Valid: true,
			},
			expectedErr: fmt.Errorf("NullInt32 received non-int32 type (int64) during Scan"),
		},
		{
			name:           "Nil Input Empty NullInt32",
			scanInput:      nil,
			nullInt32Input: &NullInt32{},
			expected: &NullInt32{
				Int32: 0,
				Valid: false,
			},
			expectedErr: nil,
		},
		{
			name:      "Nil Input Non-Empty NullInt32",
			scanInput: nil,
			nullInt32Input: &NullInt32{
				Int32: 32,
				Valid: true,
			},
			expected: &NullInt32{
				Int32: 0,
				Valid: false,
			},
			expectedErr: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			// Act
			err := test.nullInt32Input.Scan(test.scanInput)

			// Assert
			assert.Equal(t, test.expected, test.nullInt32Input)
			assert.Equal(t, test.expectedErr, err)
		})
	}
}

func TestNullInt16_Scan(t *testing.T) {

	tests := []struct {
		name           string
		scanInput      interface{}
		nullInt16Input *NullInt16
		expected       *NullInt16
		expectedErr    error
	}{
		{
			name:           "Valid Int16 Empty NullInt16",
			scanInput:      int16(64),
			nullInt16Input: &NullInt16{},
			expected: &NullInt16{
				Int16: 64,
				Valid: true,
			},
			expectedErr: nil,
		},
		{
			name:      "Valid Int16 Non-Empty NullInt16",
			scanInput: int16(64),
			nullInt16Input: &NullInt16{
				Int16: 32,
				Valid: false,
			},
			expected: &NullInt16{
				Int16: 64,
				Valid: true,
			},
			expectedErr: nil,
		},
		{
			name:           "Invalid Input Empty NullInt16",
			scanInput:      "non_int16",
			nullInt16Input: &NullInt16{},
			expected: &NullInt16{
				Int16: 0,
				Valid: false,
			},
			expectedErr: fmt.Errorf("NullInt16 received non-int16 type (string) during Scan"),
		},
		{
			name:      "Invalid Input Non-Empty NullInt16",
			scanInput: int64(64),
			nullInt16Input: &NullInt16{
				Int16: 64,
				Valid: true,
			},
			expected: &NullInt16{
				Int16: 64,
				Valid: true,
			},
			expectedErr: fmt.Errorf("NullInt16 received non-int16 type (int64) during Scan"),
		},
		{
			name:           "Nil Input Empty NullInt16",
			scanInput:      nil,
			nullInt16Input: &NullInt16{},
			expected: &NullInt16{
				Int16: 0,
				Valid: false,
			},
			expectedErr: nil,
		},
		{
			name:      "Nil Input Non-Empty NullInt16",
			scanInput: nil,
			nullInt16Input: &NullInt16{
				Int16: 32,
				Valid: true,
			},
			expected: &NullInt16{
				Int16: 0,
				Valid: false,
			},
			expectedErr: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			// Act
			err := test.nullInt16Input.Scan(test.scanInput)

			// Assert
			assert.Equal(t, test.expected, test.nullInt16Input)
			assert.Equal(t, test.expectedErr, err)
		})
	}
}

func TestNullByte_Scan(t *testing.T) {

	tests := []struct {
		name          string
		scanInput     interface{}
		nullByteInput *NullByte
		expected      *NullByte
		expectedErr   error
	}{
		{
			name:          "Valid Byte Empty NullByte",
			scanInput:     byte('i'),
			nullByteInput: &NullByte{},
			expected: &NullByte{
				Byte:  byte('i'),
				Valid: true,
			},
			expectedErr: nil,
		},
		{
			name:      "Valid Byte Non-Empty NullByte",
			scanInput: byte(64),
			nullByteInput: &NullByte{
				Byte:  32,
				Valid: false,
			},
			expected: &NullByte{
				Byte:  64,
				Valid: true,
			},
			expectedErr: nil,
		},
		{
			name:          "Invalid Input Empty NullByte",
			scanInput:     "non_byte",
			nullByteInput: &NullByte{},
			expected: &NullByte{
				Byte:  0,
				Valid: false,
			},
			expectedErr: fmt.Errorf("NullByte received non-byte type (string) during Scan"),
		},
		{
			name:      "Invalid Input Non-Empty NullByte",
			scanInput: int64(64),
			nullByteInput: &NullByte{
				Byte:  16,
				Valid: true,
			},
			expected: &NullByte{
				Byte:  16,
				Valid: true,
			},
			expectedErr: fmt.Errorf("NullByte received non-byte type (int64) during Scan"),
		},
		{
			name:          "Nil Input Empty NullByte",
			scanInput:     nil,
			nullByteInput: &NullByte{},
			expected: &NullByte{
				Byte:  0,
				Valid: false,
			},
			expectedErr: nil,
		},
		{
			name:      "Nil Input Non-Empty NullByte",
			scanInput: nil,
			nullByteInput: &NullByte{
				Byte:  32,
				Valid: true,
			},
			expected: &NullByte{
				Byte:  0,
				Valid: false,
			},
			expectedErr: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			// Act
			err := test.nullByteInput.Scan(test.scanInput)

			// Assert
			assert.Equal(t, test.expected, test.nullByteInput)
			assert.Equal(t, test.expectedErr, err)
		})
	}
}

func TestNullFloat64_Scan(t *testing.T) {

	tests := []struct {
		name             string
		scanInput        interface{}
		nullFloat64Input *NullFloat64
		expected         *NullFloat64
		expectedErr      error
	}{
		{
			name:             "Valid Float64 Empty NullFloat64",
			scanInput:        3.14159265,
			nullFloat64Input: &NullFloat64{},
			expected: &NullFloat64{
				Float64: 3.14159265,
				Valid:   true,
			},
			expectedErr: nil,
		},
		{
			name:      "Valid Float64 Non-Empty NullFloat64",
			scanInput: 3.14159265,
			nullFloat64Input: &NullFloat64{
				Float64: 63.79,
				Valid:   false,
			},
			expected: &NullFloat64{
				Float64: 3.14159265,
				Valid:   true,
			},
			expectedErr: nil,
		},
		{
			name:             "Invalid Input Empty NullFloat64",
			scanInput:        "non_float64",
			nullFloat64Input: &NullFloat64{},
			expected: &NullFloat64{
				Float64: 0,
				Valid:   false,
			},
			expectedErr: fmt.Errorf("NullFloat64 received non-float64 type (string) during Scan"),
		},
		{
			name:      "Invalid Input Non-Empty NullFloat64",
			scanInput: int64(64),
			nullFloat64Input: &NullFloat64{
				Float64: 3.14159265,
				Valid:   true,
			},
			expected: &NullFloat64{
				Float64: 3.14159265,
				Valid:   true,
			},
			expectedErr: fmt.Errorf("NullFloat64 received non-float64 type (int64) during Scan"),
		},
		{
			name:             "Nil Input Empty NullFloat64",
			scanInput:        nil,
			nullFloat64Input: &NullFloat64{},
			expected: &NullFloat64{
				Float64: 0,
				Valid:   false,
			},
			expectedErr: nil,
		},
		{
			name:      "Nil Input Non-Empty NullFloat64",
			scanInput: nil,
			nullFloat64Input: &NullFloat64{
				Float64: 32,
				Valid:   true,
			},
			expected: &NullFloat64{
				Float64: 0,
				Valid:   false,
			},
			expectedErr: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			// Act
			err := test.nullFloat64Input.Scan(test.scanInput)

			// Assert
			assert.Equal(t, test.expected, test.nullFloat64Input)
			assert.Equal(t, test.expectedErr, err)
		})
	}
}

func TestNullBool_Scan(t *testing.T) {

	tests := []struct {
		name          string
		scanInput     interface{}
		nullBoolInput *NullBool
		expected      *NullBool
		expectedErr   error
	}{
		{
			name:          "Valid Bool Empty NullBool",
			scanInput:     true,
			nullBoolInput: &NullBool{},
			expected: &NullBool{
				Bool:  true,
				Valid: true,
			},
			expectedErr: nil,
		},
		{
			name:      "Valid Bool Non-Empty NullBool",
			scanInput: false,
			nullBoolInput: &NullBool{
				Bool:  true,
				Valid: true,
			},
			expected: &NullBool{
				Bool:  false,
				Valid: true,
			},
			expectedErr: nil,
		},
		{
			name:          "Invalid Input Empty NullBool",
			scanInput:     "non_bool",
			nullBoolInput: &NullBool{},
			expected: &NullBool{
				Bool:  false,
				Valid: false,
			},
			expectedErr: fmt.Errorf("NullBool received non-bool type (string) during Scan"),
		},
		{
			name:      "Invalid Input Non-Empty NullBool",
			scanInput: int64(64),
			nullBoolInput: &NullBool{
				Bool:  true,
				Valid: true,
			},
			expected: &NullBool{
				Bool:  true,
				Valid: true,
			},
			expectedErr: fmt.Errorf("NullBool received non-bool type (int64) during Scan"),
		},
		{
			name:          "Nil Input Empty NullBool",
			scanInput:     nil,
			nullBoolInput: &NullBool{},
			expected: &NullBool{
				Bool:  false,
				Valid: false,
			},
			expectedErr: nil,
		},
		{
			name:      "Nil Input Non-Empty NullBool",
			scanInput: nil,
			nullBoolInput: &NullBool{
				Bool:  true,
				Valid: true,
			},
			expected: &NullBool{
				Bool:  false,
				Valid: false,
			},
			expectedErr: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			// Act
			err := test.nullBoolInput.Scan(test.scanInput)

			// Assert
			assert.Equal(t, test.expected, test.nullBoolInput)
			assert.Equal(t, test.expectedErr, err)
		})
	}
}

func TestNullTime_Scan(t *testing.T) {

	testTime := time.Date(2022, time.August, 22, 12, 45, 36, 239839283, time.UTC)

	tests := []struct {
		name          string
		scanInput     interface{}
		nullTimeInput *NullTime
		expected      *NullTime
		expectedErr   error
	}{
		{
			name:          "Valid Time Empty NullTime",
			scanInput:     testTime,
			nullTimeInput: &NullTime{},
			expected: &NullTime{
				Time:  testTime,
				Valid: true,
			},
			expectedErr: nil,
		},
		{
			name:      "Valid Time Non-Empty NullTime",
			scanInput: testTime,
			nullTimeInput: &NullTime{
				Time:  time.Now(),
				Valid: false,
			},
			expected: &NullTime{
				Time:  testTime,
				Valid: true,
			},
			expectedErr: nil,
		},
		{
			name:          "Invalid Input Empty NullTime",
			scanInput:     "non_time",
			nullTimeInput: &NullTime{},
			expected: &NullTime{
				Time:  time.Time{},
				Valid: false,
			},
			expectedErr: fmt.Errorf("NullTime received non-time.Time type (string) during Scan"),
		},
		{
			name:      "Invalid Input Non-Empty NullTime",
			scanInput: int64(64),
			nullTimeInput: &NullTime{
				Time:  testTime,
				Valid: true,
			},
			expected: &NullTime{
				Time:  testTime,
				Valid: true,
			},
			expectedErr: fmt.Errorf("NullTime received non-time.Time type (int64) during Scan"),
		},
		{
			name:          "Nil Input Empty NullTime",
			scanInput:     nil,
			nullTimeInput: &NullTime{},
			expected: &NullTime{
				Time:  time.Time{},
				Valid: false,
			},
			expectedErr: nil,
		},
		{
			name:      "Nil Input Non-Empty NullTime",
			scanInput: nil,
			nullTimeInput: &NullTime{
				Time:  testTime,
				Valid: true,
			},
			expected: &NullTime{
				Time:  time.Time{},
				Valid: false,
			},
			expectedErr: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			// Act
			err := test.nullTimeInput.Scan(test.scanInput)

			// Assert
			assert.Equal(t, test.expected, test.nullTimeInput)
			assert.Equal(t, test.expectedErr, err)
		})
	}
}
