package base

import (
	"image"
	"math"
)

// ------------------------------------------------------------------------------------
// 2D points (int-valued)
// ------------------------------------------------------------------------------------

type IPoint struct {
	X int
	Y int
}

func (p IPoint) Product() int {
	return p.X * p.Y
}

func (p IPoint) String() string {
	return p.ToJson().AsJSList().CompactString()
}

func (p IPoint) ToJson() JSEntity {
	return NewJSList().Add(p.X).Add(p.Y)
}

func (p IPoint) Parse(source JSEntity) DataClass {
	lst := source.AsJSList()
	x := lst.Get(0).AsInteger()
	y := lst.Get(1).AsInteger()
	return IPoint{
		X: int(x),
		Y: int(y),
	}
}

func IPointWith(x int, y int) IPoint {
	return IPoint{X: x, Y: y}
}

func IPointWithFloat(x float64, y float64) IPoint {
	return IPointWith(int(math.Round(x)), int(math.Round(y)))
}

var IPointZero = IPoint{}

func (p IPoint) IsPositive() bool {
	return p.X > 0 && p.Y > 0

}
func (p IPoint) AssertPositive() IPoint {
	if !p.IsPositive() {
		BadArg("<1IPoint coordinates are not both positive:", p)
	}
	return p
}

func (p IPoint) AspectRatio() float64 {
	return float64(p.Y) / float64(p.X)
}

// ------------------------------------------------------------------------------------
// Rectangle (int-valued)
// ------------------------------------------------------------------------------------

type Rect struct {
	Location IPoint
	Size     IPoint
}

func (r Rect) String() string {
	return r.ToJson().AsJSMap().String()
}

func (r Rect) ToJson() JSEntity {
	return NewJSMap().Put("loc", r.Location.ToJson()).Put("size", r.Size.ToJson())
}

func (r Rect) Parse(source JSEntity) DataClass {
	lst := source.AsJSMap()
	return Rect{
		Location: IPointZero.Parse(lst.GetMap("loc")).(IPoint),
		Size:     IPointZero.Parse(lst.GetMap("size")).(IPoint),
	}
}

func (r Rect) IsValid() bool {
	return r.Size.IsPositive()
}

func (r Rect) AssertValid() Rect {
	if !r.IsValid() {
		BadArg("<1Rect isn't valid:", INDENT, r)
	}
	return r
}

func RectWith(x int, y int, w int, h int) Rect {
	r := Rect{
		Location: IPoint{X: x, Y: y},
		Size:     IPoint{X: w, Y: h},
	}
	return r
}

func RectWithFloat(x float64, y float64, w float64, h float64) Rect {
	r := Rect{
		Location: IPointWithFloat(x, y),
		Size:     IPointWithFloat(w, h),
	}
	return r
}

var RectZero = Rect{}

func (r Rect) ToImageRectangle() image.Rectangle {
	return image.Rectangle{
		Min: image.Point{
			X: r.Location.X,
			Y: r.Location.Y,
		},
		Max: image.Point{
			X: r.Location.X + r.Size.X,
			Y: r.Location.Y + r.Size.Y,
		},
	}
}

func RectWithImageRect(src image.Rectangle) Rect {
	return RectWith(src.Min.X, src.Min.Y, src.Dx(), src.Dy())
}

func RectWithSize(size IPoint) Rect {
	return Rect{
		Size: size,
	}
}

func (r Rect) MoveBy(x int, y int) Rect {
	return RectWith(r.Location.X+x, r.Location.Y+y, r.Size.X, r.Size.Y)
}

func RectWithLocationAndSize(origin IPoint, size IPoint) Rect {
	return Rect{
		Location: origin,
		Size:     size,
	}
}

func (r Rect) MidPoint() IPoint {
	return IPointWith(r.Location.X+r.Size.X/2, r.Location.Y+r.Size.Y/2)
}

func (r Rect) AspectRatio() float64 {
	return r.Size.AspectRatio()
}

// Perform scaling, cropping, and/or padding to align a source rectangle to a target rectangle.
// The factor ranges from 0: maximum padding to 1: maximum cropping.  Returns the scaling factor
// applied, and the bounds of the scaled source image within the target rectangle's coordinate system.
// Both source and target rectangles are assumed to have their origins at 0,0.
func FitRectToRect(srcSize IPoint, targSize IPoint, factor float64) (float64, Rect) {
	srcSize.AssertPositive()
	targSize.AssertPositive()

	Todo("!Have an FPoint class for this")
	srcWidth := float64(srcSize.X)
	srcHeight := float64(srcSize.Y)
	targWidth := float64(targSize.X)
	targHeight := float64(targSize.Y)

	srcAspect := srcSize.AspectRatio()
	targAspect := targSize.AspectRatio()
	scaleMin := targWidth / srcWidth
	scaleMax := targHeight / srcHeight
	if targAspect < srcAspect {
		factor = 1 - factor
	}

	scale := (1-factor)*scaleMin + factor*scaleMax

	scaledWidth := scale * srcWidth
	scaledHeight := scale * srcHeight

	resultRect := RectWithFloat(
		(targWidth-scaledWidth)/2,
		(targHeight-scaledHeight)/2,
		scaledWidth, scaledHeight)

	return scale, resultRect
}
