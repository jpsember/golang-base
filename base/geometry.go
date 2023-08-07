package base

// ------------------------------------------------------------------------------------
// 2D points (int-valued)
// ------------------------------------------------------------------------------------

type IPoint struct {
	X int
	Y int
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

var IPointZero = IPoint{}

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
	return r.Size.X > 0 && r.Size.Y > 0
}

func RectWith(x int, y int, w int, h int) Rect {
	r := Rect{
		Location: IPoint{X: x, Y: y},
		Size:     IPoint{X: w, Y: h},
	}
	return r
}

var RectZero = Rect{}
