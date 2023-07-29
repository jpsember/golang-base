package webserv

import (
	. "github.com/jpsember/golang-base/base"
	"strings"
)

type DebugWidgetObj struct {
	BaseWidgetObj
	assignedColumns int
	BgndColor       string
}

type DebugWidget = *DebugWidgetObj

func NewDebugWidget(id string) DebugWidget {
	t := &DebugWidgetObj{}
	t.Id = id

	const bgColors = "#fc7f03#fcce03#58bf58#4aa3b5#cfa8ed#fa7fc1#b2f7a6#b2f7a6#90adad#3588cc#b06dfc"

	{
		const wlen = 7
		c := Rand().Intn(len(bgColors)/wlen) * wlen
		t.BgndColor = bgColors[c : c+wlen]
	}
	return t
}

func (w DebugWidget) RenderTo(m MarkupBuilder, state JSMap) {
	if !w.Visible() {
		m.RenderInvisible(w)
		return
	}

	{
		m.A(`<div id='`)
		m.A(w.Id)
		m.A(`'`)
		m.A(` style="font-size:60%; font-family:monospace; min-height:4em;"`)
		m.A(`>`)
	}
	s := strings.Builder{}
	s.WriteString("#" + w.Id)
	s.WriteString(" C" + IntToString(w.assignedColumns))
	s.WriteString(" B" + w.BgndColor)

	Todo("!Have utility method to construct and append to strings.Builder")
	m.Escape(s.String())

	m.A(`</div>`)

	m.Cr()
}

func (w DebugWidget) SetAssignedColumns(width int) {
	w.assignedColumns = width
}
