package webserv

import (
	. "github.com/jpsember/golang-base/base"
)

type TextWidgetObj struct {
	BaseWidgetObj
}

type TextWidget = *TextWidgetObj

func NewTextWidget(id string) TextWidget {
	t := &TextWidgetObj{}
	t.Id = id
	return t
}

func (w TextWidget) RenderTo(m MarkupBuilder, state JSMap) {
	// b := w.Base()
	if !w.Visible() {
		m.RenderInvisible(w)
		return
	}

	textContent, wasStatic := GetStaticOrDynamicLabel(w, state)
	//var textContent string
	//
	//Todo("have utility method for this, useful for Heading too")
	//sc := b.StaticContent()
	//hasStaticContent := sc != nil
	//if hasStaticContent {
	//	textContent = sc.(string)
	//} else {
	//	textContent = state.OptString(w.Id, "")
	//}

	h := NewHtmlString(textContent)
	if wasStatic {
		m.OpenTag(`div`)
	} else {
		m.OpenTag(`div id='` + w.Id + `'`)
	}

	for _, c := range h.Paragraphs() {
		m.A(`<p>`, c, `</p>`).Cr()
	}
	m.CloseTag()
}
