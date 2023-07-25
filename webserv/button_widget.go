package webserv

import (
	. "github.com/jpsember/golang-base/base"
)

type ButtonWidgetObj struct {
	BaseWidgetObj
	Label HtmlString
}

type ButtonWidget = *ButtonWidgetObj

func NewButtonWidget() ButtonWidget {
	return &ButtonWidgetObj{}
}

func (w ButtonWidget) RenderTo(m MarkupBuilder, state JSMap) {
	if !w.Visible() {
		m.RenderInvisible(w, "div")
		return
	}

	m.A(`<div id='`)
	m.A(w.Id)
	m.A(`'>`)

	m.DoIndent()
	m.DebugOpen(w)
	m.A(`<button class='btn btn-primary'`)
	if !w.Enabled() {
		m.A(` disabled`)
	}
	m.A(` onclick='jsButton("`)
	m.A(w.Id)
	m.A(`")'>`)
	m.A(w.Label.Escaped())
	m.A(`</button>`)
	m.Cr()

	m.DebugClose()
	m.DoOutdent()
	m.A(`</div>`)
	m.Cr()
}
