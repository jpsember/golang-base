package webserv

import (
	. "github.com/jpsember/golang-base/base"
)

type TextWidgetObj struct {
	BaseWidgetObj
}

type TextWidget = *TextWidgetObj

func NewTextWidget(id string) TextWidget {
	t := &TextWidgetObj{}
	t.Id = id
	return t
}

func (w TextWidget) RenderTo(m MarkupBuilder, state JSMap) {
	b := w.GetBaseWidget()
	if !w.Visible() {
		m.RenderInvisible(w)
		return
	}

	var textContent string

	sc := b.StaticContent()
	hasStaticContent := sc != nil
	if hasStaticContent {
		textContent = sc.(string)
	} else {
		textContent = state.OptString(w.Id, "")
	}

	h := NewHtmlString(textContent)
	if hasStaticContent {
		m.A(`<div>`)
	} else {
		m.A(`<div id='`)
		m.A(w.Id)
		m.A(`'>`)
	}

	for _, c := range h.Paragraphs() {
		m.A(`<p>`)
		m.A(c)
		m.A(`</p>`)
		m.Cr()
	}
	m.A(`</div>`)

	m.Cr()
}
