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
	//t.SetTrace(true)
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
		w.Log("RenderTo, staticContent:", w.staticContent)
	} else {
		w.Log("RenderTo, reading widget string value; state provider:", w.stateProvider, "id:", w.Id())
		textContent = s.WidgetStringValue(w)
	}
	w.Log("...text value:", Quoted(textContent))

	h := NewHtmlString(textContent)

	m.TgOpen(`div id=`).A(QUOTED, s.PrependId(w.Id()))

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
