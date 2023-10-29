package webserv

import (
	. "github.com/jpsember/golang-base/base"
	"strings"
)

type DataEditorStruct struct {
	JSMap         // embedded JSMap so we can modify object fields directly (e.g. editor.PutInt(...))
	StateProvider WidgetStateProvider
	parser        DataClass
}

type DataEditor = *DataEditorStruct

func NewDataEditor(data DataClass) DataEditor {
	return NewDataEditorWithPrefix(data, "")
}

func NewDataEditorWithPrefix(data DataClass, prefix string) DataEditor {
	Todo("!Make StateProvider an embedded struct")
	if prefix != "" {
		CheckArg(strings.HasSuffix(prefix, ":"), "expected prefix to end with ':'")
	}
	Todo("!Maybe the prefix, if nonempty, should omit the ':' which is added here?")
	dataAsJson := data.ToJson().AsJSMap()
	t := &DataEditorStruct{
		parser:        data,
		JSMap:         dataAsJson,
		StateProvider: NewStateProvider(prefix, dataAsJson),
	}
	return t
}

func (d DataEditor) Read() DataClass {
	result := d.parser.Parse(d.StateProvider.State)
	return result
}
