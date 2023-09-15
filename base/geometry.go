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
var DefaultIPoint = IPointZero

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
	return p.Yf() / p.Xf()
}

func (p IPoint) Xf() float64 {
	return float64(p.X)
}

func (p IPoint) Yf() float64 {
	return float64(p.Y)
}

func (p IPoint) ScaledBy(factor float64) IPoint {
	return IPointWithFloat(p.Xf()*factor, p.Yf()*factor)
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
// The padVsCrop ranges from 0: maximum padding to 1: maximum cropping.
// The horzBias and vertBias take effect if the dimension is being cropped, and ranges from -1 ... 1,
// where 0 causes the image to be centered along that dimension.
// A vertBias of e.g. -.25 favors cropping lower parts of the image (where Y axis points down), the intuition
// being that a portrait input will have a person's face in the upper part.
// Returns the scaling factor applied, and the bounds of the scaled source image within the target
// rectangle's coordinate system.  Both source and target rectangles are assumed to have origins at 0,0.
func FitRectToRect(srcSize IPoint, targSize IPoint, padVsCropBias float64, horzBias float64, vertBias float64) (float64, Rect) {
	srcSize.AssertPositive()
	targSize.AssertPositive()

	Todo("!Have an FPoint class for this")
	srcWidth := srcSize.Xf()
	srcHeight := srcSize.Yf()
	targWidth := targSize.Xf()
	targHeight := targSize.Yf()

	srcAspect := srcSize.AspectRatio()
	targAspect := targSize.AspectRatio()
	scaleMin := targWidth / srcWidth
	scaleMax := targHeight / srcHeight
	if targAspect < srcAspect {
		padVsCropBias = 1 - padVsCropBias
	}

	scale := (1-padVsCropBias)*scaleMin + padVsCropBias*scaleMax

	scaledWidth := scale * srcWidth
	scaledHeight := scale * srcHeight

	cropWidth := scaledWidth - targWidth
	cropHeight := scaledHeight - targHeight

	cropBiasHorzValue := 0.0
	cropBiasVertValue := 0.0

	if cropWidth > 0 {
		cropBiasHorzValue = horzBias
	}

	if cropHeight > 0 {
		cropBiasVertValue = vertBias
	}

	resultRect := RectWithFloat(
		-cropWidth*((cropBiasHorzValue*.5)+.5),
		-cropHeight*((cropBiasVertValue*.5)+.5),
		scaledWidth, scaledHeight)

	return scale, resultRect
}
