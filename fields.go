package goscanql

import "fmt"

func newFields() *fields {
	return &fields{
		orderedFieldNames: make([]string, 0),
		references:        make(map[string]interface{}),
		children:          make(map[string]*fields),
	}
}

type fields struct {
	orderedFieldNames []string
	references        map[string]interface{}
	children          map[string]*fields
}

func (f *fields) addChild(name string) *fields {
	if _, ok := f.children[name]; ok {
		panic(fmt.Errorf("child with same name (\"%s\") already exists", name))
	}

	f.children[name] = newFields()
	return f.children[name]
}

func (f *fields) addField(name string, value interface{}) {

	// assert that field hasn't already been added
	if _, ok := f.references[name]; ok {
		panic(fmt.Errorf("field with name \"%s\" already added", name))
	}

	// add field to this instance
	f.orderedFieldNames = append(f.orderedFieldNames, name)
	f.references[name] = value
}

func (f *fields) getFieldReferences() map[string]interface{} {

	m := make(map[string]interface{})

	f.crawlReferences(func(key string, value interface{}) {
		m[key] = value
	})

	return m
}

func (f *fields) getFieldByteReferences() map[string]*[]byte {

	m := make(map[string]*[]byte)

	f.crawlReferences(func(key string, value interface{}) {
		m[key] = &[]byte{}
	})

	return m
}

func (f *fields) crawlReferences(fn func(key string, value interface{})) {
	f.crawlReferencesWithPrefix("", fn)
}

func (f *fields) crawlReferencesWithPrefix(prefix string, fn func(key string, value interface{})) {

	// if there is a prefix, format it accordingly
	if prefix != "" {
		prefix = fmt.Sprintf("%s_", prefix)
	}

	// for each field, run callback (fn)
	for name, reference := range f.references {
		fn(fmt.Sprintf("%s%s", prefix, name), reference)
	}

	// crawl through children and repeat
	for name, child := range f.children {
		childPrefix := fmt.Sprintf("%s%s", prefix, name)
		child.crawlReferencesWithPrefix(childPrefix, fn)
	}
}
