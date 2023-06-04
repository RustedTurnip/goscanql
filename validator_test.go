package goscanql

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestIsStruct(t *testing.T) {

	tests := []struct {
		name     string
		input    interface{}
		expected error
	}{
		{
			name:     "Struct_NoError",
			input:    struct{}{},
			expected: nil,
		},
		{
			name:     "PointerToStruct_NoError",
			input:    &struct{}{},
			expected: nil,
		},
		{
			name:     "PointerToPointerToStruct_NoError",
			input:    referenceField(&struct{}{}),
			expected: nil,
		},
		{
			name:     "Map_ProducesError",
			input:    map[string]int{},
			expected: fmt.Errorf("input type (map[string]int) must be of type struct or pointer to struct"),
		},
		{
			name:     "PrimitiveInt_ProducesError",
			input:    5,
			expected: fmt.Errorf("input type (int) must be of type struct or pointer to struct"),
		},
		{
			name:     "PointerToPrimitive_ProducesError",
			input:    referenceField("Hello"),
			expected: fmt.Errorf("input type (string) must be of type struct or pointer to struct"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			// Arrange
			rType := reflect.TypeOf(test.input)

			// Act
			result := isStruct(rType)

			// Assert
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestIsNotArray(t *testing.T) {

	tests := []struct {
		name     string
		input    interface{}
		expected error
	}{
		{
			name:     "Array_ProducesError",
			input:    [4]int{},
			expected: fmt.Errorf("arrays are not supported ([4]int), consider using a slice instead"),
		},
		{
			name:     "MultiDimensionalArray_ProducesError",
			input:    [4][4]int{},
			expected: fmt.Errorf("arrays are not supported ([4][4]int), consider using a slice instead"),
		},
		{
			name:     "MultiDimensionalSliceArray_ProducesError",
			input:    [][4]int{},
			expected: fmt.Errorf("arrays are not supported ([4]int), consider using a slice instead"),
		},
		{
			name:     "MultiDimensionalPointerSlicePointerArray_ProducesError",
			input:    referenceField([]*[4]int{}),
			expected: fmt.Errorf("arrays are not supported ([4]int), consider using a slice instead"),
		},
		{
			name:     "PointerToArray_ProducesError",
			input:    &[4]int{},
			expected: fmt.Errorf("arrays are not supported ([4]int), consider using a slice instead"),
		},
		{
			name:     "NonArray_NoError",
			input:    struct{}{},
			expected: nil,
		},
		{
			name:     "Slice_NoError",
			input:    []struct{}{},
			expected: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			// Arrange
			rType := reflect.TypeOf(test.input)

			// Act
			result := isNotArray(rType)

			// Assert
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestIsNotMap(t *testing.T) {

	tests := []struct {
		name     string
		input    interface{}
		expected error
	}{
		{
			name:     "Map_ProducesError",
			input:    map[string]interface{}{},
			expected: fmt.Errorf("maps are not supported (map[string]interface {}), consider using a slice instead"),
		},
		{
			name:     "PointerToMap_ProducesError",
			input:    &map[string]interface{}{},
			expected: fmt.Errorf("maps are not supported (map[string]interface {}), consider using a slice instead"),
		},
		{
			name:     "SliceMap_ProducesError",
			input:    []map[string]interface{}{},
			expected: fmt.Errorf("maps are not supported (map[string]interface {}), consider using a slice instead"),
		},
		{
			name:     "MultiDimensionalSliceMap_ProducesError",
			input:    [][]map[string]interface{}{},
			expected: fmt.Errorf("maps are not supported (map[string]interface {}), consider using a slice instead"),
		},
		{
			name:     "ArrayMap_ProducesError",
			input:    [4]map[string]interface{}{},
			expected: fmt.Errorf("maps are not supported (map[string]interface {}), consider using a slice instead"),
		},
		{
			name:     "SliceArrayMap_ProducesError",
			input:    [][4]map[string]interface{}{},
			expected: fmt.Errorf("maps are not supported (map[string]interface {}), consider using a slice instead"),
		},
		{
			name:     "Struct_NoError",
			input:    struct{}{},
			expected: nil,
		},
		{
			name:     "PrimitiveType_NoError",
			input:    "",
			expected: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			// Arrange
			rType := reflect.TypeOf(test.input)

			// Act
			result := isNotMap(rType)

			// Assert
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestIsNotMultidimensionalSlice(t *testing.T) {

	tests := []struct {
		name     string
		input    interface{}
		expected error
	}{
		{
			name:     "MultidimensionalSlice_ProducesError",
			input:    [][]int{},
			expected: fmt.Errorf("multi-dimensional slices are not supported ([][]int), consider using a slice instead"),
		},
		{
			name:     "PointerMultidimensionalPointerSlice_ProducesError",
			input:    &[]*[]int{},
			expected: fmt.Errorf("multi-dimensional slices are not supported ([]*[]int), consider using a slice instead"),
		},
		{
			name:     "MultidimensionalPointerSlice_ProducesError",
			input:    &[]*[]int{},
			expected: fmt.Errorf("multi-dimensional slices are not supported ([]*[]int), consider using a slice instead"),
		},
		{
			name:     "ThreeDimensionalPointerSlice_ProducesError",
			input:    [][][]int{},
			expected: fmt.Errorf("multi-dimensional slices are not supported ([][][]int), consider using a slice instead"),
		},
		{
			name:     "Slice_NoError",
			input:    []int{},
			expected: nil,
		},
		{
			name:     "PointerSlice_NoError",
			input:    &[]int{},
			expected: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			// Arrange
			rType := reflect.TypeOf(test.input)

			// Act
			result := isNotMultidimensionalSlice(rType)

			// Assert
			assert.Equal(t, test.expected, result)
		})
	}
}
