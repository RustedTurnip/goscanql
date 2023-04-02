package goscanql

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewFieldsSlice(t *testing.T) {

	tests := []struct {
		name    string
		slice   interface{}
		fields  *fields
		compare func(interface{}, interface{}) bool
	}{
		{
			name:   "Normal Int Slice Pointer",
			slice:  &[]int{},
			fields: &fields{},
			compare: func(input, expected interface{}) bool {
				return input.(*[]int) == expected.(*[]int)
			},
		},
	}

	for _, test := range tests {

		msg := fmt.Sprintf("%s: failed", test.name)

		result := newFieldsSlice(test.slice, test.fields)

		// assert that the fieldsSlice sliceRef points to the same slice as the test's
		// initial slice
		assert.Samef(t, test.slice, result.sliceRef, msg)

		// assert that the fieldsSlice fields references includes only the fields provided
		// to the constructor
		assert.Lenf(t, result.fields, 1, msg)
		assert.Samef(t, test.fields, result.fields[0], msg)
	}
}

func TestFieldsSliceAppend(t *testing.T) {

	tests := []*struct {
		name        string
		fieldsSlice *fieldsSlice
		entry       *fields
		shouldPanic bool
	}{
		{
			name: "Append To Empty Slice",
			fieldsSlice: &fieldsSlice{
				sliceRef: &[]string{},
				fields:   []*fields{},
			},
			entry: &fields{
				slice: &fieldsSlice{
					sliceRef: &[]string{
						"one",
					},
				},
			},
			shouldPanic: false,
		},
		{
			name: "Append To Populated Slice",
			fieldsSlice: &fieldsSlice{
				sliceRef: &[]int{
					1,
				},
				fields: []*fields{
					{},
				},
			},
			entry: &fields{
				slice: &fieldsSlice{
					sliceRef: &[]int{
						2,
					},
				},
			},
			shouldPanic: false,
		},
		{
			name: "Append With Mismatched Slice Types",
			fieldsSlice: &fieldsSlice{
				sliceRef: &[]string{
					"one",
				},
				fields: []*fields{
					{},
				},
			},
			entry: &fields{
				slice: &fieldsSlice{
					sliceRef: &[]int{
						2,
					},
				},
			},
			shouldPanic: true,
		},
	}

	for _, test := range tests {

		msg := fmt.Sprintf("%s: failed", test.name)

		// take snapshot values for later comparison
		originalLength := reflect.ValueOf(test.fieldsSlice.sliceRef).Elem().Len()
		originalValue := reflect.ValueOf(test.fieldsSlice.sliceRef).Interface()

		originalFieldsLength := len(test.fieldsSlice.fields)

		// run subject under test
		if test.shouldPanic {
			assert.Panicsf(
				t,
				func() {
					test.fieldsSlice.append(test.entry)
				},
				msg)

			continue
		}
		test.fieldsSlice.append(test.entry)

		// assert that the underlying fieldsSlice slice has been affected as expected
		assert.Equalf(t, originalLength+1, reflect.ValueOf(test.fieldsSlice.sliceRef).Elem().Len(), msg)
		assert.Same(t, originalValue, reflect.ValueOf(test.fieldsSlice.sliceRef).Interface(), msg)

		// assert that the fieldsSlice fields slice has been affected as expected
		assert.Equalf(t, originalFieldsLength+1, len(test.fieldsSlice.fields), msg)
		assert.Samef(t, test.entry, test.fieldsSlice.fields[len(test.fieldsSlice.fields)-1], "")
	}
}

func TestGetExisting(t *testing.T) {

	tests := []struct {
		name           string
		fieldsSlice    *fieldsSlice
		lookup         *fields
		expectedFields func([]*fields) *fields
	}{
		{
			name: "Exists Existing Fields Entity",
			fieldsSlice: &fieldsSlice{
				fields: []*fields{
					{
						orderedFieldNames: []string{
							"fieldOne",
						},
						references: map[string]interface{}{
							"fieldOne": &[]string{
								"a",
							},
						},
					},
				},
			},
			lookup: &fields{
				orderedFieldNames: []string{
					"fieldOne",
				},
				references: map[string]interface{}{
					"fieldOne": &[]string{
						"a",
					},
				},
			},
			expectedFields: func(f []*fields) *fields {
				return f[0]
			},
		},
		{
			name: "Exists Empty Fields Slice",
			fieldsSlice: &fieldsSlice{
				fields: []*fields{},
			},
			lookup: &fields{
				orderedFieldNames: []string{
					"fieldOne",
				},
				references: map[string]interface{}{
					"fieldOne": &[]string{
						"",
					},
				},
			},
			expectedFields: func(f []*fields) *fields {
				return nil
			},
		},
		{
			name: "Exists Fields Entry Not In Populated Fields Slice",
			fieldsSlice: &fieldsSlice{
				fields: []*fields{
					{
						orderedFieldNames: []string{
							"fieldOne",
						},
						references: map[string]interface{}{
							"fieldOne": &[]string{
								"a",
							},
						},
					},
				},
			},
			lookup: &fields{
				orderedFieldNames: []string{
					"fieldOne",
				},
				references: map[string]interface{}{
					"fieldOne": &[]string{
						"not a",
					},
				},
			},
			expectedFields: func(f []*fields) *fields {
				return nil
			},
		},
	}

	for _, test := range tests {

		msg := fmt.Sprintf("%s: failed", test.name)

		// assert that the expected existing *fields is referenced by comparing the pointer address
		// of the fields in the fieldsSlice fields slice to the resulting getExisting response
		result := test.fieldsSlice.getExisting(test.lookup)
		expected := test.expectedFields(test.fieldsSlice.fields)
		assert.Samef(t, expected, result, msg)
	}
}

func TestEmpty(t *testing.T) {

	tests := []struct {
		name        string
		fieldsSlice *fieldsSlice
	}{
		{
			name: "Empty Slice With Elements",
			fieldsSlice: &fieldsSlice{
				sliceRef: &[]string{
					"one",
					"two",
				},
			},
		},
		{
			name: "Empty Slice With No Elements",
			fieldsSlice: &fieldsSlice{
				sliceRef: &[]string{},
			},
		},
	}

	for _, test := range tests {

		originalPointer := reflect.ValueOf(test.fieldsSlice.sliceRef).Interface()

		// execute sut
		test.fieldsSlice.empty()

		// assert that the original pointer (sliceRef) still points to the same location
		assert.Samef(t, originalPointer, test.fieldsSlice.sliceRef, "")

		// assert that the emptied slice is initialised (not nil)
		assert.NotNilf(t, test.fieldsSlice.sliceRef, "")

		// assert that the length of the slice is now 0
		assert.Equalf(t, 0, reflect.ValueOf(test.fieldsSlice.sliceRef).Elem().Len(), "")
	}
}
