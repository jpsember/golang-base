package webserv

import (
	. "github.com/jpsember/golang-base/base"
)

type IPoint struct {
	X int
	Y int
}

func IPointWith(x int, y int) IPoint {
	return IPoint{X: x, Y: y}
}

type GridCellObj struct {
	view  any
	X     int
	Y     int
	Width int
	GrowX int
	GrowY int
}
type GridCell = *GridCellObj

func (g GridCell) IsEmpty() bool {
	return g.view == nil
}

type CellWeightList struct {
	weights []int
}

func NewCellWeightList() CellWeightList {
	return CellWeightList{
		weights: []int{},
	}
}

func (w CellWeightList) Set(index int, weight int) {
	w.GrowTo(1 + index)
	w.weights[index] = weight
}

func (w CellWeightList) GrowTo(size int) {
	for len(w.weights) < size {
		w.weights = append(w.weights, 0)
	}
}

func (w CellWeightList) Get(index int) int {
	if len(w.weights) <= index {
		return 0
	}
	return w.weights[index]
}

/**
 * Data for a View that contains a grid of child views
 */

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

/**
func (g Grid)  ( ) {

*/

func (g Grid) String() string {
	return ToString("Grid, context:", g.mDebugContext)

}
func (g Grid) DebugContext() string {
	return g.mDebugContext
}

//public <T extends Widget> T widget() {
//  return (T) mWidget;
//}

func (g Grid) SetWidget(widget Widget) {
	g.mWidget = widget
}

func (g Grid) NumColumns() int {
	return len(g.ColumnSizes)
}

func (g Grid) NextCellLocation() IPoint {
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
	if y < 0 || y >= g.NumColumns() {
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

	colGrowFlags := NewCellWeightList()
	rowGrowFlags := NewCellWeightList()

	for _, cell := range g.Cells().Array() {
		if cell.IsEmpty() {
			continue
		}

		// If view occupies multiple cells horizontally, don't propagate its grow flag
		if cell.GrowX > 0 && cell.Width == 1 {
			if colGrowFlags.Get(cell.X) < cell.GrowX {
				colGrowFlags.Set(cell.X, cell.GrowX)
			}
		}
		// If view occupies multiple cells vertically, don't propagate its grow flag
		// (at present, we don't support views stretching across multiple rows)
		if cell.GrowY > 0 {
			if rowGrowFlags.Get(cell.Y) < cell.GrowY {
				rowGrowFlags.Set(cell.Y, cell.GrowY)
			}
		}
	}

	// Now propagate grow flags from bit sets back to individual cells
	for _, cell := range g.Cells().Array() {

		if cell.IsEmpty() {
			continue
		}

		for x := cell.X; x < cell.X+cell.Width; x++ {
			cell.GrowX = MaxInt(cell.GrowX, colGrowFlags.Get(x))
		}
		cell.GrowY = rowGrowFlags.Get(cell.Y)
	}
}
