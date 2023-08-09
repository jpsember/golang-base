package jimg

import (
	. "github.com/jpsember/golang-base/base"
)

type ImageFitStrategy int

const (
	CROP ImageFitStrategy = iota
	LETTERBOX
	HYBRID
)

type ImageFitStruct struct {
	TargetSize      IPoint
	Strategy        ImageFitStrategy
	targetRectangle Rect
}

type ImageFit = *ImageFitStruct

func NewImageFit() ImageFit {
	t := &ImageFitStruct{}
	return t
}

func (m ImageFit) WithSourceSize(sourceSize IPoint) ImageFit {
	Todo("This shouldn't be fluid, as it is the last thing called to invoke the calculation")
	sourceSize.AssertPositive()
	targetSize := m.TargetSize.AssertPositive()

	w := float64(sourceSize.X)
	h := float64(sourceSize.Y)
	u := float64(targetSize.X)
	v := float64(targetSize.Y)

	lambdaCrop := float64(1)
	lambdaLbox := float64(1)

	switch m.Strategy {
	default:
		BadArg("strategy:", m.Strategy)
	case CROP:
		lambdaLbox = 0
	case LETTERBOX:
		lambdaCrop = 0
	case HYBRID:
	}

	sourceAspect := h / w
	targetAspect := v / u
	if sourceAspect < targetAspect {
		temp := lambdaCrop
		lambdaCrop = lambdaLbox
		lambdaLbox = temp
	}
	Pr("aspect src:", sourceAspect, "target:", targetAspect)

	// I apply a cost function c as a function of the scale factor s:
	//
	//  c(s)   L_c(u - sw)^2 + L_l(v - sh)^2
	//
	// and take the derivative to find when c(s) is minimized, to yield optimal scale s*:
	//
	//  s* = L_c(wu) + L_l(hv)
	//       -------------------
	//       L_c(w^2) + L_l(h^2)
	//
	s := (lambdaCrop*w*u + lambdaLbox*h*v) / (lambdaCrop*w*w + lambdaLbox*h*h)

	resultWidth := s * w
	resultHeight := s * h

	m.targetRectangle = RectWithFloat((u-resultWidth)*.5, (v-resultHeight)*.5, resultWidth,
		resultHeight).AssertValid()
	Pr("targetRect:", m.targetRectangle)
	Pr("u,v:", u, v)
	Pr("w,h:", w, h)
	Pr("lambda crop:", lambdaCrop, "lbox:", lambdaLbox)
	Pr("resultW,H:", resultWidth, resultHeight)
	return m
}

func (m ImageFit) TargetRect() Rect {
	return m.targetRectangle.AssertValid()
}
