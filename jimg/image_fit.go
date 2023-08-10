package jimg

import (
	. "github.com/jpsember/golang-base/base"
)

type ImageFitStruct struct {
	TargetSize        IPoint
	SourceSize        IPoint
	targetRectangle   Rect
	FactorCropH       float64
	FactorCropV       float64
	FactorPadH        float64
	FactorPadV        float64
	ScaleFactor       float64
	ScaledSourceRect  Rect
	sourceAspectRatio float64
	targetAspectRatio float64
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

func (m ImageFit) evaluateCost(s float64) float64 {
	c := m.FactorCropH
	d := m.FactorPadV
	e := m.FactorPadH
	f := m.FactorCropV

	u := float64(m.TargetSize.X)
	v := float64(m.TargetSize.Y)
	w := float64(m.SourceSize.X)
	h := float64(m.SourceSize.Y)
	if m.targetAspectRatio > m.sourceAspectRatio {
		return c*(s*s*w*h-u*s*h) + d*(u*v-u*s*h)
	} else {
		return e*(u*v-2*w*v) + f*(s*s*w*h-s*w*v)
	}
}

//
//func (m ImageFit) evaluateCostTargetAspGreater(s float64) float64 {
//	c := m.FactorCropH
//	d := m.FactorPadV
//	//e := m.FactorPadH
//	//f := m.FactorCropV
//
//	u := float64(m.TargetSize.X)
//	v := float64(m.TargetSize.Y)
//	w := float64(m.SourceSize.X)
//	h := float64(m.SourceSize.Y)
//
//	return c*(s*s*w*h-u*s*h) + d*(u*v-u*s*h)
//}
//
//func (m ImageFit) evaluateCostTargetAspLess(s float64) float64 {
//	//c := m.FactorCropH
//	//d := m.FactorPadV
//	e := m.FactorPadH
//	f := m.FactorCropV
//
//	u := float64(m.TargetSize.X)
//	v := float64(m.TargetSize.Y)
//	w := float64(m.SourceSize.X)
//	h := float64(m.SourceSize.Y)
//
//	return e*(u*v-2*w*v) + f*(s*s*w*h-s*w*v)
//}

func (m ImageFit) evaluateNear(s float64, fn func(s float64) float64) {
	Halt("hey ho")
	for i := 0; i < 20; i++ {
		s2 := s * (1.0 + float64(i-10)*0.05)
		c := m.evaluateCost(s2)
		Pr("s:", s2, "cost:", c)
	}
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

	pr := PrIf(false)

	pr("Crop/Pad factors:", c, d, e, f)

	var s float64
	m.targetAspectRatio = aspectRatioFromSize(m.TargetSize)
	m.sourceAspectRatio = aspectRatioFromSize(m.SourceSize)

	if m.targetAspectRatio > m.sourceAspectRatio {

		// The optimal scale is somewhere between u/w and v/h
		scaleMin := u / w
		scaleMax := v / h

		pr("u:", u, "c:", c, "d:", d, "w:", w)
		if c <= 0 {
			s = scaleMax
		} else {
			s = (u * (c - d)) / (2 * c * w)
			if s < scaleMin || s > scaleMax {
				c1 := m.evaluateCost(scaleMin)
				c2 := m.evaluateCost(scaleMax)
				pr("cost for scaleMin:", scaleMin, "is:", c1)
				pr("cost for scaleMax:", scaleMax, "is:", c2)
				if c1 < c2 {
					s = scaleMin
				} else {
					s = scaleMax
				}
			} else {
				m.evaluateNear(s, m.evaluateCost)
			}

		}
	} else {
		pr("v:", v, "e:", e, "f:", f, "h:", h)

		// The optimal scale is somewhere between v/h and u/w
		scaleMin := v / h
		scaleMax := u / w

		if f <= 0 {
			s = scaleMax
		} else {
			s = (v * (e + f)) / (2 * f * h)

			if s < scaleMin || s > scaleMax {
				c1 := m.evaluateCost(scaleMin)
				c2 := m.evaluateCost(scaleMax)
				pr("cost for scaleMin:", scaleMin, "is:", c1)
				pr("cost for scaleMax:", scaleMax, "is:", c2)
				if c1 < c2 {
					s = scaleMin
				} else {
					s = scaleMax
				}
			} else {
				m.evaluateNear(s, m.evaluateCost)
			}
		}
		pr("s:", s)
	}
	pr("sourceAsp:", m.sourceAspectRatio)
	pr("targetAsp:", m.targetAspectRatio)
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
