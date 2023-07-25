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

	// While <input> are span tags, our widget should be considered a block element

	// The outermost element must have id "foo", since we will be replacing that id's outerhtml
	// to perform AJAX updates.
	//
	// The HTML input element has id "foo.aux"

	m.A(`<div id='`)
	m.A(w.Id)
	m.A(`'>`)
	m.DoIndent()

	m.DebugOpen(w)

	value := WidgetStringValue(state, w.Id)
	m.A(`<input type='text' id='`)
	m.A(w.Id)
	m.A(`.aux`)
	m.A(`' value='`)
	m.A(NewHtmlString(value).String())
	m.A(`' onchange='jsVal("`)
	m.A(w.Id)
	m.A(`")'>`)
	m.Cr()

	m.DebugClose()
}
