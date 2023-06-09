package goscanql

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

type cyclicExample struct {
	Str   string         `goscanql:"str"`
	Cycle *cyclicExample `goscanql:"cycle"`
}

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

func TestIsNotFunc(t *testing.T) {

	tests := []struct {
		name     string
		input    interface{}
		expected error
	}{
		{
			name:     "Func_ProducesError",
			input:    func() {},
			expected: fmt.Errorf("functions are not supported (func())"),
		},
		{
			name:     "PointerToFunc_ProducesError",
			input:    referenceField(func() {}),
			expected: fmt.Errorf("functions are not supported (func())"),
		},
		{
			name:     "SliceFunc_ProducesError",
			input:    []func(){},
			expected: fmt.Errorf("functions are not supported (func())"),
		},
		{
			name:     "MultiDimensionalSliceFunc_ProducesError",
			input:    [][]func(){},
			expected: fmt.Errorf("functions are not supported (func())"),
		},
		{
			name:     "ArrayFunc_ProducesError",
			input:    [4]func(){},
			expected: fmt.Errorf("functions are not supported (func())"),
		},
		{
			name:     "SliceArrayFunc_ProducesError",
			input:    [][4]func(){},
			expected: fmt.Errorf("functions are not supported (func())"),
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
			result := isNotFunc(rType)

			// Assert
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestIsNotChan(t *testing.T) {

	tests := []struct {
		name     string
		input    interface{}
		expected error
	}{
		{
			name:     "Chan_ProducesError",
			input:    make(chan int),
			expected: fmt.Errorf("channels are not supported (chan int)"),
		},
		{
			name:     "PointerToChan_ProducesError",
			input:    referenceField(make(chan int)),
			expected: fmt.Errorf("channels are not supported (chan int)"),
		},
		{
			name:     "SliceChan_ProducesError",
			input:    []chan int{},
			expected: fmt.Errorf("channels are not supported (chan int)"),
		},
		{
			name:     "MultiDimensionalSliceChan_ProducesError",
			input:    [][]chan int{},
			expected: fmt.Errorf("channels are not supported (chan int)"),
		},
		{
			name:     "ArrayChan_ProducesError",
			input:    [4]chan int{},
			expected: fmt.Errorf("channels are not supported (chan int)"),
		},
		{
			name:     "SliceArrayChan_ProducesError",
			input:    [][4]chan int{},
			expected: fmt.Errorf("channels are not supported (chan int)"),
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
			result := isNotChan(rType)

			// Assert
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestValidateType(t *testing.T) {

	tests := []struct {
		name     string
		input    interface{}
		expected error
	}{
		{
			name: "StructInput_NoError",
			input: struct {
				AValidField struct{} `goscanql:"a_valid_field"`
			}{},
			expected: nil,
		},
		{
			name: "PointerStructInput_NoError",
			input: &struct {
				AValidField struct{} `goscanql:"a_valid_field"`
			}{},
			expected: nil,
		},
		{
			name:     "NonStructInput_ProducesError",
			input:    []int{},
			expected: fmt.Errorf("input type ([]int) must be of type struct or pointer to struct"),
		},
		{
			name:     "SliceStructInput_ProducesError",
			input:    []struct{}{},
			expected: fmt.Errorf("input type ([]struct {}) must be of type struct or pointer to struct"),
		},
		{
			name: "StructWithArrayInput_ProducesError",
			input: struct {
				A [4]int `goscanql:"a"`
			}{},
			expected: fmt.Errorf("arrays are not supported ([4]int), consider using a slice instead"),
		},
		{
			name: "StructWithMapInput_ProducesError",
			input: struct {
				M map[string]interface{} `goscanql:"m"`
			}{},
			expected: fmt.Errorf("maps are not supported (map[string]interface {}), consider using a slice instead"),
		},
		{
			name: "StructWithMultiDimensionalSliceInput_ProducesError",
			input: struct {
				MS [][]struct{} `goscanql:"ms"`
			}{},
			expected: fmt.Errorf("multi-dimensional slices are not supported ([][]struct {}), consider using a slice instead"),
		},
		{
			name: "StructWithFuncInput_ProducesError",
			input: struct {
				Fn func() `goscanql:"fn"`
			}{},
			expected: fmt.Errorf("functions are not supported (func())"),
		},
		{
			name: "StructWithChanInput_ProducesError",
			input: struct {
				Ch chan int `goscanql:"ch"`
			}{},
			expected: fmt.Errorf("channels are not supported (chan int)"),
		},
		{
			name: "StructCycleInput_ProducesError",
			input: struct {
				EC cyclicExample `goscanql:"ec"`
			}{},
			expected: fmt.Errorf("goscanql does not support cyclic structs: struct { EC goscanql.cyclicExample \"goscanql:\\\"ec\\\"\" }"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			// Act
			result := validateType(test.input)

			// Assert
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestGetPointerRootType(t *testing.T) {

	tests := []struct {
		name     string
		input    interface{}
		expected interface{}
	}{
		{
			name:     "NonPointerType_ReturnsSame",
			input:    1,
			expected: 1,
		},
		{
			name:     "PointerType_ReturnsRoot",
			input:    referenceField(1),
			expected: 1,
		},
		{
			name:     "SliceType_ReturnsSlice",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "PointerSliceType_ReturnsSlice",
			input:    &[]string{},
			expected: []string{},
		},
		{
			name:     "PointerPointerSliceType_ReturnsSlice",
			input:    referenceField(&[]string{}),
			expected: []string{},
		},
		{
			name:     "PointerMapType_ReturnsSlice",
			input:    &map[string]interface{}{},
			expected: map[string]interface{}{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			// Assemble
			input := reflect.TypeOf(test.input)
			expected := reflect.TypeOf(test.expected)

			// Act
			result := getPointerRootType(input)

			// Assert
			assert.Equal(t, expected, result)
		})
	}
}

func TestGetSliceRootType(t *testing.T) {

	tests := []struct {
		name     string
		input    interface{}
		expected interface{}
	}{
		{
			name:     "NonSlice_ReturnsSame",
			input:    1, // int
			expected: 1, // int
		},
		{
			name:     "Slice_ReturnsSliceType",
			input:    []int{},
			expected: 1, // int
		},
		{
			name:     "SliceOfIntSlices_ReturnsIntType",
			input:    [][]int{},
			expected: 1, // int
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			// Assemble
			input := reflect.TypeOf(test.input)
			expected := reflect.TypeOf(test.expected)

			// Act
			result := getSliceRootType(input)

			// Assert
			assert.Equal(t, expected, result)
		})
	}
}

type extraNestedCycleExample struct {
	I     int                           `goscanql:"i"`
	ENCED extraNestedCycleExampleNested `goscanql:"enced"`
}

type extraNestedCycleExampleNested struct {
	ENCE *extraNestedCycleExampleNested `goscanql:"ence"`
}

func TestVerifyNoCycles(t *testing.T) {

	tests := []struct {
		name     string
		input    interface{}
		expected error
	}{
		{
			name:     "NonStructType_NoError",
			input:    1, // int
			expected: nil,
		},
		{
			name: "CyclicStruct_ProducesError",
			input: struct {
				Str string         `goscanql:"str"`
				CE  *cyclicExample `goscanql:"ce"`
			}{},
			expected: fmt.Errorf("goscanql does not support cyclic structs: struct { Str string \"goscanql:\\\"str\\\"\"; CE *goscanql.cyclicExample \"goscanql:\\\"ce\\\"\" }"),
		},
		{
			name:     "NestedCyclicStruct_ProducesError",
			input:    extraNestedCycleExample{},
			expected: fmt.Errorf("goscanql does not support cyclic structs: goscanql.extraNestedCycleExample"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			// Assemble
			input := reflect.TypeOf(test.input)

			// Act
			result := verifyNoCycles(input)

			// Assert
			assert.Equal(t, test.expected, result)
		})
	}
}
