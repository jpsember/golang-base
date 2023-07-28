package webserv

import (
	. "github.com/jpsember/golang-base/base"
)

// A Widget that displays editable text
type HeadingWidgetStruct struct {
	BaseWidgetObj
	size int
}

type HeadingWidget = *HeadingWidgetStruct

func NewHeadingWidget(id string, size int) HeadingWidget {
	w := HeadingWidgetStruct{
		size: size,
	}
	w.Id = id
	return &w
}

func (w HeadingWidget) RenderTo(m MarkupBuilder, state JSMap) {
	if !w.Visible() {
		m.RenderInvisible(w)
	} else {
		value := WidgetStringValue(state, w.Id)
		tag := `h` + IntToString(w.size)
		m.A(`<`).A(tag).A(` id='`).A(w.Id).A(`'>`)
		m.Escape(value)
		m.A(`</`).A(tag).A(`>`)
	}
	m.Cr()
}
