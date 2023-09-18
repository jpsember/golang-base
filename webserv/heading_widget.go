package webserv

import (
	. "github.com/jpsember/golang-base/base"
)

// A Widget that displays editable text
type HeadingWidgetStruct struct {
	BaseWidgetObj
}

type HeadingWidget = *HeadingWidgetStruct

func NewHeadingWidget(id string) HeadingWidget {
	w := HeadingWidgetStruct{}
	w.InitBase(id)
	return &w
}

func (w HeadingWidget) RenderTo(s Session, m MarkupBuilder) {
	Todo("!is this the most appropriate accessor?")
	textContent := ReadWidgetString(w, s)
	Pr("HeadingWidget", w.Id(), "RenderTo; textContent:", Quoted(textContent))
	if Alert("setting some non-empty text") && textContent == "" {
		textContent = "abra cadabra"
	}
	tag := wsHeadingSize[w.Size()]
	m.A(`<`, tag)

	// Have some special handling for the Micro size; very small text, and right justified
	if w.size == SizeMicro {
		m.A(` style="font-size:50%"`)
	}
	tx := wsHeadingAlign[w.Align()]
	if tx != "" {
		m.A(` class="`, tx, `"`)
	}
	m.A(` id='`, w.BaseId, `'>`)
	m.Escape(textContent)
	m.A(`</`, tag, `>`)
	m.Cr()
}

var wsHeadingSize = map[WidgetSize]string{
	SizeHuge:    "h1",
	SizeLarge:   "h2",
	SizeMedium:  "h3",
	SizeSmall:   "h4",
	SizeTiny:    "h5",
	SizeMicro:   "h6",
	SizeDefault: "h3",
}

var wsHeadingAlign = map[WidgetAlign]string{
	AlignRight:  "text-end",
	AlignCenter: "text-center",
	AlignLeft:   "text-left",
}
