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

func NewTextWidget(id string, size WidgetSize, fixedHeight int) TextWidget {
	t := &TextWidgetObj{
		size:        size,
		fixedHeight: fixedHeight,
	}
	t.InitBase(id)
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
	var textContent string
	if w.staticContent != nil {
		textContent = w.staticContent.(string)
	} else {
		textContent = s.WidgetStringValue(w)
	}

	h := NewHtmlString(textContent)

	m.TgOpen(`div id=`)
	m.A(QUOTED, w.BaseId)

	if w.size != SizeDefault && w.size != SizeMedium {
		m.Style(`font-size:`, textSize[w.size], `em;`)
	}

	if w.fixedHeight != 0 {
		m.Style(`height:`, w.fixedHeight, `em;`)
		if Alert("adding background color") {
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
