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
		existingOneChildren  map[string]*fields
		existingManyChildren map[string]*fields
		expectedRelationship relationship
		expectedErr          error
	}{
		{
			name:                 "Add New One-to-One Child Struct",
			inputName:            "child",
			inputObj:             &struct{}{},
			expectedRelationship: oneRelationship,
			expectedErr:          nil,
		},
		{
			name:                 "Add New One-to-Many Child Struct",
			inputName:            "child",
			inputObj:             &[]*struct{}{},
			expectedRelationship: manyRelationship,
			expectedErr:          nil,
		},
		{
			name:      "Add New One-to-One Child Struct WIth Name Collision",
			inputName: "arbitrary_name",
			inputObj:  &struct{}{},
			existingOneChildren: map[string]*fields{
				"arbitrary_name": nil,
			},
			expectedRelationship: oneRelationship,
			expectedErr:          fmt.Errorf("child already exists with name \"%s\"", "arbitrary_name"),
		},
		{
			name:      "Add New One-to-Many Child Struct WIth Name Collision",
			inputName: "arbitrary_name_many",
			inputObj:  &struct{}{},
			existingOneChildren: map[string]*fields{
				"arbitrary_name_many": nil,
			},
			expectedRelationship: oneRelationship,
			expectedErr:          fmt.Errorf("child already exists with name \"%s\"", "arbitrary_name_many"),
		},
	}

	for _, test := range tests {

		msg := fmt.Sprintf("%s: failed", test.name)

		// initialise fields
		f := &fields{
			oneToOnes:  map[string]*fields{},
			oneToManys: map[string]*fields{},
		}

		// initialise children if test requires
		if test.existingOneChildren != nil {
			f.oneToOnes = test.existingOneChildren
		}

		if test.existingManyChildren != nil {
			f.oneToManys = test.existingManyChildren
		}

		// execute sut
		err := f.addNewChild(test.inputName, test.inputObj)

		// assert that test returned expected error
		assert.Equalf(t, test.expectedErr, err, msg)

		// if error returned, continue as remaining asserts are nullified
		if err != nil {
			continue
		}

		// assess result based on expected relationship
		if test.expectedRelationship == oneRelationship {

			// check that the new field has been added to one-to-one children
			assert.Containsf(t, f.oneToOnes, test.inputName, msg)
			// and that it hasn't been added to the one-to-manys children
			assert.NotContainsf(t, f.oneToManys, test.inputName, msg)

		} else {

			// check that the new field has been added to one-to-manys children
			assert.Containsf(t, f.oneToManys, test.inputName, msg)
			// and that it hasn't been added to the one-to-ones children
			assert.NotContainsf(t, f.oneToOnes, test.inputName, msg)

		}
	}
}
