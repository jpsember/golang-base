package webserv

import (
	. "github.com/jpsember/golang-base/base"
)

type DataEditorStruct struct {
	JSMap
	WidgetStateProvider
	parser DataClass
}

type DataEditor = *DataEditorStruct

func NewDataEditorWithPrefix(data DataClass, prefix string) DataEditor {
	j := data.ToJson().AsJSMap()
	t := &DataEditorStruct{
		parser:              data,
		JSMap:               j,
		WidgetStateProvider: NewStateProvider(prefix, j),
	}
	return t
}

func NewDataEditor(data DataClass) DataEditor {
	return NewDataEditorWithPrefix(data, "")
}

func (d DataEditor) Read() DataClass {
	result := d.parser.Parse(d)
	return result
}
