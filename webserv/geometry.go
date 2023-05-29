package webserv

type IPoint struct {
	X int
	Y int
}

func IPointWith(x int, y int) IPoint {
	return IPoint{X: x, Y: y}
}
