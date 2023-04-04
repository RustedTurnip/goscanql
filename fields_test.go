package goscanql

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
	"time"
)

func referenceField[T any](field T) *T {
	return &field
}

func TestInitialiseFields(t *testing.T) {

	type childExample struct {
		Foo int    `goscanql:"foo"`
		Bar string `goscanql:"bar"`
	}

	objExample := &struct {
		Id                  int    `goscanql:"id"`
		Name                string `goscanql:"name"`
		UnnamedField        string
		TimeExample         time.Time       `goscanql:"time"`
		Child               childExample    `goscanql:"child"`
		ChildPointer        *childExample   `goscanql:"child_pointer"`
		ChildPointerPointer **childExample  `goscanql:"child_pointer_pointer"`
		Children            []childExample  `goscanql:"children"`
		ChildrenPointer     *[]childExample `goscanql:"children_pointer"`
	}{}

	subject := &fields{
		obj:                  objExample,
		slice:                nil,
		orderedFieldNames:    []string{},
		orderedOneToOneNames: []string{},
		references:           map[string]interface{}{},
		byteReferences:       map[string]*[]byte{},
		oneToOnes:            map[string]*fields{},
		oneToManys:           map[string]*fields{},
	}

	newExpectedChildExampleFields := func(obj interface{}, asSlice bool) *fields {
		f := &fields{
			obj:   obj,
			slice: nil,
			orderedFieldNames: []string{
				"foo",
				"bar",
			},
			orderedOneToOneNames: []string{},
			references: map[string]interface{}{
				"foo": referenceField(0),
				"bar": referenceField(""),
			},
			byteReferences: map[string]*[]byte{
				"foo": {},
				"bar": {},
			},
			oneToOnes:  map[string]*fields{},
			oneToManys: map[string]*fields{},
		}

		if asSlice {
			f.slice = &fieldsSlice{
				sliceRef: &[]childExample{
					{},
				},
				fields: []*fields{
					f,
				},
			}
		}

		return f
	}

	expected := &fields{
		obj:   objExample,
		slice: nil,
		orderedFieldNames: []string{
			"id",
			"name",
			"time",
		},
		orderedOneToOneNames: []string{
			"child",
			"child_pointer",
			"child_pointer_pointer",
		},
		references: map[string]interface{}{
			"id":   &objExample.Id,
			"name": &objExample.Name,
			"time": referenceField(time.Time{}),
		},
		byteReferences: map[string]*[]byte{
			"id":   {},
			"name": {},
			"time": {},
		},
		oneToOnes: map[string]*fields{
			"child":                 newExpectedChildExampleFields(&childExample{}, false),
			"child_pointer":         newExpectedChildExampleFields(referenceField(&childExample{}), false),
			"child_pointer_pointer": newExpectedChildExampleFields(referenceField(referenceField(&childExample{})), false),
		},
		oneToManys: map[string]*fields{
			"children":         newExpectedChildExampleFields(&childExample{}, true),
			"children_pointer": newExpectedChildExampleFields(&childExample{}, true),
		},
	}

	subject.initialise("")

	// assert that the general structure produced is as expected, this assertion does not inspect
	// the memory addresses of pointers (only the underlying values)
	assert.Equalf(t, expected, subject, "")

	// assert that all the pointers refer to fields of the original object
	assert.Samef(t, &objExample.Id, subject.references["id"], "")
	assert.Samef(t, &objExample.Name, subject.references["name"], "")
	assert.Samef(t, &objExample.Child.Foo, subject.oneToOnes["child"].references["foo"], "")
	assert.Samef(t, &objExample.Child.Bar, subject.oneToOnes["child"].references["bar"], "")
	assert.Samef(t, &objExample.ChildPointer.Foo, subject.oneToOnes["child_pointer"].references["foo"], "")
	assert.Samef(t, &objExample.ChildPointer.Bar, subject.oneToOnes["child_pointer"].references["bar"], "")
	assert.Samef(t, &(*objExample.ChildPointerPointer).Foo, subject.oneToOnes["child_pointer_pointer"].references["foo"], "")
	assert.Samef(t, &(*objExample.ChildPointerPointer).Bar, subject.oneToOnes["child_pointer_pointer"].references["bar"], "")
	assert.Samef(t, &objExample.Children[0].Foo, subject.oneToManys["children"].references["foo"], "")
	assert.Samef(t, &objExample.Children[0].Bar, subject.oneToManys["children"].references["bar"], "")
	assert.Samef(t, &(*objExample.ChildrenPointer)[0].Foo, subject.oneToManys["children_pointer"].references["foo"], "")
	assert.Samef(t, &(*objExample.ChildrenPointer)[0].Bar, subject.oneToManys["children_pointer"].references["bar"], "")
}

func TestNewFields(t *testing.T) {

	type testExample struct {
		Foo int    `goscanql:"foo"`
		Bar string `goscanql:"bar"`
	}

	testInputs := map[string]interface{}{
		"Simple Non-Slice Input": &testExample{},
		"Simple Slice Input": &[]*testExample{
			{},
		},
	}

	tests := []struct {
		name        string
		expected    *fields
		expectedErr error
	}{
		{
			name: "Simple Non-Slice Input",
			expected: &fields{
				obj: testInputs["Simple Non-Slice Input"].(*testExample),
				orderedFieldNames: []string{
					"foo",
					"bar",
				},
				orderedOneToOneNames: []string{},
				references: map[string]interface{}{
					"foo": referenceField(0),
					"bar": referenceField(""),
				},
				byteReferences: map[string]*[]byte{
					"foo": {},
					"bar": {},
				},
				oneToOnes:  map[string]*fields{},
				oneToManys: map[string]*fields{},
			},
			expectedErr: nil,
		},
		{
			name: "Simple Slice Input",
			expected: &fields{
				obj: &(*testInputs["Simple Slice Input"].(*[]*testExample))[0],
				slice: &fieldsSlice{
					sliceRef: testInputs["Simple Slice Input"],
				},
				orderedFieldNames: []string{
					"foo",
					"bar",
				},
				orderedOneToOneNames: []string{},
				references: map[string]interface{}{
					"foo": referenceField(0),
					"bar": referenceField(""),
				},
				byteReferences: map[string]*[]byte{
					"foo": {},
					"bar": {},
				},
				oneToOnes:  map[string]*fields{},
				oneToManys: map[string]*fields{},
			},
			expectedErr: nil,
		},
	}

	for _, test := range tests {

		msg := fmt.Sprintf("%s: failed", test.name)

		// post-process expected *fields where slice is supposed to be instantiated (because test cannot)
		// reference itself
		if test.expected.slice != nil {
			test.expected.slice.fields = append(test.expected.slice.fields, test.expected)
		}

		// execute sut
		result, err := newFields(testInputs[test.name])

		// assert value equality between expected and result
		assert.Equalf(t, test.expected, result, msg)
		assert.Equalf(t, test.expectedErr, err, msg)

		// if test errored, continue to next as following assertions are nullified
		if err != nil {
			continue
		}

		// assert pointer equality to ensure that the original inputs are still referenced by the
		// resulting fields
		if test.expected.slice != nil {

			// if slice, asser that the sliceRef points to the original input
			assert.Samef(t, testInputs[test.name], result.slice.sliceRef, msg)

			// and that the fields obj points to the first entry of the slice
			assert.Samef(t, reflect.ValueOf(testInputs[test.name]).Elem().Index(0).Addr().Interface(), result.obj, msg)

		} else {

			// else if not slice, assert that fields obj points directly to input
			assert.Samef(t, testInputs[test.name], result.obj, msg)
		}
	}
}

func TestAddNewChild(t *testing.T) {

	type relationship int

	const (
		oneRelationship relationship = iota
		manyRelationship
	)

	tests := []struct {
		name                 string
		inputName            string
		inputObj             interface{}
		fields               *fields
		expectedRelationship relationship
		expectedErr          error
	}{
		{
			name:      "Add New One-to-One Child Struct",
			inputName: "child",
			inputObj:  &struct{}{},
			fields: &fields{
				orderedOneToOneNames: []string{},
				oneToOnes:            map[string]*fields{},
				oneToManys:           map[string]*fields{},
			},
			expectedRelationship: oneRelationship,
			expectedErr:          nil,
		},
		{
			name:      "Add New One-to-Many Child Struct",
			inputName: "child",
			inputObj:  &[]*struct{}{},
			fields: &fields{
				oneToManys: map[string]*fields{},
			},
			expectedRelationship: manyRelationship,
			expectedErr:          nil,
		},
		{
			name:      "Add New One-to-One Child Struct With Name Collision",
			inputName: "arbitrary_name",
			inputObj:  &struct{}{},
			fields: &fields{
				orderedOneToOneNames: []string{
					"arbitrary_name",
				},
				oneToOnes: map[string]*fields{
					"arbitrary_name": nil,
				},
			},
			expectedRelationship: oneRelationship,
			expectedErr:          fmt.Errorf("child already exists with name \"%s\"", "arbitrary_name"),
		},
		{
			name:      "Add New One-to-Many Child Struct With Name Collision",
			inputName: "arbitrary_name_many",
			inputObj:  &struct{}{},
			fields: &fields{
				oneToManys: map[string]*fields{
					"arbitrary_name_many": nil,
				},
			},
			expectedRelationship: oneRelationship,
			expectedErr:          fmt.Errorf("child already exists with name \"%s\"", "arbitrary_name_many"),
		},
		{
			name:      "Add New One-to-One Child Struct With Other Relationship Name Collision",
			inputName: "arbitrary_name",
			inputObj:  &struct{}{},
			fields: &fields{
				oneToManys: map[string]*fields{
					"arbitrary_name": nil,
				},
			},
			expectedRelationship: oneRelationship,
			expectedErr:          fmt.Errorf("child already exists with name \"%s\"", "arbitrary_name"),
		},
		{
			name:      "Add New One-to-Many Child Struct With Other Relationship Name Collision",
			inputName: "arbitrary_name",
			inputObj:  &struct{}{},
			fields: &fields{
				orderedOneToOneNames: []string{
					"arbitrary_name",
				},
				oneToOnes: map[string]*fields{
					"arbitrary_name": nil,
				},
			},
			expectedRelationship: oneRelationship,
			expectedErr:          fmt.Errorf("child already exists with name \"%s\"", "arbitrary_name"),
		},
	}

	for _, test := range tests {

		msg := fmt.Sprintf("%s: failed", test.name)

		// execute sut
		err := test.fields.addNewChild(test.inputName, test.inputObj)

		// assert that test returned expected error
		assert.Equalf(t, test.expectedErr, err, msg)

		// if error returned, continue as remaining asserts are nullified
		if err != nil {
			continue
		}

		// assess result based on expected relationship
		if test.expectedRelationship == oneRelationship {

			// check that the new field has been added to one-to-one children
			assert.Containsf(t, test.fields.oneToOnes, test.inputName, msg)
			// and that it hasn't been added to the one-to-manys children
			assert.NotContainsf(t, test.fields.oneToManys, test.inputName, msg)

			// assert that the child name has been added to the
			assert.Equalf(t, test.inputName, test.fields.orderedOneToOneNames[len(test.fields.orderedOneToOneNames)-1], msg)

		} else {

			// check that the new field has been added to one-to-manys children
			assert.Containsf(t, test.fields.oneToManys, test.inputName, msg)
			// and that it hasn't been added to the one-to-ones children
			assert.NotContainsf(t, test.fields.oneToOnes, test.inputName, msg)

		}
	}
}

func TestAddField(t *testing.T) {

	tests := []struct {
		name        string
		inputName   string
		inputObj    interface{}
		fields      *fields
		expected    *fields
		expectedErr error
	}{
		{
			name:      "Add Single Field Without Collision",
			inputName: "field_name",
			inputObj:  referenceField(0),
			fields: &fields{
				orderedFieldNames: []string{},
				references:        map[string]interface{}{},
				byteReferences:    map[string]*[]byte{},
			},
			expected: &fields{
				orderedFieldNames: []string{
					"field_name",
				},
				references: map[string]interface{}{
					"field_name": referenceField(0),
				},
				byteReferences: map[string]*[]byte{
					"field_name": {},
				},
			},
			expectedErr: nil,
		},
		{
			name:      "Add Single Field With Collision",
			inputName: "field_name",
			inputObj:  referenceField(0),
			fields: &fields{
				orderedFieldNames: []string{
					"field_name",
				},
				references: map[string]interface{}{
					"field_name": referenceField(0),
				},
				byteReferences: map[string]*[]byte{
					"field_name": {},
				},
			},
			expected: &fields{
				orderedFieldNames: []string{
					"field_name",
				},
				references: map[string]interface{}{
					"field_name": referenceField(0),
				},
				byteReferences: map[string]*[]byte{
					"field_name": {},
				},
			},
			expectedErr: fmt.Errorf("field with name \"field_name\" already added"),
		},
	}

	for _, test := range tests {

		err := test.fields.addField(test.inputName, test.inputObj)

		// assert that error is expected
		assert.Equalf(t, test.expectedErr, err, "")

		// continue to next if nil as following asserts are nullified
		if err != nil {
			continue
		}

		// assert that the resulting struct is the same (value-wise) as the expected fields
		assert.Equalf(t, test.expected, test.fields, "")

		// assert that the added field points to the exact same object as originally provided
		assert.Samef(t, test.inputObj, test.fields.references[test.inputName], "")
	}
}

var (
	referenceTestExample = &fields{
		references: map[string]interface{}{
			"foo": referenceField(0),
			"bar": referenceField(""),
		},
		byteReferences: map[string]*[]byte{
			"foo": {' '},
			"bar": {' '},
		},
		oneToOnes: map[string]*fields{
			"single_child": {
				references: map[string]interface{}{
					"time": referenceField(time.Time{}),
				},
				byteReferences: map[string]*[]byte{
					"time": {' '},
				},
			},
			"null_child": {
				references: map[string]interface{}{
					"time": referenceField(time.Time{}),
				},
				byteReferences: map[string]*[]byte{
					"time": {}, // as all byteReferences for this *fields are empty, it is considered nil
				},
			},
		},
		oneToManys: map[string]*fields{
			"many_children": {
				references: map[string]interface{}{
					"many_foo": referenceField(0),
					"many_bar": referenceField(""),
				},
				byteReferences: map[string]*[]byte{
					"many_foo": {' '},
					"many_bar": {' '},
				},
			},
			"null_children": {
				references: map[string]interface{}{
					"foo": referenceField(0),
				},
				byteReferences: map[string]*[]byte{
					"time": {}, // as all byteReferences for this *fields are empty, it is considered nil
				},
			},
		},
	}
)

func TestGetFieldReferences(t *testing.T) {

	expected := map[string]interface{}{
		"foo":                    referenceTestExample.references["foo"],
		"bar":                    referenceTestExample.references["bar"],
		"single_child_time":      referenceTestExample.oneToOnes["single_child"].references["time"],
		"many_children_many_foo": referenceTestExample.oneToManys["many_children"].references["many_foo"],
		"many_children_many_bar": referenceTestExample.oneToManys["many_children"].references["many_bar"],
	}

	msg := "Get Field References: failed"

	result := referenceTestExample.getFieldReferences()

	// assert that the result matches expected (by value)
	assert.Equalf(t, expected, result, msg)

	// assert that the result matches expected (by reference)
	for k, v := range expected {
		assert.Samef(t, v, result[k], msg)
	}
}

func TestGetByteReferences(t *testing.T) {

	expected := map[string]*[]byte{
		"foo":                    referenceTestExample.byteReferences["foo"],
		"bar":                    referenceTestExample.byteReferences["bar"],
		"single_child_time":      referenceTestExample.oneToOnes["single_child"].byteReferences["time"],
		"null_child_time":        referenceTestExample.oneToOnes["null_child"].byteReferences["time"],
		"many_children_many_foo": referenceTestExample.oneToManys["many_children"].byteReferences["many_foo"],
		"many_children_many_bar": referenceTestExample.oneToManys["many_children"].byteReferences["many_bar"],
		"null_children_time":     referenceTestExample.oneToManys["null_children"].byteReferences["time"],
	}

	msg := "Get Byte References: failed"

	result := referenceTestExample.getByteReferences()

	// assert that the result matches expected (by value)
	assert.Equalf(t, expected, result, msg)

	// assert that the result matches expected (by reference)
	for k, v := range expected {
		assert.Samef(t, v, result[k], msg)
	}
}

func TestCrawlFields(t *testing.T) {

	tests := []struct {
		name     string
		fn       func(map[string]*fields) func(string, *fields) bool
		expected map[string]*fields
	}{
		{
			name: "Crawl All Fields",
			fn: func(result map[string]*fields) func(string, *fields) bool {
				return func(prefix string, f *fields) bool {
					result[prefix] = f
					return false
				}
			},
			expected: map[string]*fields{
				"":              referenceTestExample,
				"single_child":  referenceTestExample.oneToOnes["single_child"],
				"null_child":    referenceTestExample.oneToOnes["null_child"],
				"many_children": referenceTestExample.oneToManys["many_children"],
				"null_children": referenceTestExample.oneToManys["null_children"],
			},
		},
		{
			name: "Crawl All Fields With Early Exit",
			fn: func(result map[string]*fields) func(string, *fields) bool {
				return func(prefix string, f *fields) bool {

					result[prefix] = f

					// early exit
					if prefix == "" {
						return true
					}

					return false
				}
			},
			expected: map[string]*fields{
				"": referenceTestExample,
			},
		},
	}

	for _, test := range tests {

		msg := fmt.Sprintf("%s: failed", test.name)

		// result for reached fields
		result := map[string]*fields{}

		// execute sut
		referenceTestExample.crawlFields(test.fn(result))

		// assert that result and expected match by value
		assert.Equalf(t, test.expected, result, msg)

		// assert that result and expected match by reference
		for k, v := range test.expected {
			assert.Samef(t, v, result[k], msg)
		}
	}
}

func TestBuildReferenceName(t *testing.T) {

	tests := []struct {
		name        string
		inputPrefix string
		inputName   string
		expected    string
	}{
		{
			name:        "Build Reference Name With Prefix and Name",
			inputPrefix: "prefix",
			inputName:   "field_name",
			expected:    "prefix_field_name",
		},
		{
			name:        "Build Reference Name With Just Prefix",
			inputPrefix: "prefix",
			inputName:   "",
			expected:    "prefix",
		},
		{
			name:        "Build Reference Name With Just Name",
			inputPrefix: "",
			inputName:   "field_name",
			expected:    "field_name",
		},
		{
			name:        "Build Reference Name Without Input",
			inputPrefix: "",
			inputName:   "",
			expected:    "",
		},
	}

	for _, test := range tests {
		msg := fmt.Sprintf("%s: failed", test.name)
		assert.Equalf(t, test.expected, buildReferenceName(test.inputPrefix, test.inputName), msg)
	}
}
