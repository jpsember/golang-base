package webserv

type GridCellObj struct {
	view  Widget
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
