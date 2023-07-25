package webserv

import (
	. "github.com/jpsember/golang-base/base"
)

type TextWidgetObj struct {
	BaseWidgetObj
	Text string // if nonempty, the static content of the widget
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

	textContent := w.Text
	m.A(`<div`)
	dynamic := textContent == ""
	if dynamic {
		m.A(` id='`)
		m.A(w.Id)
		m.A(`'`)
		textContent = state.OptString(w.Id, "No text found")
	}
	m.A(`>`)

	m.DoIndent()
	m.DebugOpen(w)

	Todo("make w.Text field an HtmlString")
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
