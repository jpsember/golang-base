package webserv

import (
	. "github.com/jpsember/golang-base/base"
)

type TextWidgetObj struct {
	BaseWidgetObj
	Text HtmlString // if not nil, the static content of the widget
}

type TextWidget = *TextWidgetObj

func NewTextWidget(id string) TextWidget {
	t := &TextWidgetObj{}
	t.Id = id
	return t
}

func (w TextWidget) RenderTo(m MarkupBuilder, state JSMap) {
	if !w.Visible() {
		m.RenderInvisible(w)
		return
	}

	textContent := w.Text
	m.A(`<div`)
	if textContent == nil {
		m.A(` id='`)
		m.A(w.Id)
		m.A(`'`)
		s := state.OptString(w.Id, "No text found")
		textContent = NewHtmlString(s)
	}
	m.A(`>`)

	for _, c := range textContent.Paragraphs() {
		m.A(`<p>`)
		m.A(c)
		m.A(`</p>`)
		m.Cr()
	}

	m.A(`</div>`)

	m.Cr()
}
