package webserv

import (
	. "github.com/jpsember/golang-base/base"
)

type ButtonWidgetObj struct {
	BaseWidgetObj
	Label    HtmlString
	listener ButtonWidgetListener
}

type ButtonWidget = *ButtonWidgetObj

func NewButtonWidget(listener ButtonWidgetListener) ButtonWidget {
	if listener == nil {
		listener = doNothingButtonListener
	}
	b := &ButtonWidgetObj{
		listener: listener,
	}
	b.BaseWidgetObj.LowListen = buttonListenWrapper
	return b
}

func buttonListenWrapper(sess Session, widget Widget, value string) (string, error) {
	b := widget.(ButtonWidget)
	err := b.listener(sess, widget)
	return "", err
}

func doNothingButtonListener(sess Session, widget Widget) error {
	Alert("<1#50Button has no listener yet:", widget.Id())
	return nil
}

type ButtonWidgetListener func(sess Session, widget Widget) error

func (w ButtonWidget) RenderTo(m MarkupBuilder, state JSMap) {
	if !w.Visible() {
		m.RenderInvisible(w)
		return
	}

	//tx := TextAlignStr(w.Align())
	if w.size == SizeTiny {
		// For now, interpreting SizeTiny to mean a non-underlined, link-styled button that is very small:
		m.A(`<div class='py-1' id='`, w.BaseId, `'>`)

		//m.A(>`)
		m.DoIndent()
		m.A(`<button class='btn btn-link text-decoration-none `)
		if w.Align() == AlignRight {
			m.A(`float-end `)
		}
		m.A(`' style='font-size: 0.6em'`)
	} else {

		// Adding py-3 here to put some vertical space between button and other widgets
		m.A(`<div class='py-3' id='`, w.BaseId, `'>`)

		m.DoIndent()

		m.A(`<button class='btn btn-primary `)
		if w.Align() == AlignRight {
			m.A(`float-end `)
		}

		if w.size != SizeDefault {
			m.A(MapValue(btnTextSize, w.size))
		}
		m.A(`'`)
	}

	if !w.Enabled() {
		m.A(` disabled`)
	}

	m.A(` onclick='jsButton("`, w.BaseId, `")'>`)
	m.Escape(w.Label)
	m.A(`</button>`)
	m.Cr()

	m.DoOutdent()
	m.A(`</div>`)
	m.Cr()
}

var btnTextSize = map[WidgetSize]string{
	SizeLarge: "btn-lg",
	SizeSmall: "btn-sm",
}
