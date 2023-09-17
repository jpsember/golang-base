package webserv

import (
	. "github.com/jpsember/golang-base/base"
)

type GridCell struct {
	Width int
}

// A concrete Widget that can contain others
type ContainerWidgetObj struct {
	BaseWidgetObj
	children []Widget
	cells    []GridCell
	columns  int // The columns to apply to child widgets
}

type ContainerWidget = *ContainerWidgetObj

func NewContainerWidget(id string) ContainerWidget {
	w := ContainerWidgetObj{
		columns: 12,
	}
	w.BaseId = id
	return &w
}

func (w ContainerWidget) String() string {
	return "<" + w.BaseId + " ContainerWidget>"
}

func (w ContainerWidget) Children() []Widget {
	return w.children
}

func (w ContainerWidget) ClearChildren() {
	w.children = nil
	w.cells = nil
	// Reset the columns to the default (12)
	w.columns = 12
}

func (w ContainerWidget) AddChild(c Widget, manager WidgetManager) {
	w.children = append(w.children, c)
	pr := PrIf(false)
	pr(VERT_SP, "ContainerWidget", w.BaseId, "adding child", c.Id(), "to container", w.BaseId, "columns:", w.columns)
	cols := w.columns
	if cols == 0 {
		BadState("no pending columns for widget:", c.Id())
	}
	cell := GridCell{
		Width: cols,
	}
	pr("added cell, now:", w.cells)
	w.cells = append(w.cells, cell)
}

func (w ContainerWidget) SetColumns(columns int) {
	w.columns = columns
}

func (w ContainerWidget) RenderTo(s Session, m MarkupBuilder) {
	CheckState(len(w.cells) == len(w.children))
	// It is the job of the widget that *contains* us to set the columns that we
	// are to occupy, not ours.
	m.Comments(`ContainerWidget`, w.IdSummary())
	m.OpenTag(`div id='` + w.BaseId + `'`)
	if len(w.children) != 0 {
		m.OpenTag(`div class='row'`)
		for index, child := range w.children {
			cell := w.cells[index]
			str := `div class="col-sm-` + IntToString(cell.Width) + `"`
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
		m.CloseTag().Br()
	}
	m.CloseTag() // ContainerWidget
}
