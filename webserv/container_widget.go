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

func NewContainerWidget(id string) GridWidget {
	w := GridWidgetStruct{}
	w.InitBase(id)
	return &w
}

func (w GridWidget) String() string {
	return "<" + w.BaseId + " GridWidget>"
}

func (w GridWidget) Children() []Widget {
	return w.children
}

func (w GridWidget) ClearChildren() {
	w.children = nil
}

func (w GridWidget) AddChild(c Widget, manager WidgetManager) {
	cols := manager.pendingChildColumns
	if cols == 0 {
		BadState("no pending columns for widget:", c.Id())
	}
	c.SetColumns(cols)
	w.children = append(w.children, c)
	pr := PrIf(false)
	pr(VERT_SP, "GridWidget", w.BaseId, "adding child", c.Id(), "to container", w.BaseId, "columns:", w.Columns())
}

func (w GridWidget) RemoveChild(c Widget) {
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
	m.TgOpen(`div id=`).A(QUOTED, w.BaseId).TgContent()
	m.Comments(`GridWidget`, w.IdSummary())
	if len(w.children) != 0 {
		m.TgOpen(`div class='row'`).TgContent()
		for _, child := range w.children {
			str := `div class="col-sm-` + IntToString(child.Columns()) + `"`
			if WidgetDebugRenderingFlag {
				str += ` style="background-color:` + DebugColorForString(child.Id()) + `;`
				str += `border-style:double;`
				str += `"`
			}
			m.Comments(`child`).OpenTag(str)
			{
				verify := m.VerifyBegin()
				RenderWidget(child, s, m)
				m.VerifyEnd(verify, child)
			}
			m.CloseTag()
		}
		m.TgClose().Br()
	}
	m.TgClose() // GridWidget
}
