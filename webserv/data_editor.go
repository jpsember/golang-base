package webserv

import (
	. "github.com/jpsember/golang-base/base"
)

type DataEditorStruct struct {
	BaseObject
	parser DataClass
	state  JSMap
}

type DataEditor = *DataEditorStruct

func NewDataEditor(data DataClass) DataEditor {
	t := &DataEditorStruct{
		parser: data,
		state:  data.ToJson().AsJSMap(),
	}
	t.ProvideName("DataEditor for " + TypeOf(data))
	t.AlertVerbose()
	t.Log("constructed state:", INDENT, t.state)
	return t
}

func (d DataEditor) State() JSMap {
	return d.state
}

func (d DataEditor) Read() DataClass {
	result := d.parser.Parse(d.state)
	d.Log("Read():", INDENT, result)
	return result
}

func (d DataEditor) StateProvider() WidgetStateProvider {
	return NewStateProvider("", d.state)
}
