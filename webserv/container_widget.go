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
	children *Array[Widget]
	cells    *Array[GridCell]
	columns  int // The columns to apply to child widgets
}

type ContainerWidget = *ContainerWidgetObj

func NewContainerWidget(id string) ContainerWidget {
	Todo("!might simplify a lot of things if widgets had references to their own managers")
	w := ContainerWidgetObj{
		children: NewArray[Widget](),
		cells:    NewArray[GridCell](),
		columns:  12,
	}
	w.BaseId = id
	return &w
}

func (w ContainerWidget) String() string {
	return "<" + w.BaseId + " ContainerWidget>"
}

func (w ContainerWidget) Children() *Array[Widget] {
	return w.children
}

func (w ContainerWidget) ClearChildren() {
	w.children.Clear()
	w.cells.Clear()
	// Reset the columns to the default (12)
	w.columns = 12
}

func (w ContainerWidget) AddChild(c Widget, manager WidgetManager) {
	w.children.Add(c)
	pr := PrIf(false)
	pr(VERT_SP, "ContainerWidget", w.BaseId, "adding child", c.Id(), "to container", w.BaseId, "columns:", w.columns)
	cols := w.columns
	if cols == 0 {
		BadState("no pending columns for widget:", c.Id())
	}
	cell := GridCell{
		Width: cols,
	}
	if w.cells.NonEmpty() {
		c := w.cells.Last()
		cell.Location = IPointWith(c.Location.X+c.Width, c.Location.Y)
	}
	if cell.Location.X+cell.Width > MaxColumns {
		cell.Location = IPointWith(0, cell.Location.Y+1)
		Todo("!add support for cell heights > 1")
	}
	pr("added cell, now:", w.cells)
	w.cells.Add(cell)
}

func (w ContainerWidget) SetColumns(columns int) {
	w.columns = columns
}

func (w ContainerWidget) RenderTo(m MarkupBuilder, state JSMap) {
	CheckState(w.cells.Size() == w.children.Size())
	// It is the job of the widget that *contains* us to set the columns that we
	// are to occupy, not ours.
	m.Comments(`ContainerWidget`, w.IdSummary()).OpenTag(`div id='` + w.BaseId + `'`)
	if w.Visible() {
		prevPoint := IPointWith(0, -1)
		for index, child := range w.children.Array() {
			cell := w.cells.Get(index)
			// If this cell lies in a row below the current, Close the current and start a new one
			if cell.Location.Y > prevPoint.Y {
				if prevPoint.Y >= 0 {
					m.CloseTag() // row
				}
				m.OpenTag(`div class='row'`)
				prevPoint = IPointWith(0, cell.Location.Y)
			}

			s := `div class="col-sm-` + IntToString(cell.Width) + `"`
			if WidgetDebugRenderingFlag {
				s += ` style="background-color:` + DebugColorForString(child.Id()) + `;`
				s += `border-style:double;`
				s += `"`
			}
			m.Comments(`child`).OpenTag(s)
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
			child.RenderTo(m, state)
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
