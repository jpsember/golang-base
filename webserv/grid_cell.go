package webserv

type GridCellObj struct {
	View  Widget
	X     int
	Y     int
	Width int
	GrowX int
	GrowY int
}
type GridCell = *GridCellObj

func NewGridCell() GridCell {
	return &GridCellObj{}
}
func (g GridCell) IsEmpty() bool {
	return g.View == nil
}
