package webserv

import (
	. "github.com/jpsember/golang-base/base"
)

type ButtonWidgetObj struct {
	BaseWidgetObj
	Label HtmlString
	size  WidgetSize
}

type ButtonWidget = *ButtonWidgetObj

func NewButtonWidget(size WidgetSize) ButtonWidget {
	b := &ButtonWidgetObj{}
	b.size = size
	return b
}

func (w ButtonWidget) RenderTo(m MarkupBuilder, state JSMap) {
	if !w.Visible() {
		m.RenderInvisible(w)
		return
	}

	m.A(`<div id='`, w.Id, `'>`)

	m.DoIndent()
	m.A(`<button class='btn btn-primary `)
	if w.size != SizeDefault {
		m.A(MapValue(wsSize, w.size))
	}
	m.A(`'`)
	if !w.Enabled() {
		m.A(` disabled`)
	}
	m.A(` onclick='jsButton("`, w.Id, `")'>`)
	m.Escape(w.Label)
	m.A(`</button>`)
	m.Cr()

	m.DoOutdent()
	m.A(`</div>`)
	m.Cr()
}

var wsSize = BuildMap[WidgetSize, string](
	SizeLarge, "btn-lg", SizeSmall, "btn-sm")
