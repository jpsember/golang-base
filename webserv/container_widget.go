package webserv

import (
	. "github.com/jpsember/golang-base/base"
	"reflect"
	"strings"
)

type GridCell struct {
	Location IPoint
	Width    int
}

func (g *GridCell) String() string {
	m := NewJSMap()
	m.Put("", "GridCell")
	m.Put("Location", g.Location.String())
	m.Put("Width", g.Width)
	return m.AsString()
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
	if len(w.cells) != 0 {
		c := Last(w.cells)
		cell.Location = IPointWith(c.Location.X+c.Width, c.Location.Y)
	}
	if cell.Location.X+cell.Width > MaxColumns {
		cell.Location = IPointWith(0, cell.Location.Y+1)
		Todo("!add support for cell heights > 1")
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
	m.Comments(`ContainerWidget`, w.IdSummary()).OpenTag(`div id='` + w.BaseId + `'`)
	if w.Visible() {
		prevPoint := IPointWith(0, -1)
		for index, child := range w.children {
			cell := w.cells[index]
			// If this cell lies in a row below the current, Close the current and start a new one
			if cell.Location.Y > prevPoint.Y {
				if prevPoint.Y >= 0 {
					m.CloseTag() // row
				}
				m.OpenTag(`div class='row'`)
				prevPoint = IPointWith(0, cell.Location.Y)
			}

			str := `div class="col-sm-` + IntToString(cell.Width) + `"`
			if WidgetDebugRenderingFlag {
				str += ` style="background-color:` + DebugColorForString(child.Id()) + `;`
				str += `border-style:double;`
				str += `"`
			}
			m.Comments(`child`).OpenTag(str)
			if false && WidgetDebugRenderingFlag {
				// Render a div that contains some information
				{
					m.A(`<div style="font-size:50%; font-family:monospace;">`)
				}

				id := child.Id()
				if id[0] != '.' /* || Alert("Including anon ids" )*/ {
					m.A(`Id:`, id, ` `)
				}
				m.A(`Cols:`, cell.Width, ` `)
				{
					widgetType := reflect.TypeOf(child).String()
					i := strings.LastIndex(widgetType, ".")
					widgetType = widgetType[i+1:]
					widgetType = strings.TrimSuffix(widgetType, "Obj")
					m.A(widgetType, ` `)
				}

				m.A(`</div>`).Cr()
			}

			verify := m.VerifyBegin()
			child.RenderTo(s, m)
			m.VerifyEnd(verify, child)

			m.CloseTag() // child
			prevPoint = IPointWith(cell.Location.X+cell.Width, cell.Location.Y)
		}
		if prevPoint.Y >= 0 {
			m.CloseTag() // row
			m.Br()
		}
	}
	m.CloseTag() // ContainerWidget
}
