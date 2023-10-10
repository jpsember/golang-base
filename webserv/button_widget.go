package webserv

import (
	. "github.com/jpsember/golang-base/base"
)

// This (and other widget listeners) can call Session.Context() to get the context, if any.  For
// list widgets, the context will return the element id.
type ButtonWidgetListener func(sess *SessionStruct, widget Widget, message string)

type ButtonWidgetObj struct {
	BaseWidgetObj
	Label    HtmlString
	listener ButtonWidgetListener
}

type ButtonWidget = *ButtonWidgetObj

func NewButtonWidget(id string, listener ButtonWidgetListener) ButtonWidget {
	if listener == nil {
		listener = doNothingButtonListener
	}
	b := &ButtonWidgetObj{
		listener: listener,
	}
	b.InitBase(id)
	b.LowListen = b.buttonListenWrapper
	return b
}

func (b ButtonWidget) buttonListenWrapper(sess Session, widget Widget, value string) (any, error) {
	b.listener(sess, widget, value)
	return nil, nil
}

func doNothingButtonListener(sess Session, widget Widget, message string) {
	Alert("<1#50Button has no listener yet:", widget.Id(), "message:", message)
}

func RenderButton(s Session, m MarkupBuilder, w_BaseId string, actionId string, enabled bool, w_Label any, w_size WidgetSize, w_align WidgetAlign, vertPadding int) {
	vertPaddingExpr := `py-` + IntToString(vertPadding)

	Todo("!Can probably get rid of vertPadding if we explicitly add spacing rows somehow")
	if w_size == SizeTiny {
		// For now, interpreting SizeTiny to mean a non-underlined, link-styled button that is very small:
		m.A(`<div class='`, vertPaddingExpr, `' id='`, w_BaseId, `'>`)

		//m.A(>`)
		m.DoIndent()
		m.A(`<button class='btn btn-link text-decoration-none `)
		if w_align == AlignRight {
			m.A(`float-end `)
		}
		m.A(`' style='font-size: 0.6em'`)
	} else {

		m.A(`<div class='`, vertPaddingExpr, `' id='`, w_BaseId, `'>`)

		m.DoIndent()

		m.A(`<button class='btn btn-primary `)
		if w_align == AlignRight {
			m.A(`float-end `)
		}

		if w_size != SizeDefault {
			m.A(MapValue(btnTextSize, w_size))
		}
		m.A(`'`)
	}

	if !enabled {
		m.A(` disabled`)
	}

	Todo("!Detect escaping of quotes (debug only)")
	Todo("!Prefer single quotes over doubles, as they don't produce &quot; when escaping for html/javascript")
	m.A(` onclick="jsButton('`, s.PrependId(actionId), `')"`, `>`)
	m.Escape(w_Label)
	m.A(`</button>`)
	m.Cr()

	m.DoOutdent()
	m.A(`</div>`)
	m.Cr()
}

func (w ButtonWidget) RenderTo(s Session, m MarkupBuilder) {
	RenderButton(s, m, w.Id(), w.Id(), w.Enabled(), w.Label, w.size, w.align, 1)
}

var btnTextSize = map[WidgetSize]string{
	SizeLarge: "btn-lg",
	SizeSmall: "btn-sm",
}
