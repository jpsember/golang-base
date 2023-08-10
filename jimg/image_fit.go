package jimg

import (
	. "github.com/jpsember/golang-base/base"
)

func FitRectToRect(src Rect, targ Rect, factor float64) (float64, Rect) {
	Todo("better to assume origins of rects are at 0,0?")
	src.AssertValid()
	targ.AssertValid()

	srcWidth := float64(src.Size.X)
	srcHeight := float64(src.Size.Y)
	targWidth := float64(targ.Size.X)
	targHeight := float64(targ.Size.Y)

	srcAspect := src.AspectRatio()
	targAspect := targ.AspectRatio()
	scaleMin := targWidth / srcWidth
	scaleMax := targHeight / srcHeight
	if targAspect < srcAspect {
		factor = 1 - factor
	}

	scale := (1-factor)*scaleMin + factor*scaleMax

	scaledWidth := scale * srcWidth
	scaledHeight := scale * srcHeight

	resultRect := RectWithFloat(
		float64(targ.Location.X)+(targWidth-scaledWidth)/2,
		float64(targ.Location.Y)+(targHeight-scaledHeight)/2,
		scaledWidth, scaledHeight)

	return scale, resultRect
}
