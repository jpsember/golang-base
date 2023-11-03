package webserv

import (
	. "github.com/jpsember/golang-base/base"
)

type TextWidgetObj struct {
	BaseWidgetObj
	size        WidgetSize
	fixedHeight int
}

type TextWidget = *TextWidgetObj

func NewTextWidget(id string, size WidgetSize, fixedHeight int, clickListener ClickWidgetListener) TextWidget {
	t := &TextWidgetObj{
		size:        size,
		fixedHeight: fixedHeight,
	}
	t.InitBase(id)

	if clickListener != nil {
		t.SetLowListener(func(sess Session, widget Widget, value string, args WidgetArgs) (any, error) {
			clickListener(sess, widget, args)
			return nil, nil
		})
	}
	return t
}

var textSize = map[WidgetSize]string{
	SizeMicro:  `.4`,
	SizeTiny:   `.5`,
	SizeSmall:  `.7`,
	SizeMedium: ``,
	SizeLarge:  ` style='font-size:1.3em'`,
	SizeHuge:   ` style='font-size:1.8em'`,
}

func (w TextWidget) RenderTo(s Session, m MarkupBuilder) {
	pr := PrIf("TextWidget.RenderTo", false)
	var textContent string
	sc := w.StaticContent()
	if sc != nil {
		textContent = sc.(string)
		pr("staticContent:", sc)
	} else {
		pr("RenderTo, widget id:", w.Id(), "reading string value; state provider:", w.stateProvider)
		textContent = s.WidgetStringValue(w)
	}
	pr("text content:", QUO, textContent)

	h := NewHtmlString(textContent)

	effectiveId := s.PrependId(w.Id())

	m.TgOpen(`div id=`).A(QUO, effectiveId)

	if w.LowListen != nil {
		m.A(` onclick="jsButton('`, effectiveId, `')"`)
	}
	if w.size != SizeDefault && w.size != SizeMedium {
		m.Style(`font-size:`, textSize[w.size], `em;`)
	}

	if w.fixedHeight != 0 {
		m.Style(`height:`, w.fixedHeight, `em;`)
		if Alert("!adding background color") {
			m.Style(`background-color:#fcc;`)
		}
	}

	m.TgContent()
	{
		for _, c := range h.Paragraphs() {
			m.A(`<p>`, c, `</p>`).Cr()
		}

	}
	m.TgClose()
}
