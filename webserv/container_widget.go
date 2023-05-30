package webserv

import (
	. "github.com/jpsember/golang-base/base"
	//. "github.com/jpsember/golang-base/json"
)

// A concrete Widget that can contain others
type ContainerWidgetObj struct {
	BaseWidgetObj
	Children *Array[Widget]
}

type ContainerWidget = *ContainerWidgetObj

func NewContainerWidget() ContainerWidget {
	w := ContainerWidgetObj{
		Children: NewArray[Widget](),
	}
	return &w
}

func (c ContainerWidget) AddChild(w Widget, gc GridCell) {
	w.GetBaseWidget().Bounds = RectWith(gc.X, gc.Y, gc.Width, 1)
	c.Children.Add(w)
}

func (w ContainerWidget) RenderTo(m MarkupBuilder) {

	desc := `ContainerWidget ` + w.IdSummary()
	m.OpenHtml(`p`, desc).A(desc).CloseHtml(`p`, ``)

	if w.Children.NonEmpty() {
		// We will assume all child views are in grid order
		// We will also assume that they define some number of rows, where each row is completely full
		prevRect := RectWith(-1, -1, 0, 0)
		for _, child := range w.Children.Array() {
			bw := child.GetBaseWidget()
			b := &bw.Bounds
			CheckArg(b.IsValid())
			if b.Location.Y > prevRect.Location.Y {
				if prevRect.Location.Y >= 0 {
					m.CloseHtml("div", "end of row")
					m.Br()
				}
				m.Br()
				m.OpenHtml(`div class="row"`, `start of row`)
				m.Cr()
			}
			prevRect = *b
			m.OpenHtml(`div class="col-sm-`+IntToString(b.Size.W)+`"`, `child`)
			child.RenderTo(m)
			m.CloseHtml(`div`, `child`)
		}
		m.CloseHtml("div", "row")
		m.Br()
	}
}

func (w ContainerWidget) assignViewsToGridLayout(grid Grid) {
	pr := PrIf(true)
	pr("assignViewsToGridLayout, grid:", INDENT, grid)

	grid.PropagateGrowFlags()
	containerWidget := grid.Widget().(ContainerWidget)
	pr("number of children:", containerWidget.Children.Size())

	gridWidth := grid.NumColumns()
	gridHeight := grid.NumRows()

	for gridY := 0; gridY < gridHeight; gridY++ {
		for gridX := 0; gridX < gridWidth; gridX++ {
			cell := grid.cellAt(gridX, gridY)
			if cell.IsEmpty() {
				continue
			}

			// If cell's coordinates don't match our iteration coordinates, we've
			// already added this cell
			if cell.X != gridX || cell.Y != gridY {
				continue
			}

			widget := cell.Widget

			containerWidget.AddChild(widget, cell)
		}
	}
}
