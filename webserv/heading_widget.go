package webserv

import (
	. "github.com/jpsember/golang-base/base"
)

// A Widget that displays editable text
type HeadingWidgetStruct struct {
	BaseWidgetObj
	size WidgetSize
}

type HeadingWidget = *HeadingWidgetStruct

func NewHeadingWidget(id string, size WidgetSize) HeadingWidget {
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
		tag := widgetSizeToHeadingTag(w.size)
		m.A(`<`).A(tag).A(` id='`).A(w.Id).A(`'>`)
		m.Escape(value)
		m.A(`</`).A(tag).A(`>`)
	}
	m.Cr()
}

var wsHeadingSize = BuildMap[WidgetSize, string](
	SizeHuge, "h1", SizeLarge, "h2", SizeMedium, "h3", SizeSmall, "h4", SizeTiny, "h5", SizeMicro, "h6",
	SizeDefault, "h3")

func widgetSizeToHeadingTag(widgetSize WidgetSize) string {
	return MapValue(wsHeadingSize, widgetSize)
}
