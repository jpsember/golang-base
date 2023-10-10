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
	pr := PrIf("TextWidget.RenderTo", true)
	var textContent string
	if w.staticContent != nil {
		textContent = w.staticContent.(string)
		w.Log("RenderTo, staticContent:", w.staticContent)
	} else {
		pr("RenderTo, reading widget string value; state provider:", w.stateProvider, "id:", w.Id())
		textContent = s.WidgetStringValue(w)
	}
	w.Log("...text value:", Quoted(textContent))

	h := NewHtmlString(textContent)

	Alert("Do we really want to prepend the id here?")
	Alert("The StateProvider.prefix is overloaded; 1) add prefix to widget id when rendering; 2) interpret to change semantics on events (list)")
	prefixedId := s.PrependId(w.Id())

	m.TgOpen(`div id=`).A(QUO, prefixedId)

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
