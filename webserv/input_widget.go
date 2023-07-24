package webserv

import (
	. "github.com/jpsember/golang-base/base"
)

// A Widget that displays editable text
type InputWidgetObj struct {
	BaseWidgetObj
}

type InputWidget = *InputWidgetObj

func NewInputWidget(id string) InputWidget {
	w := InputWidgetObj{
		BaseWidgetObj{
			Id: id,
		},
	}
	return &w
}

func (w InputWidget) RenderTo(m MarkupBuilder, state JSMap) {
	if !w.Visible() {
		m.RenderInvisible(w, "span")
		return
	}
	value := WidgetStringValue(state, w.Id)
	m.A(`<input type='text' id='`)
	m.A(w.Id)
	m.A(`' value='`)
	m.A(EscapedHtml(value).String())
	m.A(`' onchange='jsVal("`)
	m.A(w.Id)
	m.A(`")'>`)
	m.Cr()
}
