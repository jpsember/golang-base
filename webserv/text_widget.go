package webserv

import (
	. "github.com/jpsember/golang-base/base"
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
		m.RenderInvisible(w, "div")
		return
	}

	m.A(`<div id='`)
	m.A(w.Id)
	m.A(`'>`)
	m.DoIndent()
	m.DebugOpen(w)

	textContent := state.OptString(w.Id, "No text found")
	paras := EscapedHtmlIntoParagraphs(textContent)
	for _, c := range paras {
		m.A(`<p>`)
		m.A(c.String())
		m.A(`</p>`)
		m.Cr()
	}

	m.DebugClose()
	m.DoOutdent()
	m.A(`</div>`)
	m.Cr()
}
