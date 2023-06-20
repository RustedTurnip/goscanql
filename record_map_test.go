package goscanql

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func Test_fieldByTag(t *testing.T) {

	testInputs := map[string]interface{}{
		"NormalGoscanqlTaggedStruct": struct {
			Foo       string   `goscanql:"foo"`
			Bar       int      `goscanql:"bar"`
			Arbitrary struct{} `goscanql:"arbitrary"`
		}{},
		"OtherTaggedStruct": struct {
			Foo       string   `json:"foo"`
			Bar       int      `json:"bar"`
			Arbitrary struct{} `json:"arbitrary"`
		}{},
		"NonTaggedStruct": struct {
			Foo       string
			Bar       int
			Arbitrary struct{}
		}{},
		"NestedTaggedStruct": struct {
			Bar       int `goscanql:"bar"`
			Arbitrary struct {
				Foo string `goscanql:"foo"`
			} `goscanql:"arbitrary"`
		}{},
	}

	tests := []struct {
		name          string
		inputTag      string
		inputValueKey string
		expected      interface{}
	}{
		{
			name:          "GivenGoscanqlTag_ThenFieldValueReturned",
			inputTag:      "foo",
			inputValueKey: "NormalGoscanqlTaggedStruct",
			expected: referenceField(reflect.ValueOf(testInputs["NormalGoscanqlTaggedStruct"]).
				FieldByName("Foo")),
		},
		{
			name:          "GivenOtherTaggedStruct_ThenNilReturned",
			inputTag:      "foo",
			inputValueKey: "OtherTaggedStruct",
			expected:      nil,
		},
		{
			name:          "GivenNonTaggedStruct_ThenNilReturned",
			inputTag:      "foo",
			inputValueKey: "NonTaggedStruct",
			expected:      nil,
		},
		{
			name:          "GivenNestedTaggedStruct_ThenNilReturned",
			inputTag:      "foo",
			inputValueKey: "NestedTaggedStruct",
			expected:      nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			// Arrange
			inputValue := reflect.ValueOf(testInputs[test.inputValueKey])

			// Act
			result := fieldByTag(test.inputTag, inputValue)

			// Assert

			// this must be assessed differently as test.expected is nil of type interface, and result
			// is nil of type reflect.Value, which is perceived as a mismatch. See here for more info:
			// https://stackoverflow.com/a/19766621
			if test.expected == nil {
				assert.Nil(t, result)
				return
			}

			assert.Equal(t, test.expected, result)
		})
	}
}

func Test_getRootValue(t *testing.T) {

	rootPrimitive := 0

	tests := []struct {
		name  string
		input interface{}
	}{
		{
			name:  "GivenPointerToPrimitive_ThenPrimitiveIsReturned",
			input: &rootPrimitive,
		},
		{
			name:  "GivenPointerToPointerToPrimitive_ThenPrimitiveIsReturned",
			input: referenceField(&rootPrimitive),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			// Arrange
			input := reflect.ValueOf(test.input)
			// rootPrimitive must be passed in as pointer otherwise reflect.Value will be created for a copy of
			// rootPrimitive instead of the original
			expected := reflect.ValueOf(&rootPrimitive).Elem()

			// Act
			result := getRootValue(input)

			// Assert
			assert.Equal(t, expected, result)
		})
	}

}
