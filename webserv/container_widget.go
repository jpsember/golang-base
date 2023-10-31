package webserv

import (
	. "github.com/jpsember/golang-base/base"
)

// A concrete Widget that can contain others, using Bootstrap's grid system of rows and columns
type GridWidgetStruct struct {
	BaseWidgetObj
	children []Widget
}

type GridWidget = *GridWidgetStruct

func NewContainerWidget(id string, clickListener ButtonWidgetListener) GridWidget {
	w := GridWidgetStruct{}
	w.InitBase(id)
	if clickListener != nil {
		w.SetLowListener(func(sess Session, widget Widget, value string, args []string) (any, error) {
			clickListener(sess, widget, value)
			return nil, nil
		})
	}
	return &w
}

func (w GridWidget) String() string {
	return "<" + w.Id() + " GridWidget>"
}

func (w GridWidget) Children() []Widget {
	return w.children
}

func (w GridWidget) ClearChildren() {
	w.children = nil
}

func (w GridWidget) AddChild(c Widget, manager WidgetManager) {
	cols := manager.stackedState().pendingChildColumns
	if cols == 0 {
		BadState("no pending columns for widget:", c.Id())
	}
	c.SetColumns(cols)
	w.children = append(w.children, c)
	pr := PrIf("", false)
	pr(VERT_SP, "GridWidget", w.Id(), "adding child", c.Id(), "to container", w.Id(), "columns:", w.Columns())
}

func (w GridWidget) RemoveChild(c Widget) {
	Die("probably shouldn't be necessary")
	for index, child := range w.children {
		if child == c {
			w.children = DeleteSliceElements(w.children, index, 1)
			return
		}
	}
	BadArg("Child wasn't in container:", c.Id())
}

func (w GridWidget) RenderTo(s Session, m MarkupBuilder) {
	// It is the job of the widget that *contains* us to set the columns that we
	// are to occupy, not ours.
	Todo("!Don't add markup that is outside of the div<widget id>, else it will pile up due to ajax refreshes")
	m.TgOpen(`div id=`).A(QUO, s.PrependId(w.Id())).TgContent()
	m.Comments(`GridWidget`, w.IdSummary())

	anyPlotted := false
	for _, child := range w.children {
		if child.Detached() {
			continue
		}
		if !anyPlotted {
			anyPlotted = true
			m.TgOpen(`div class='row'`).TgContent()
		}

		m.TgOpen(`div class="col-sm-`).A(child.Columns(), `"`)
		if WidgetDebugRenderingFlag {
			m.Style(`background-color:`, DebugColorForString(child.Id()), `;`)
			m.Style(`border-style:double;`)
		}

		if w.LowListen != nil {
			m.A(` onclick="jsButton('`, s.ClickPrefix(), w.Id(), `')"`)
		}

		m.TgContent()
		{
			verify := m.VerifyBegin()
			RenderWidget(child, s, m)
			m.VerifyEnd(verify, child)
		}
		m.TgClose()
	}
	if anyPlotted {
		m.TgClose().Br()
	}
	m.TgClose()
}
