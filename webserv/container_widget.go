package webserv

import (
	. "github.com/jpsember/golang-base/base"
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
	w := ContainerWidgetObj{
		children: NewArray[Widget](),
		cells:    NewArray[GridCell](),
		columns:  12,
	}
	w.Id = id
	return &w
}

func (w ContainerWidget) GetChildren() []Widget {
	return w.children.Array()
}

func (w ContainerWidget) AddChild(c Widget, manager WidgetManager) {
	w.children.Add(c)
	pr := PrIf(false)
	pr("adding widget to container:", INDENT, w)
	cols := w.columns
	if cols == 0 {
		BadState("no pending columns for widget:", c.GetBaseWidget().Id)
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
	w.cells.Add(cell)

	if Alert("!Have a 'debug features' for this kind of thing") {
		if dw, ok := c.(DebugWidget); ok {
			dw.SetAssignedColumns(cell.Width)
		}

	}
}

func (w ContainerWidget) columnsTag(columns int, optionalWidget Widget) string {
	s := `div class="col-sm-` + IntToString(columns) + `"`

	if Alert("!Have debug flag for this") {
		if optionalWidget != nil {
			if deb, ok := optionalWidget.(DebugWidget); ok {
				c := deb.BgndColor
				if c != "" {
					s += ` style="background-color:` + deb.BgndColor + `"`
				}
			}
		}
	}
	return s
}

func (w ContainerWidget) SetColumns(columns int) {
	w.columns = columns
}

func (w ContainerWidget) RenderTo(m MarkupBuilder, state JSMap) {
	m.Comments(false)
	desc := `ContainerWidget ` + w.IdSummary()
	// It is the job of the widget that *contains* us to set the columns that we
	// are to occupy, not ours.
	m.OpenHtml(`div id='`+w.Id+`'`, desc)
	if w.Visible() {
		prevPoint := IPointWith(0, -1)
		for index, child := range w.children.Array() {
			cell := w.cells.Get(index)
			// If this cell lies in a row below the current, Close the current and start a new one
			if cell.Location.Y > prevPoint.Y {
				if prevPoint.Y >= 0 {
					m.CloseHtml("div", "end of row")
				}
				m.Br()
				m.OpenHtml(`div class='row'`, `start of row`)
				m.Cr()
				prevPoint = IPointWith(0, cell.Location.Y)
			}

			// If cell lies to right of current, add space
			spaceColumns := cell.Location.X - prevPoint.X
			if spaceColumns > 0 {
				m.OpenHtml(w.columnsTag(spaceColumns, nil), `spacer`)
				child.RenderTo(m, state)
				m.CloseHtml(`div`, `spacer`)
			}

			m.OpenHtml(w.columnsTag(cell.Width, child), `child`)
			child.RenderTo(m, state)
			m.CloseHtml(`div`, `child`)
			prevPoint = IPointWith(cell.Location.X+cell.Width, cell.Location.Y)
		}
		if prevPoint.Y >= 0 {
			m.CloseHtml("div", "row")
			m.Br()
		}
	}
	m.CloseHtml(`div`, desc)
	m.Comments(true)
}

func (w ContainerWidget) LayoutChildren(manager WidgetManager) {
	// We no longer need to do anything here, as the cells generated by AddChild() do most of the work
	pr := PrIf(false)
	pr("LayoutChildren:", INDENT, w)
}
