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
	m.A(`<input type="text" id=`)
	m.Quoted(w.Id)
	m.A(` value=`)
	m.Quoted(EscapedHtml(value).String())
	m.A(` onchange=`)
	m.Quoted(`jsVal('` + w.Id + `')`)
	m.A(`>`)
	m.Cr()
}

func (w InputWidget) ReceiveValue(sess Session, value string) {
	if Alert("Modifying value") {
		value += "<<<---modified"
	}
	sess.State.Put(w.Id, value)
	// Request a repaint of the widget
	sess.Repaint(w.Id)
}
