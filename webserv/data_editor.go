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

func NewDataEditor(data DataClass) DataEditor {
	j := data.ToJson().AsJSMap()
	t := &DataEditorStruct{
		parser:              data,
		JSMap:               j,
		WidgetStateProvider: NewStateProvider("", j),
	}
	return t
}

func (d DataEditor) Read() DataClass {
	result := d.parser.Parse(d)
	return result
}
