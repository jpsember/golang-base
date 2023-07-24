package webserv

import (
	. "github.com/jpsember/golang-base/base"
	"html"
)

type TextWidgetObj struct {
	BaseWidgetObj
	Text string
}

type TextWidget = *TextWidgetObj

func NewTextWidget() TextWidget {
	return &TextWidgetObj{}
}

func (w TextWidget) RenderTo(m MarkupBuilder, state JSMap) {
	if !w.Visible() {
		m.RenderInvisible(w, "p")
		return
	}
	m.A(`<pr id='`)
	m.A(w.Id)
	m.A(`'>`)
	Todo("Perhaps assume text is already escaped?")
	Todo("Perhaps treat linefeeds as distinct paragraphs?")
	
	textContent := state.OptString(w.Id, "No text found")
	m.A(html.EscapeString(textContent))
	m.A(`</pr>`)
	m.Cr()
}
