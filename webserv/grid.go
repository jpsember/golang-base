package webserv

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/json"
)

// Data for a View that contains a grid of child views
type GridObj struct {
	mDebugContext string
	mCells        *Array[GridCell]
	ColumnSizes   []int

	mCachedNextCellLocation IPoint
	mNextCellKnown          bool
	mWidget                 Widget
}
type Grid = *GridObj

func NewGrid() Grid {
	g := GridObj{
		mDebugContext: "<no context>",
		mCells:        NewArray[GridCell](),
	}

	return &g
}

func (g Grid) SetContext(debugContext string) {
	g.mDebugContext = debugContext
}

func (g Grid) String() string {
	m := NewJSMap()
	m.Put("context", g.mDebugContext)
	m.Put("# cells", g.mCells.Size())
	m.Put("ColumnSizes", JSListWith(g.ColumnSizes))
	return m.String()
}

func (g Grid) DebugContext() string {
	return g.mDebugContext
}

func (g Grid) SetWidget(widget Widget) {
	g.mWidget = widget
}
func (g Grid) Widget() Widget {
	return g.mWidget
}

func (g Grid) NumColumns() int {
	return len(g.ColumnSizes)
}

func (g Grid) NextCellLocation() IPoint {
	if !g.mNextCellKnown {
		Pr("Calculating next cell location, # cells:", g.mCells.Size())
		x := 0
		y := 0
		if g.mCells.NonEmpty() {

			lastCell := g.mCells.Last()

			x = lastCell.X + lastCell.Width
			y = lastCell.Y
			CheckState(x <= g.NumColumns())

			Pr("end of last cell:", x, y)
			if x == g.NumColumns() {
				x = 0
				y += 1
				Pr("bumped to next row")
			}
		}
		g.mCachedNextCellLocation = IPointWith(x, y)
		Pr("next cell loc:", x, y)
		g.mNextCellKnown = true
	}
	return g.mCachedNextCellLocation
}

func (g Grid) NumRows() int {
	nextLoc := g.NextCellLocation()
	y := nextLoc.Y
	if nextLoc.X > 0 {
		y++
	}
	return y
}

func (g Grid) checkValidColumn(x int) int {
	if x < 0 || x >= g.NumColumns() {
		BadArg("not a valid column:", x)
	}
	return x
}

func (g Grid) checkValidRow(y int) int {
	if y < 0 || y >= g.NumRows() {
		BadArg("not a valid row:", y)
	}
	return y
}

func (g Grid) cellAt(x int, y int) GridCell {
	return g.mCells.Get(g.checkValidRow(y)*g.NumColumns() + g.checkValidColumn(x))
}

func (g Grid) AddCell(cell GridCell) {
	g.mCells.Add(cell)
	g.mNextCellKnown = false
}

/**
 * Get list of cells... must be considered READ ONLY
 */
func (g Grid) Cells() *Array[GridCell] {
	return g.mCells
}

func (g Grid) PropagateGrowFlags() {

	cs := g.Cells().Size()
	var colGrowFlags = make([]int, cs)
	var rowGrowFlags = make([]int, cs)

	for _, cell := range g.Cells().Array() {
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
	for _, cell := range g.Cells().Array() {

		if cell.IsEmpty() {
			continue
		}

		for x := cell.X; x < cell.X+cell.Width; x++ {
			cell.GrowX = MaxInt(cell.GrowX, colGrowFlags[x])
		}
		cell.GrowY = rowGrowFlags[cell.Y]
	}
}
