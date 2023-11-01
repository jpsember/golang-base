package webserv

import (
	. "github.com/jpsember/golang-base/base"
)

type ButtonWidgetListener func(sess *SessionStruct, widget Widget, args WidgetArgs)

type ButtonWidgetObj struct {
	BaseWidgetObj
	Label    HtmlString
	listener ButtonWidgetListener
}

type ButtonWidget = *ButtonWidgetObj

func NewButtonWidget(id string, listener ButtonWidgetListener) ButtonWidget {
	Todo("Why do we need a separate listener field, in addition to the LowListener?")
	if listener == nil {
		listener = doNothingButtonListener
	}
	b := &ButtonWidgetObj{
		listener: listener,
	}
	b.InitBase(id)
	b.SetLowListener(b.buttonListenWrapper)
	return b
}

func (b ButtonWidget) buttonListenWrapper(sess Session, widget Widget, value string, args WidgetArgs) (any, error) {
	b.listener(sess, widget, args)
	return nil, nil
}

func doNothingButtonListener(sess Session, widget Widget, args WidgetArgs) {
	Alert("<1#50Button has no listener yet:", widget.Id(), "args:", args)
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
