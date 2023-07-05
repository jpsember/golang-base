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

var IPointZero = IPoint{}

func (p IPoint) String() string {
	return JSListWith([]int{p.X, p.Y}).CompactString()
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

func (r *Rect) String() string {
	return JSListWith([]int{r.Location.X, r.Location.Y, r.Size.W, r.Size.H}).CompactString()
}

func RectWith(x int, y int, w int, h int) Rect {
	r := Rect{
		Location: IPoint{X: x, Y: y},
		Size:     Size{W: w, H: h},
	}
	return r
}
