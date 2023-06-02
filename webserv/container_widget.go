package webserv

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/json"
)

// A concrete Widget that can contain others
type ContainerWidgetObj struct {
	BaseWidgetObj
	children               *Array[Widget]
	cells                  *Array[GridCell]
	columnSizes            []int
	cachedNextCellLocation IPoint
	nextCellKnown          bool
}

type ContainerWidget = *ContainerWidgetObj

func NewContainerWidget(id string, columnSizes []int) ContainerWidget {
	w := ContainerWidgetObj{
		children:    NewArray[Widget](),
		cells:       NewArray[GridCell](),
		columnSizes: columnSizes,
	}
	w.Id = id
	return &w
}

func (w ContainerWidget) GetChildren() []Widget {
	return w.children.Array()
}

func (w ContainerWidget) RenderTo(m MarkupBuilder, state JSMap) {

	desc := `ContainerWidget ` + w.IdSummary()
	m.OpenHtml(`p`, desc).A(desc).CloseHtml(`p`, ``)

	if w.children.NonEmpty() {
		// We will assume all child views are in grid order
		// We will also assume that they define some number of rows, where each row is completely full
		prevRect := RectWith(-1, -1, 0, 0)
		for _, child := range w.children.Array() {
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
			Todo("base widths on a 12-column scale")
			m.OpenHtml(`div class="col-sm-`+IntToString(b.Size.W*4)+`"`, `child`)
			child.RenderTo(m, state)
			m.CloseHtml(`div`, `child`)
		}
		m.CloseHtml("div", "row")
		m.Br()
	}
}

func (w ContainerWidget) EndRow(manager WidgetManager) {
	if w.nextCellLocation().X != 0 {
		manager.Spanx().AddHorzSpace()
		w.nextCellKnown = false
	}
}

func (w ContainerWidget) LayoutChildren(manager WidgetManager) {
	Todo("cells need refactoring, padding with empty if not full")
	pr := PrIf(false)
	pr("LayoutChildren:", INDENT, w)

	Todo("try to avoid having Layout call back to manager, adding additional children, etc")
	// If current row is only partially complete, add space to its end
	if w.nextCellLocation().X != 0 {
		manager.Spanx().AddHorzSpace()
	}

	w.propagateGrowFlags()
	pr("number of children:", w.children.Size())

	gridWidth := w.numColumns()
	gridHeight := w.numRows()

	pr("grid size:", gridWidth, gridHeight)
	for gridY := 0; gridY < gridHeight; gridY++ {
		for gridX := 0; gridX < gridWidth; gridX++ {
			cell := w.cellAt(gridX, gridY)
			if cell.IsEmpty() {
				continue
			}
			// If cell's coordinates don't match our iteration coordinates, we've
			// already added this cell
			if cell.X != gridX || cell.Y != gridY {
				continue
			}
			wg := cell.Widget
			wg.GetBaseWidget().Bounds = RectWith(cell.X, cell.Y, cell.Width, 1)
		}
	}
}

func (m ContainerWidget) AddChild(c Widget, manager WidgetManager) {
	m.children.Add(c)

	pr := PrIf(true)

	pr("adding widget to container:", INDENT, m)

	Todo("can we simplify things by not having a separate Cell object for the grid?")
	cell := NewGridCell()
	cell.Widget = c
	Todo("does a cell need to have a widget pointer?")
	nextGridCellLocation := m.nextCellLocation()
	cell.X = nextGridCellLocation.X
	cell.Y = nextGridCellLocation.Y

	pr("determine loc, size in cells; SpanXCount:", manager.SpanXCount)
	// determine location and size, in cells, of component
	cols := 1
	if manager.SpanXCount != 0 {
		remainingCols := m.numColumns() - cell.X
		pr("num columns:", m.numColumns())
		pr("remaining cols:", remainingCols)
		if manager.SpanXCount < 0 {
			cols = remainingCols
		} else {
			if manager.SpanXCount > remainingCols {
				BadState("requested span of ", manager.SpanXCount, " yet only ", remainingCols, " remain")
			}
			cols = manager.SpanXCount
		}
	}
	cell.Width = cols

	cell.GrowX = manager.GrowXWeight
	cell.GrowY = manager.GrowYWeight

	// If any of the spanned columns have 'grow' flag set, set it for this component
	for i := cell.X; i < cell.X+cell.Width; i++ {
		colSize := m.columnSizes[i]
		cell.GrowX = MaxInt(cell.GrowX, colSize)
	}

	pr("cell width:", cell.Width)
	// "paint" the cells this view occupies by storing a copy of the entry in each cell
	for i := 0; i < cols; i++ {
		m.addCell(cell)
	}

}

func (g ContainerWidget) String() string {
	m := NewJSMap()
	m.Put("", "ContainerWidget")
	m.Put("# cells", g.cells.Size())
	m.Put("ColumnSizes", JSListWith(g.columnSizes))
	return m.String()
}

func (g ContainerWidget) numColumns() int {
	return len(g.columnSizes)
}

func (g ContainerWidget) nextCellLocation() IPoint {
	if !g.nextCellKnown {
		x := 0
		y := 0
		if g.cells.NonEmpty() {

			lastCell := g.cells.Last()

			x = lastCell.X + lastCell.Width
			y = lastCell.Y
			CheckState(x <= g.numColumns())
			if x == g.numColumns() {
				x = 0
				y += 1
			}
		}
		g.cachedNextCellLocation = IPointWith(x, y)
		g.nextCellKnown = true
	}
	return g.cachedNextCellLocation
}

func (g ContainerWidget) numRows() int {
	nextLoc := g.nextCellLocation()
	y := nextLoc.Y
	if nextLoc.X > 0 {
		y++
	}
	return y
}

func (g ContainerWidget) checkValidColumn(x int) int {
	if x < 0 || x >= g.numColumns() {
		BadArg("not a valid column:", x)
	}
	return x
}

func (g ContainerWidget) checkValidRow(y int) int {
	if y < 0 || y >= g.numRows() {
		BadArg("not a valid row:", y)
	}
	return y
}

func (g ContainerWidget) cellAt(x int, y int) GridCell {
	i := g.checkValidRow(y)*g.numColumns() + g.checkValidColumn(x)
	return g.cells.Get(i)
}

func (g ContainerWidget) addCell(cell GridCell) {
	g.cells.Add(cell)
	Pr("addCell, size now:", g.cells.Size())
	g.nextCellKnown = false
}

func (g ContainerWidget) propagateGrowFlags() {
	Todo("PropagateGrowFlags can no doubt be simplified")
	cs := g.cells.Size()
	var colGrowFlags = make([]int, cs)
	var rowGrowFlags = make([]int, cs)

	for _, cell := range g.cells.Array() {
		if cell.IsEmpty() {
			continue
		}

		// If view occupies multiple cells horizontally, don't propagate its grow flag
		if cell.GrowX > 0 && cell.Width == 1 {
			if colGrowFlags[cell.X] < cell.GrowX {
				colGrowFlags[cell.X] = cell.GrowX
			}
		}
		// If view occupies multiple cells vertically, don't propagate its grow flag
		// (at present, we don't support views stretching across multiple rows)
		if cell.GrowY > 0 {
			if rowGrowFlags[cell.Y] < cell.GrowY {
				rowGrowFlags[cell.Y] = cell.GrowY
			}
		}
	}

	// Now propagate grow flags from bit sets back to individual cells
	for _, cell := range g.cells.Array() {

		if cell.IsEmpty() {
			continue
		}

		for x := cell.X; x < cell.X+cell.Width; x++ {
			cell.GrowX = MaxInt(cell.GrowX, colGrowFlags[x])
		}
		cell.GrowY = rowGrowFlags[cell.Y]
	}
}
