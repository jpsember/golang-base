package webserv

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/json"
)

// A concrete Widget that can contain others
type ContainerWidgetObj struct {
	BaseWidgetObj
	Children *Array[Widget]

	mDebugContext string
	mCells        *Array[GridCell]
	ColumnSizes   []int

	mCachedNextCellLocation IPoint
	mNextCellKnown          bool
}

type ContainerWidget = *ContainerWidgetObj

func NewContainerWidget(context string) ContainerWidget {
	w := ContainerWidgetObj{
		Children:      NewArray[Widget](),
		mDebugContext: context,
		mCells:        NewArray[GridCell](),
	}
	return &w
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

func (w ContainerWidget) LayoutChildren(manager WidgetManager) {
	pr := PrIf(true)
	pr("LayoutChildren:", INDENT, w)

	Todo("try to avoid having Layout call back to manager, adding additional children, etc")
	// If current row is only partially complete, add space to its end
	if w.NextCellLocation().X != 0 {
		manager.Spanx().AddHorzSpace()
	}

	w.propagateGrowFlags()
	pr("number of children:", w.Children.Size())

	gridWidth := w.NumColumns()
	gridHeight := w.NumRows()

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
	Todo("do we need to expose lots of fields of WidgetManager here?")
	m.Children.Add(c)

	pr := PrIf(true)

	pr("adding widget to container:", INDENT, m)

	Todo("can we simplify things by not having a separate Cell object for the grid?")
	cell := NewGridCell()
	cell.Widget = c
	Todo("does a cell need to have a widget pointer?")
	nextGridCellLocation := m.NextCellLocation()
	cell.X = nextGridCellLocation.X
	cell.Y = nextGridCellLocation.Y

	// determine location and size, in cells, of component
	cols := 1
	if manager.mSpanXCount != 0 {
		remainingCols := m.NumColumns() - cell.X
		if manager.mSpanXCount < 0 {
			cols = remainingCols
		} else {
			if manager.mSpanXCount > remainingCols {
				BadState("requested span of ", manager.mSpanXCount, " yet only ", remainingCols, " remain")
			}
			cols = manager.mSpanXCount
		}
	}
	cell.Width = cols

	cell.GrowX = manager.mGrowXFlag
	cell.GrowY = manager.mGrowYFlag

	// If any of the spanned columns have 'grow' flag set, set it for this component
	for i := cell.X; i < cell.X+cell.Width; i++ {
		colSize := m.ColumnSizes[i]
		cell.GrowX = MaxInt(cell.GrowX, colSize)
	}

	// "paint" the cells this view occupies by storing a copy of the entry in each cell
	for i := 0; i < cols; i++ {
		m.AddCell(cell)
	}

}

func (g ContainerWidget) String() string {
	m := NewJSMap()
	m.Put("", "ContainerWidget")
	m.Put("context", g.mDebugContext)
	m.Put("# cells", g.mCells.Size())
	m.Put("ColumnSizes", JSListWith(g.ColumnSizes))
	return m.String()
}

func (g ContainerWidget) DebugContext() string {
	return g.mDebugContext
}

func (g ContainerWidget) NumColumns() int {
	return len(g.ColumnSizes)
}

func (g ContainerWidget) NextCellLocation() IPoint {
	if !g.mNextCellKnown {
		x := 0
		y := 0
		if g.mCells.NonEmpty() {

			lastCell := g.mCells.Last()

			x = lastCell.X + lastCell.Width
			y = lastCell.Y
			CheckState(x <= g.NumColumns())
			if x == g.NumColumns() {
				x = 0
				y += 1
			}
		}
		g.mCachedNextCellLocation = IPointWith(x, y)
		g.mNextCellKnown = true
	}
	return g.mCachedNextCellLocation
}

func (g ContainerWidget) NumRows() int {
	nextLoc := g.NextCellLocation()
	y := nextLoc.Y
	if nextLoc.X > 0 {
		y++
	}
	return y
}

func (g ContainerWidget) checkValidColumn(x int) int {
	if x < 0 || x >= g.NumColumns() {
		BadArg("not a valid column:", x)
	}
	return x
}

func (g ContainerWidget) checkValidRow(y int) int {
	if y < 0 || y >= g.NumRows() {
		BadArg("not a valid row:", y)
	}
	return y
}

func (g ContainerWidget) cellAt(x int, y int) GridCell {
	return g.mCells.Get(g.checkValidRow(y)*g.NumColumns() + g.checkValidColumn(x))
}

func (g ContainerWidget) AddCell(cell GridCell) {
	g.mCells.Add(cell)
	g.mNextCellKnown = false
}

/**
 * Get list of cells... must be considered READ ONLY
 */
func (g ContainerWidget) cells() *Array[GridCell] {
	return g.mCells
}

func (g ContainerWidget) propagateGrowFlags() {
	Todo("PropagateGrowFlags can no doubt be simplified")
	cs := g.cells().Size()
	var colGrowFlags = make([]int, cs)
	var rowGrowFlags = make([]int, cs)

	for _, cell := range g.cells().Array() {
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
	for _, cell := range g.cells().Array() {

		if cell.IsEmpty() {
			continue
		}

		for x := cell.X; x < cell.X+cell.Width; x++ {
			cell.GrowX = MaxInt(cell.GrowX, colGrowFlags[x])
		}
		cell.GrowY = rowGrowFlags[cell.Y]
	}
}
