package webserv

import (
	. "github.com/jpsember/golang-base/base"
)

type DebugWidgetObj struct {
	BaseWidgetObj
}

type DebugWidget = *DebugWidgetObj

// Deprecated.  Maybe just use BaseWidget?
func NewDebugWidget(id string) DebugWidget {
	Todo("<1Use BaseWidget instead?")
	t := &DebugWidgetObj{}
	t.Id = id
	return t
}

func (w DebugWidget) RenderTo(m MarkupBuilder, state JSMap) {
	if !w.Visible() {
		m.RenderInvisible(w)
		return
	}

	m.A(`<div id='`)
	m.A(w.Id)
	m.A(`'>`)
	m.A(`</div>`)
	m.Cr()
}
