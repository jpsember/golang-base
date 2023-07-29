package webserv

import (
	. "github.com/jpsember/golang-base/base"
	"strings"
)

type DebugWidgetObj struct {
	BaseWidgetObj
	assignedColumns int
	bgndColor       string
}

type DebugWidget = *DebugWidgetObj

func NewDebugWidget(id string) DebugWidget {
	t := &DebugWidgetObj{}
	t.Id = id
	return t
}

func (w DebugWidget) RenderTo(m MarkupBuilder, state JSMap) {
	if !w.Visible() {
		m.RenderInvisible(w)
		return
	}

	const bgColors = "#fc7f03#fcce03#58bf58#4aa3b5#cfa8ed#fa7fc1#b2f7a6#b2f7a6#90adad#3588cc#b06dfc"

	{
		m.A(`<div id='`)
		m.A(w.Id)
		m.A(`'`)

		if w.bgndColor == "" {
			const wlen = 7
			c := Rand().Intn(len(bgColors)/wlen) * wlen
			w.bgndColor = bgColors[c : c+wlen]
		}

		m.Pr(` style="background-color:`, w.bgndColor, `;"`, `>`)
	}
	s := strings.Builder{}
	s.WriteString(ToString("Id:", w.Id))
	s.WriteString(ToString(" Cols:", w.assignedColumns))
	s.WriteString(ToString(" Bg:", w.bgndColor))

	Todo("!Have utility method to construct and append to strings.Builder")
	m.Escape(s.String())

	m.A(`</div>`)

	m.Cr()
}

func (w DebugWidget) SetAssignedColumns(width int) {
	w.assignedColumns = width
}
