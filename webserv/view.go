package webserv

import (
	"fmt"
	. "github.com/jpsember/golang-base/base"
)

type Point struct {
	X int
	Y int
}

type Size struct {
	W int
	H int
}

type Rect struct {
	Location Point
	Size     Size
}

func (r Rect) String() string {
	return fmt.Sprintf("[x:%v y:%v w:%v h:%v]", r.Location.X, r.Location.Y, r.Size.W, r.Size.H)
}

type HtmlString struct {
	Source string
}

type ViewClass interface {
}

type View = *ViewStruct

type ViewStruct struct {
	Bounds   Rect
	Markup   string
	Children Array[*ViewStruct]
	Class    ViewClass
}

func NewView() View {
	v := ViewStruct{
		Markup: "hey",
	}
	return &v
}

func (v View) RenderView() {

}

func RectWith(x int, y int, w int, h int) Rect {
	r := Rect{
		Location: Point{X: x, Y: y},
		Size:     Size{W: w, H: h},
	}
	return r
}
