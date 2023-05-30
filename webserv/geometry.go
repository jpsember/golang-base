package webserv

import "fmt"

type IPoint struct {
	X int
	Y int
}

func IPointWith(x int, y int) IPoint {
	return IPoint{X: x, Y: y}
}

type Size struct {
	W int
	H int
}

type Rect struct {
	Location IPoint
	Size     Size
}

func (r *Rect) IsValid() bool {
	return r.Size.W > 0 && r.Size.H > 0
}

func (r Rect) String() string {
	return fmt.Sprintf("[x:%v y:%v w:%v h:%v]", r.Location.X, r.Location.Y, r.Size.W, r.Size.H)
}

func RectWith(x int, y int, w int, h int) Rect {
	r := Rect{
		Location: IPoint{X: x, Y: y},
		Size:     Size{W: w, H: h},
	}
	return r
}
