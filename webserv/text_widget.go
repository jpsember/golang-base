package webserv

import (
	. "github.com/jpsember/golang-base/base"
)

type TextWidgetObj struct {
	BaseWidgetObj
	size WidgetSize
}

type TextWidget = *TextWidgetObj

func NewTextWidget(id string, size WidgetSize) TextWidget {
	t := &TextWidgetObj{
		size: size,
	}
	t.BaseId = id
	return t
}

func (w TextWidget) RenderTo(m MarkupBuilder, state JSMap) {
	if !w.Visible() {
		m.RenderInvisible(w)
		return
	}

	textContent, wasStatic := GetStaticOrDynamicLabel(w, state)

	h := NewHtmlString(textContent)

	args := NewArray[any]()
	args.Add(`div`)

	if !wasStatic {
		args.Add(`div id='` + w.BaseId + `'`)
		m.OpenTag(`div id='` + w.BaseId + `'`)
	}
	if w.size != SizeDefault {
		Todo("?A better way to do this, no doubt")
		args.Add(textSize[w.size])
	}

	m.OpenTag(args.Array()...)

	{

		for _, c := range h.Paragraphs() {
			m.A(`<p>`, c, `</p>`).Cr()
		}

	}
	m.CloseTag()
}

var textSize = map[WidgetSize]string{
	SizeMicro:  ` style='font-size:.4em'`,
	SizeTiny:   ` style='font-size:.5em'`,
	SizeSmall:  ` style='font-size:.7em'`,
	SizeMedium: ``,
	SizeLarge:  ` style='font-size:1.3em'`,
	SizeHuge:   ` style='font-size:1.8em'`,
}
