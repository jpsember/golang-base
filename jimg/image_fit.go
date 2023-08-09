package jimg

import (
	. "github.com/jpsember/golang-base/base"
)

type ImageFitStruct struct {
	TargetSize       IPoint
	SourceSize       IPoint
	targetRectangle  Rect
	FactorCropH      float64
	FactorCropV      float64
	FactorPadH       float64
	FactorPadV       float64
	ScaleFactor      float64
	ScaledSourceRect Rect
}

type ImageFit = *ImageFitStruct

func NewImageFit() ImageFit {
	t := &ImageFitStruct{
		FactorCropH: .66,
		FactorCropV: .66,
		FactorPadH:  .33,
		FactorPadV:  .33,
	}
	return t
}

func (m ImageFit) Optimize() {
	c := m.FactorCropH
	d := m.FactorPadV
	e := m.FactorPadH
	f := m.FactorCropV

	u := float64(m.TargetSize.X)
	v := float64(m.TargetSize.Y)
	w := float64(m.SourceSize.X)
	h := float64(m.SourceSize.Y)

	pr := PrIf(true)

	pr("Crop/Pad factors:", c, d, e, f)

	var s float64
	targetAspect := aspectRatioFromSize(m.TargetSize)
	sourceAspect := aspectRatioFromSize(m.SourceSize)
	if targetAspect > sourceAspect {
		pr("u:", u, "c:", c, "d:", d, "w:", w)
		if c <= 0 {
			s = v / h
		} else {
			s = (u * (c - d)) / (2 * c * w)
		}
	} else {
		if f <= 0 {
			s = u / w
		} else {
			s = (v * (e + f)) / (2 * f * h)
		}
	}
	pr("sourceAsp:", sourceAspect)
	pr("targetAsp:", targetAspect)
	pr("scale factor:", s)

	m.ScaleFactor = s

	sx := (u - (s * w)) / 2
	sy := (v - (s * h)) / 2
	sw := sx + s*w
	sh := sy + s*h
	m.ScaledSourceRect = RectWithFloat(sx, sy, sw, sh)
	pr("scaled source rect:", m.ScaledSourceRect)
}

func (m ImageFit) TargetRect() Rect {
	return m.targetRectangle.AssertValid()
}

func aspectRatioFromSize(size IPoint) float64 {
	return aspectRatio(size.X, size.Y)
}

func aspectRatio(width int, height int) float64 {
	CheckArg(width > 0 && height > 0)
	return float64(height) / float64(width)
}
