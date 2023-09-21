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

func (w TextWidget) RenderTo(s Session, m MarkupBuilder) {
	var textContent string
	if w.staticContent != nil {
		textContent = w.staticContent.(string)
	} else {
		textContent = s.WidgetStringValue(w)
	}

	h := NewHtmlString(textContent)

	Pr("id:", w.BaseId, "w.fixedHeight:", w.fixedHeight)

	m.A(`<div id='`, w.BaseId, `' `)

	if w.size != SizeDefault {
		Todo("?A better way to do this, no doubt")
		m.A(textSize[w.size])
	}

	Todo("You can't repeat style tags; only the first will be kept")
	if w.fixedHeight != 0 {
		Pr("wtf!!!!!")
		m.A(` style="height:`, w.fixedHeight, `em; background-color:#fcc;"`)
	}

	Pr(m.String())

	m.A(`>`).DoIndent()
	m.A("fixed height:", w.fixedHeight)

	{
		for _, c := range h.Paragraphs() {
			m.A(`<p>`, c, `</p>`).Cr()
		}

	}
	m.DoOutdent()
	m.A(`</div>`).Cr()
}

var textSize = map[WidgetSize]string{
	SizeMicro:  ` style='font-size:.4em'`,
	SizeTiny:   ` style='font-size:.5em'`,
	SizeSmall:  ` style='font-size:.7em'`,
	SizeMedium: ``,
	SizeLarge:  ` style='font-size:1.3em'`,
	SizeHuge:   ` style='font-size:1.8em'`,
}
