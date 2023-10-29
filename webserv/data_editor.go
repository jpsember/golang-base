package webserv

import (
	. "github.com/jpsember/golang-base/base"
)

type DataEditorStruct struct {
	JSMap
	StateProvider WidgetStateProvider
	parser        DataClass
}

type DataEditor = *DataEditorStruct

func NewDataEditor(data DataClass) DataEditor {
	return NewDataEditorWithPrefix(data, "")
}

func NewDataEditorWithPrefix(data DataClass, prefix string) DataEditor {
	Todo("!Make StateProvider an embedded struct")
	Todo("!Editor doesn't need an explicit JSMap, instead it can use the embedded WidgetStateProvider's")
	j := data.ToJson().AsJSMap()
	t := &DataEditorStruct{
		parser:        data,
		JSMap:         j,
		StateProvider: NewStateProvider(prefix, j),
	}
	return t
}

func (d DataEditor) Read() DataClass {
	result := d.parser.Parse(d)
	return result
}
