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

	m.A(`<div id='`, w.BaseId, `' `)

	if w.size != SizeDefault && w.size != SizeMedium {
		m.StyleOn().A(`font-size:`, textSize[w.size], `em;`).StyleOff()
	}

	Todo("You can't repeat style tags; only the first will be kept")
	if w.fixedHeight != 0 {
		m.StyleOn().A(`height:`, w.fixedHeight, `em;`).StyleOff()
		m.StyleOn().A(`background-color:#fcc;`).StyleOff()
	}

	m.A(`>`).DoIndent()
	{
		for _, c := range h.Paragraphs() {
			m.A(`<p>`, c, `</p>`).Cr()
		}

	}
	m.DoOutdent()
	m.A(`</div>`).Cr()
}
