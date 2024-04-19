package goscanql

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func referenceField[T any](field T) *T {
	return &field
}

type exampleScanner struct {
	id string
}

func (e exampleScanner) Scan(_ interface{}) error {
	return nil
}

func (e exampleScanner) ID() []byte {
	return []byte(e.id)
}

func TestInitialiseFields(t *testing.T) {
	type childExample struct {
		Foo int    `goscanql:"foo"`
		Bar string `goscanql:"bar"`
	}

	objExample := &struct {
		ID                  int    `goscanql:"id"`
		Name                string `goscanql:"name"`
		UnnamedField        string
		TimeExample         time.Time        `goscanql:"time"`
		Scanner             exampleScanner   `goscanql:"scanner"`
		ScannerPointer      *exampleScanner  `goscanql:"scanner_pointer"`
		Child               childExample     `goscanql:"child"`
		ChildPointer        *childExample    `goscanql:"child_pointer"`
		ChildPointerPointer **childExample   `goscanql:"child_pointer_pointer"`
		Children            []childExample   `goscanql:"children"`
		ChildrenPointer     *[]childExample  `goscanql:"children_pointer"`
		ChildrenScanners    []exampleScanner `goscanql:"children_scanners"`
	}{}

	subject := &fields{
		obj:                  objExample,
		orderedFieldNames:    []string{},
		orderedScannerNames:  []string{},
		orderedOneToOneNames: []string{},
		references:           map[string]interface{}{},
		scannerReferences:    map[string]Scanner{},
		nullFields:           map[string]*nullBytes{},
		oneToOnes:            map[string]*fields{},
		oneToManys:           map[string]*fields{},
	}

	newExpectedChildExampleFields := func(obj interface{}) *fields {
		f := &fields{
			obj: obj,
			orderedFieldNames: []string{
				"foo",
				"bar",
			},
			orderedScannerNames:  []string{},
			orderedOneToOneNames: []string{},
			references: map[string]interface{}{
				"foo": referenceField(0),
				"bar": referenceField(""),
			},
			scannerReferences: map[string]Scanner{},
			nullFields: map[string]*nullBytes{
				"foo": {isNil: true},
				"bar": {isNil: true},
			},
			oneToOnes:  map[string]*fields{},
			oneToManys: map[string]*fields{},
		}

		return f
	}

	expected := &fields{
		obj: objExample,
		orderedFieldNames: []string{
			"id",
			"name",
			"time",
		},
		orderedScannerNames: []string{
			"scanner",
			"scanner_pointer",
		},
		orderedOneToOneNames: []string{
			"child",
			"child_pointer",
			"child_pointer_pointer",
		},
		references: map[string]interface{}{
			"id":   referenceField(0),
			"name": referenceField(""),
			"time": referenceField(time.Time{}),
		},
		scannerReferences: map[string]Scanner{
			"scanner":         &objExample.Scanner,
			"scanner_pointer": referenceField(exampleScanner{}),
		},
		nullFields: map[string]*nullBytes{
			"id":              {isNil: true},
			"name":            {isNil: true},
			"time":            {isNil: true},
			"scanner":         {isNil: true},
			"scanner_pointer": {isNil: true},
		},
		oneToOnes: map[string]*fields{
			"child":                 newExpectedChildExampleFields(&childExample{}),
			"child_pointer":         newExpectedChildExampleFields(referenceField(&childExample{})),
			"child_pointer_pointer": newExpectedChildExampleFields(referenceField(referenceField(&childExample{}))),
		},
		oneToManys: map[string]*fields{
			"children":         newExpectedChildExampleFields(&childExample{}),
			"children_pointer": newExpectedChildExampleFields(&childExample{}),
			"children_scanners": {
				obj:               &exampleScanner{},
				orderedFieldNames: []string{},
				orderedScannerNames: []string{
					"",
				},
				orderedOneToOneNames: []string{},
				references:           map[string]interface{}{},
				scannerReferences: map[string]Scanner{
					"": &exampleScanner{},
				},
				nullFields: map[string]*nullBytes{
					"": {isNil: true},
				},
				oneToOnes:  map[string]*fields{},
				oneToManys: map[string]*fields{},
			},
		},
	}

	msg := "Initialised Fields Test: failed"

	// execute sut
	err := subject.initialise("")

	// assert that the test doesn't return an error
	assert.Equalf(t, nil, err, msg)

	// assert that the general structure produced is as expected, this assertion does not inspect
	// the memory addresses of pointers (only the underlying values)
	assert.Equalf(t, expected, subject, "")

	// assert that all the pointers refer to fields of the original object
	assert.Samef(t, &objExample.ID, subject.references["id"], msg)
	assert.Samef(t, &objExample.Name, subject.references["name"], msg)
	assert.Samef(t, &objExample.TimeExample, subject.references["time"], msg)
	assert.Samef(t, &objExample.Scanner, subject.scannerReferences["scanner"], msg)
	assert.Samef(t, objExample.ScannerPointer, subject.scannerReferences["scanner_pointer"], msg)
	assert.Samef(t, &objExample.Child.Foo, subject.oneToOnes["child"].references["foo"], msg)
	assert.Samef(t, &objExample.Child.Bar, subject.oneToOnes["child"].references["bar"], msg)
	assert.Samef(t, &objExample.ChildPointer.Foo, subject.oneToOnes["child_pointer"].references["foo"], msg)
	assert.Samef(t, &objExample.ChildPointer.Bar, subject.oneToOnes["child_pointer"].references["bar"], msg)
	assert.Samef(t, &(*objExample.ChildPointerPointer).Foo, subject.oneToOnes["child_pointer_pointer"].references["foo"], msg)
	assert.Samef(t, &(*objExample.ChildPointerPointer).Bar, subject.oneToOnes["child_pointer_pointer"].references["bar"], msg)
	assert.Samef(t, &objExample.Children[0].Foo, subject.oneToManys["children"].references["foo"], msg)
	assert.Samef(t, &objExample.Children[0].Bar, subject.oneToManys["children"].references["bar"], msg)
	assert.Samef(t, &(*objExample.ChildrenPointer)[0].Foo, subject.oneToManys["children_pointer"].references["foo"], msg)
	assert.Samef(t, &(*objExample.ChildrenPointer)[0].Bar, subject.oneToManys["children_pointer"].references["bar"], msg)
	assert.Samef(t, &objExample.ChildrenScanners[0], subject.oneToManys["children_scanners"].scannerReferences[""], msg)
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
				orderedScannerNames:  []string{},
				orderedOneToOneNames: []string{},
				references: map[string]interface{}{
					"foo": referenceField(0),
					"bar": referenceField(""),
				},
				scannerReferences: map[string]Scanner{},
				nullFields: map[string]*nullBytes{
					"foo": {isNil: true},
					"bar": {isNil: true},
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
				orderedFieldNames: []string{
					"foo",
					"bar",
				},
				orderedScannerNames:  []string{},
				orderedOneToOneNames: []string{},
				references: map[string]interface{}{
					"foo": referenceField(0),
					"bar": referenceField(""),
				},
				scannerReferences: map[string]Scanner{},
				nullFields: map[string]*nullBytes{
					"foo": {isNil: true},
					"bar": {isNil: true},
				},
				oneToOnes:  map[string]*fields{},
				oneToManys: map[string]*fields{},
			},
			expectedErr: nil,
		},
	}

	for _, test := range tests {
		msg := fmt.Sprintf("%s: failed", test.name)

		// execute sut
		result, err := newFields(testInputs[test.name])

		// assert value equality between expected and result
		assert.Equalf(t, test.expected, result, msg)
		assert.Equalf(t, test.expectedErr, err, msg)

		// if test errored, continue to next as following assertions are nullified
		if err != nil {
			continue
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
				nullFields:        map[string]*nullBytes{},
			},
			expected: &fields{
				orderedFieldNames: []string{
					"field_name",
				},
				references: map[string]interface{}{
					"field_name": referenceField(0),
				},
				nullFields: map[string]*nullBytes{
					"field_name": {isNil: true},
				},
			},
			expectedErr: nil,
		},
		{
			name:      "Add Single Field With Field Collision",
			inputName: "field_name",
			inputObj:  referenceField(0),
			fields: &fields{
				orderedFieldNames: []string{
					"field_name",
				},
				references: map[string]interface{}{
					"field_name": referenceField(0),
				},
				nullFields: map[string]*nullBytes{
					"field_name": {isNil: true},
				},
			},
			expected:    nil, // N/A for this test
			expectedErr: fmt.Errorf("field with name \"field_name\" already added"),
		},
		{
			name:      "Add Single Field With Scanner Collision",
			inputName: "field_name",
			inputObj:  referenceField(0),
			fields: &fields{
				orderedFieldNames: []string{},
				orderedScannerNames: []string{
					"field_name",
				},
				scannerReferences: map[string]Scanner{
					"field_name": nil,
				},
				nullFields: map[string]*nullBytes{
					"field_name": {isNil: true},
				},
			},
			expected:    nil, // N/A for this test
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

func TestAddScanner(t *testing.T) {
	tests := []struct {
		name         string
		inputName    string
		inputScanner Scanner
		fields       *fields
		expected     *fields
		expectedErr  error
	}{
		{
			name:         "Add Single Scanner Without Collision",
			inputName:    "field_name",
			inputScanner: &exampleScanner{},
			fields: &fields{
				orderedScannerNames: []string{},
				scannerReferences:   map[string]Scanner{},
				nullFields:          map[string]*nullBytes{},
			},
			expected: &fields{
				orderedScannerNames: []string{
					"field_name",
				},
				scannerReferences: map[string]Scanner{
					"field_name": &exampleScanner{},
				},
				nullFields: map[string]*nullBytes{
					"field_name": {isNil: true},
				},
			},
			expectedErr: nil,
		},
		{
			name:         "Add Single Scanner With Scanner Collision",
			inputName:    "field_name",
			inputScanner: &exampleScanner{},
			fields: &fields{
				orderedScannerNames: []string{
					"field_name",
				},
				scannerReferences: map[string]Scanner{
					"field_name": &exampleScanner{},
				},
				nullFields: map[string]*nullBytes{
					"field_name": {isNil: true},
				},
			},
			expected:    nil, // N/A for this test
			expectedErr: fmt.Errorf("field with name \"field_name\" already added"),
		},
		{
			name:         "Add Single Scanner With Field Collision",
			inputName:    "field_name",
			inputScanner: &exampleScanner{},
			fields: &fields{
				orderedFieldNames: []string{
					"field_name",
				},
				orderedScannerNames: []string{},
				references: map[string]interface{}{
					"field_name": nil,
				},
				nullFields: map[string]*nullBytes{
					"field_name": {isNil: true},
				},
			},
			expected:    nil, // N/A for this test
			expectedErr: fmt.Errorf("field with name \"field_name\" already added"),
		},
	}

	for _, test := range tests {
		err := test.fields.addScanner(test.inputName, test.inputScanner)

		// assert that error is expected
		assert.Equalf(t, test.expectedErr, err, "")

		// continue to next if nil as following asserts are nullified
		if err != nil {
			continue
		}

		// assert that the resulting struct is the same (value-wise) as the expected fields
		assert.Equalf(t, test.expected, test.fields, "")

		// assert that the added field points to the exact same object as originally provided
		assert.Samef(t, test.inputScanner, test.fields.scannerReferences[test.inputName], "")
	}
}

var (
	referenceTestExample = &fields{
		orderedFieldNames: []string{
			"foo",
			"bar",
		},
		orderedScannerNames: []string{
			"scanner",
			"another_scanner",
		},
		orderedOneToOneNames: []string{
			"single_child",
			"null_child",
		},
		references: map[string]interface{}{
			"foo": referenceField(36),
			"bar": referenceField("Hello, World!"),
		},
		scannerReferences: map[string]Scanner{
			"scanner":         &exampleScanner{},
			"another_scanner": &exampleScanner{},
		},
		nullFields: map[string]*nullBytes{
			"foo":             {isNil: false},
			"bar":             {isNil: false},
			"scanner":         {isNil: false},
			"another_scanner": {isNil: false},
		},
		oneToOnes: map[string]*fields{
			"single_child": {
				orderedFieldNames: []string{
					"time",
				},
				references: map[string]interface{}{
					"time": referenceField(time.Time{}),
				},
				nullFields: map[string]*nullBytes{
					"time": {isNil: false},
				},
			},
			"null_child": {
				orderedFieldNames: []string{
					"time",
				},
				references: map[string]interface{}{
					"time": referenceField(time.Time{}),
				},
				nullFields: map[string]*nullBytes{
					"time": {isNil: true}, // as all nullFields for this *fields are nil, it is considered nil
				},
			},
		},
		oneToManys: map[string]*fields{
			"many_children": {
				orderedFieldNames: []string{
					"many_foo",
					"many_bar",
				},
				references: map[string]interface{}{
					"many_foo": referenceField(72),
					"many_bar": referenceField("Hello, worlds!"),
				},
				nullFields: map[string]*nullBytes{
					"many_foo": {isNil: false},
					"many_bar": {isNil: false},
				},
			},
			"null_children": {
				orderedFieldNames: []string{
					"foo",
				},
				references: map[string]interface{}{
					"foo": referenceField(0),
				},
				nullFields: map[string]*nullBytes{
					"foo": {isNil: true}, // as all nullFields for this *fields are nil, it is considered nil
				},
			},
		},
	}
)

func TestGetFieldReferences(t *testing.T) {
	expected := map[string]interface{}{
		"foo":                    referenceTestExample.references["foo"],
		"bar":                    referenceTestExample.references["bar"],
		"scanner":                referenceTestExample.scannerReferences["scanner"],
		"another_scanner":        referenceTestExample.scannerReferences["another_scanner"],
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
	expected := map[string]*nullBytes{
		"foo":                    referenceTestExample.nullFields["foo"],
		"bar":                    referenceTestExample.nullFields["bar"],
		"scanner":                referenceTestExample.nullFields["scanner"],
		"another_scanner":        referenceTestExample.nullFields["another_scanner"],
		"single_child_time":      referenceTestExample.oneToOnes["single_child"].nullFields["time"],
		"null_child_time":        referenceTestExample.oneToOnes["null_child"].nullFields["time"],
		"many_children_many_foo": referenceTestExample.oneToManys["many_children"].nullFields["many_foo"],
		"many_children_many_bar": referenceTestExample.oneToManys["many_children"].nullFields["many_bar"],
		"null_children_foo":      referenceTestExample.oneToManys["null_children"].nullFields["foo"],
	}

	msg := "Get Byte References: failed"

	result := referenceTestExample.getNullFieldReferences()

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

					// early exit if empty string by returning true, else continue
					return prefix == ""
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

func TestGetBytePrint(t *testing.T) {
	expectedBytePrint := []byte(`{foo:36}{bar:"Hello, World!"}{scanner:}{another_scanner:}{single_child_time:time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC)}{null_child_time:time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC)}`)
	assert.Equalf(t, expectedBytePrint, referenceTestExample.getBytePrint(""), "Get Byte Print Test: failed")
}

func TestGetHash(t *testing.T) {
	expectedHash := []byte{87, 52, 237, 215, 223, 186, 202, 75, 79, 182, 214, 206, 45, 250, 62, 135, 31, 127, 190, 176}
	assert.Equalf(t, string(expectedHash), referenceTestExample.getHash(), "Get Hash Test: failed")
}

func TestIsNil(t *testing.T) {
	tests := []struct {
		name     string
		fields   *fields
		expected bool
	}{
		{
			name: "IsNil All Nil Fields",
			fields: &fields{
				nullFields: map[string]*nullBytes{
					"foo": {isNil: true},
					"bar": {isNil: true},
				},
			},
			expected: true,
		},
		{
			name: "IsNil Some Nil Fields",
			fields: &fields{
				nullFields: map[string]*nullBytes{
					"foo": {isNil: true},
					"bar": {isNil: true},
					"a":   {isNil: false},
				},
			},
			expected: false,
		},
		{
			name: "IsNil No Nil Fields",
			fields: &fields{
				nullFields: map[string]*nullBytes{
					"foo": {isNil: false},
					"bar": {isNil: false},
				},
			},
			expected: false,
		},
		{
			name: "IsNil No Nil Fields But Nil Child",
			fields: &fields{
				nullFields: map[string]*nullBytes{
					"foo": {isNil: false},
					"bar": {isNil: false},
				},
				oneToManys: map[string]*fields{
					"": {
						nullFields: map[string]*nullBytes{
							"foo": {isNil: true},
							"bar": {isNil: true},
						},
					},
				},
			},
			expected: false,
		},
	}

	for _, test := range tests {
		msg := fmt.Sprintf("%s: failed", test.name)
		assert.Equalf(t, test.expected, test.fields.isNil(), msg)
	}
}

func TestIsMatch(t *testing.T) {
	tests := []struct {
		name     string
		fields   *fields
		comparee *fields
		expected bool
	}{
		{
			name: "IsMatch Equal Fields",
			fields: &fields{
				orderedFieldNames: []string{
					"foo",
				},
				references: map[string]interface{}{
					"foo": referenceField("hello!"),
				},
			},
			comparee: &fields{
				orderedFieldNames: []string{
					"foo",
				},
				references: map[string]interface{}{
					"foo": referenceField("hello!"),
				},
			},
			expected: true,
		},
		{
			name: "IsMatch Equal Fields and One-to-One Children and Scanner",
			fields: &fields{
				orderedFieldNames: []string{
					"bar",
				},
				orderedScannerNames: []string{
					"scanner",
				},
				references: map[string]interface{}{
					"bar": referenceField(63),
				},
				scannerReferences: map[string]Scanner{
					"scanner": &exampleScanner{
						id: "123456789",
					},
				},
				orderedOneToOneNames: []string{
					"foobar",
				},
				oneToOnes: map[string]*fields{
					"foobar": {
						orderedFieldNames: []string{
							"foo",
						},
						references: map[string]interface{}{
							"foo": &[]byte{1, 2, 3},
						},
					},
				},
			},
			comparee: &fields{
				orderedFieldNames: []string{
					"bar",
				},
				orderedScannerNames: []string{
					"scanner",
				},
				references: map[string]interface{}{
					"bar": referenceField(63),
				},
				scannerReferences: map[string]Scanner{
					"scanner": &exampleScanner{
						id: "123456789",
					},
				},
				orderedOneToOneNames: []string{
					"foobar",
				},
				oneToOnes: map[string]*fields{
					"foobar": {
						orderedFieldNames: []string{
							"foo",
						},
						references: map[string]interface{}{
							"foo": &[]byte{1, 2, 3},
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "IsMatch Not Equal Fields",
			fields: &fields{
				orderedFieldNames: []string{
					"foo",
				},
				references: map[string]interface{}{
					"foo": referenceField("hello!"),
				},
			},
			comparee: &fields{
				orderedFieldNames: []string{
					"foo",
				},
				references: map[string]interface{}{
					"foo": referenceField("hello!!"),
				},
			},
			expected: false,
		},
		{
			name: "IsMatch Not Equal Fields and One-to-One Children",
			fields: &fields{
				orderedFieldNames: []string{
					"bar",
				},
				references: map[string]interface{}{
					"bar": referenceField(63),
				},
				orderedOneToOneNames: []string{
					"foobar",
				},
				oneToOnes: map[string]*fields{
					"foobar": {
						orderedFieldNames: []string{
							"foo",
						},
						references: map[string]interface{}{
							"foo": &[]byte{1, 2, 3},
						},
					},
				},
			},
			comparee: &fields{
				orderedFieldNames: []string{
					"bar",
				},
				references: map[string]interface{}{
					"bar": referenceField(63),
				},
				orderedOneToOneNames: []string{
					"foobar",
				},
				oneToOnes: map[string]*fields{
					"foobar": {
						orderedFieldNames: []string{
							"foo",
						},
						references: map[string]interface{}{
							"foo": &[]byte{1, 2, 4},
						},
					},
				},
			},
			expected: false,
		},
		{
			name: "IsMatch Not Equal Scanners",
			fields: &fields{
				orderedFieldNames: []string{
					"bar",
				},
				orderedScannerNames: []string{
					"scanner",
				},
				references: map[string]interface{}{
					"bar": referenceField(63),
				},
				scannerReferences: map[string]Scanner{
					"scanner": &exampleScanner{
						id: "123456789",
					},
				},
				orderedOneToOneNames: []string{
					"foobar",
				},
				oneToOnes: map[string]*fields{
					"foobar": {
						orderedFieldNames: []string{
							"foo",
						},
						references: map[string]interface{}{
							"foo": &[]byte{1, 2, 3},
						},
					},
				},
			},
			comparee: &fields{
				orderedFieldNames: []string{
					"bar",
				},
				orderedScannerNames: []string{
					"scanner",
				},
				references: map[string]interface{}{
					"bar": referenceField(63),
				},
				orderedOneToOneNames: []string{
					"foobar",
				},
				scannerReferences: map[string]Scanner{
					"scanner": &exampleScanner{
						id: "987654321",
					},
				},
				oneToOnes: map[string]*fields{
					"foobar": {
						orderedFieldNames: []string{
							"foo",
						},
						references: map[string]interface{}{
							"foo": &[]byte{1, 2, 3},
						},
					},
				},
			},
			expected: false,
		},
	}

	for _, test := range tests {
		assert.Equalf(t, test.expected, test.fields.isMatch(test.comparee), "")
	}
}
