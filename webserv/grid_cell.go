package webserv

import . "github.com/jpsember/golang-base/json"

type GridCellObj struct {
	Widget Widget
	X      int
	Y      int
	Width  int
	GrowX  int
	GrowY  int
}
type GridCell = *GridCellObj

func NewGridCell() GridCell {
	return &GridCellObj{}
}
func (g GridCell) IsEmpty() bool {
	return g.Widget == nil
}

func (g GridCell) String() string {
	m := NewJSMap()
	m.Put("", "GridCell").Put("X", g.X).Put("Y", g.Y).Put("Width", g.Width).Put("GrowX", g.GrowX).Put("GrowY", g.GrowY)
	if g.Widget != nil {
		m.Put("widget", g.Widget.GetId())
	}
	return m.AsString()
}
