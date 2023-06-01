package webserv

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/json"
)

// A Widget that displays editable text
type InputWidgetObj struct {
	BaseWidgetObj
}

type InputWidget = *InputWidgetObj

func NewInputWidget(id string, size int) InputWidget {
	w := InputWidgetObj{
		BaseWidgetObj{
			Id: id,
		},
	}
	return &w
}

func (w InputWidget) RenderTo(m MarkupBuilder, state JSMap) {
	value := WidgetStringValue(state, w.Id)

	if value == "" {
		Alert("changing value to something unusual")
		value = `hello "friend"`
	}
	m.A(`<input type="text" id=`).Quoted(w.Id).A(` value=`).Quoted(EscapedHtml(value).String()).A(` onchange=`).Quoted(`jsVal('` + w.Id + `')`).A(`>`).Cr()
}
