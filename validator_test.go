package goscanql

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

type cyclicExample struct {
	Str   string         `sql:"str"`
	Cycle *cyclicExample `sql:"cycle"`
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

type arrayScanner [4]string

func (a arrayScanner) Scan(_ interface{}) error {
	return nil
}

func (m arrayScanner) ID() []byte {
	return nil
}

type arrayType [4]string

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
			name:     "ArrayOfArrayScanners_ProducesError",
			input:    [6]arrayScanner{},
			expected: fmt.Errorf("arrays are not supported ([6]goscanql.arrayScanner), consider using a slice instead"),
		},
		{
			name:     "ArrayType_ProducesError",
			input:    arrayType{},
			expected: fmt.Errorf("arrays are not supported (goscanql.arrayType), consider using a slice instead"),
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
		{
			name:     "ArrayScanner_NoError",
			input:    arrayScanner{},
			expected: nil,
		},
		{
			name:     "SliceOfArrayScanners_NoError",
			input:    []arrayScanner{},
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

type mapScanner map[int]string

func (m mapScanner) Scan(_ interface{}) error {
	return nil
}

func (m mapScanner) ID() []byte {
	return nil
}

type mapType map[int]string

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
			name:     "MapType_ProducesError",
			input:    mapType{},
			expected: fmt.Errorf("maps are not supported (goscanql.mapType), consider using a slice instead"),
		},
		{
			name:     "SliceOfMapTypes_ProducesError",
			input:    []mapType{},
			expected: fmt.Errorf("maps are not supported (goscanql.mapType), consider using a slice instead"),
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
		{
			name:     "MapScanner_NoError",
			input:    mapScanner{},
			expected: nil,
		},
		{
			name:     "SliceOfMapScanners_NoError",
			input:    []mapScanner{},
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

type multidimensionalSliceScanner [][]string

func (m multidimensionalSliceScanner) Scan(_ interface{}) error {
	return nil
}

func (m multidimensionalSliceScanner) ID() []byte {
	return nil
}

type multidimensionalSliceType [][]string

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
			name:     "MultidimensionalSliceType_ProducesError",
			input:    multidimensionalSliceType{},
			expected: fmt.Errorf("multi-dimensional slices are not supported (goscanql.multidimensionalSliceType), consider using a slice instead"),
		},
		{
			name:     "SliceOfMultidimensionalSliceTypes_ProducesError",
			input:    []multidimensionalSliceType{},
			expected: fmt.Errorf("multi-dimensional slices are not supported ([]goscanql.multidimensionalSliceType), consider using a slice instead"),
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
		{
			name:     "MultidimensionalSliceScanner_NoError",
			input:    multidimensionalSliceScanner{},
			expected: nil,
		},
		{
			name:     "SliceOfMultidimensionalSliceScanner_NoError",
			input:    []multidimensionalSliceScanner{},
			expected: nil,
		}, {
			name:     "MultidimensionalSliceOfMultidimensionalSliceScanner_NoError",
			input:    [][]multidimensionalSliceScanner{},
			expected: fmt.Errorf("multi-dimensional slices are not supported ([][]goscanql.multidimensionalSliceScanner), consider using a slice instead"),
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

type funcScanner func()

func (f funcScanner) Scan(_ interface{}) error {
	return nil
}

func (f funcScanner) ID() []byte {
	return nil
}

type funcType func()

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
			name:     "SliceArrayFunc_ProducesError",
			input:    [][4]func(){},
			expected: fmt.Errorf("functions are not supported (func())"),
		},
		{
			name:     "FuncType_ProducesError",
			input:    funcType(func() {}),
			expected: fmt.Errorf("functions are not supported (goscanql.funcType)"),
		},
		{
			name:     "SliceOfFuncTypes_ProducesError",
			input:    []funcType{},
			expected: fmt.Errorf("functions are not supported (goscanql.funcType)"),
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
		{
			name:     "FuncScanner_NoError",
			input:    funcScanner(func() {}),
			expected: nil,
		},
		{
			name:     "SliceOfFuncScanners_NoError",
			input:    []funcScanner{},
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

type chanScanner chan int

func (c chanScanner) Scan(_ interface{}) error {
	return nil
}

func (c chanScanner) ID() []byte {
	return nil
}

type chanType chan int

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
			name:     "ChanType_ProducesError",
			input:    chanType(make(chan int)),
			expected: fmt.Errorf("channels are not supported (goscanql.chanType)"),
		},
		{
			name:     "SliceOfChanTypes_ProducesError",
			input:    []chanType{},
			expected: fmt.Errorf("channels are not supported (goscanql.chanType)"),
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
		{
			name:     "ChanScanner_NoError",
			input:    chanScanner(make(chan int)),
			expected: nil,
		},
		{
			name:     "SliceOfChanScanners_NoError",
			input:    []chanScanner{},
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

func TestIsNotCustomInterface(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected error
	}{
		{
			name:     "CustomInterface_ProducesError",
			input:    (*Scanner)(nil),
			expected: fmt.Errorf("interface types other than interface{} are not supported (goscanql.Scanner)"),
		},
		{
			name:     "SliceOfCustomInterface_ProducesError",
			input:    []Scanner{},
			expected: fmt.Errorf("interface types other than interface{} are not supported (goscanql.Scanner)"),
		},
		{
			name:     "MultidimensionalSliceOfCustomInterface_ProducesError",
			input:    [][]Scanner{},
			expected: fmt.Errorf("interface types other than interface{} are not supported (goscanql.Scanner)"),
		},
		{
			name:     "ArrayOfCustomInterface_ProducesError",
			input:    [4]Scanner{},
			expected: fmt.Errorf("interface types other than interface{} are not supported (goscanql.Scanner)"),
		},
		{
			name:     "ArraySliceOfCustomInterface_ProducesError",
			input:    [4][]Scanner{},
			expected: fmt.Errorf("interface types other than interface{} are not supported (goscanql.Scanner)"),
		},
		{
			name:     "NonInterfaceType_NoError",
			input:    1,
			expected: nil,
		},
		{
			name:     "Interface_NoError",
			input:    (*interface{})(nil),
			expected: nil,
		},
		{
			name:     "StructThatImplementsCustomInterface_NoError",
			input:    NullTime{},
			expected: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Arrange
			rt := reflect.TypeOf(test.input)

			// Act
			err := isNotCustomInterface(rt)

			// Assert
			assert.Equal(t, test.expected, err)
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
				AValidField struct{} `sql:"a_valid_field"`
			}{},
			expected: nil,
		},
		{
			name: "PointerStructInput_NoError",
			input: &struct {
				AValidField struct{} `sql:"a_valid_field"`
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
				A [4]int `sql:"a"`
			}{},
			expected: fmt.Errorf("arrays are not supported ([4]int), consider using a slice instead"),
		},
		{
			name: "StructWithMapInput_ProducesError",
			input: struct {
				M map[string]interface{} `sql:"m"`
			}{},
			expected: fmt.Errorf("maps are not supported (map[string]interface {}), consider using a slice instead"),
		},
		{
			name: "StructWithMultiDimensionalSliceInput_ProducesError",
			input: struct {
				MS [][]struct{} `sql:"ms"`
			}{},
			expected: fmt.Errorf("multi-dimensional slices are not supported ([][]struct {}), consider using a slice instead"),
		},
		{
			name: "StructWithFuncInput_ProducesError",
			input: struct {
				Fn func() `sql:"fn"`
			}{},
			expected: fmt.Errorf("functions are not supported (func())"),
		},
		{
			name: "StructWithChanInput_ProducesError",
			input: struct {
				Ch chan int `sql:"ch"`
			}{},
			expected: fmt.Errorf("channels are not supported (chan int)"),
		},
		{
			name: "StructCycleInput_ProducesError",
			input: struct {
				EC cyclicExample `sql:"ec"`
			}{},
			expected: fmt.Errorf("goscanql does not support cyclic structs: struct { EC goscanql.cyclicExample \"sql:\\\"ec\\\"\" }"),
		},
		{
			name: "StructWithMultiDimensionalSliceScannerInput_NoError",
			input: struct {
				MS multidimensionalSliceScanner `sql:"ms"`
			}{},
			expected: nil,
		},
		{
			name: "StructWithMultiDimensionalSliceInputTypedField_ProducesError",
			input: struct {
				MS multidimensionalSliceType `sql:"ms"`
			}{},
			expected: fmt.Errorf("multi-dimensional slices are not supported (goscanql.multidimensionalSliceType), consider using a slice instead"),
		},
		{
			name: "SliceOfStructWithMultiDimensionalSliceScannerInput_NoError",
			input: struct {
				MS []multidimensionalSliceScanner `sql:"ms"`
			}{},
			expected: nil,
		},
		{
			name: "MultiDimensionalStructWithMultiDimensionalSliceScannerInput_ProducesError",
			input: struct {
				MS [][]multidimensionalSliceScanner `sql:"ms"`
			}{},
			expected: fmt.Errorf("multi-dimensional slices are not supported ([][]goscanql.multidimensionalSliceScanner), consider using a slice instead"),
		},
		{
			name: "StructWithAnyInterfaceAsField_NoError",
			input: struct {
				I interface{} `sql:"i"`
			}{},
			expected: nil,
		},
		{
			name: "StructWithNonAnyInterfaceAsField_ProducesError",
			input: struct {
				S Scanner `sql:"s"`
			}{},
			expected: fmt.Errorf("interface types other than interface{} are not supported (goscanql.Scanner)"),
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
	I     int                           `sql:"i"`
	ENCED extraNestedCycleExampleNested `sql:"enced"`
}

type extraNestedCycleExampleNested struct {
	ENCE *extraNestedCycleExampleNested `sql:"ence"`
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
				Str string         `sql:"str"`
				CE  *cyclicExample `sql:"ce"`
			}{},
			expected: fmt.Errorf("goscanql does not support cyclic structs: struct { Str string \"sql:\\\"str\\\"\"; CE *goscanql.cyclicExample \"sql:\\\"ce\\\"\" }"),
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
