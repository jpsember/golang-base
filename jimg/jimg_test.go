package jimg_test

import (
	. "github.com/jpsember/golang-base/base"
	"github.com/jpsember/golang-base/jimg"
	"github.com/jpsember/golang-base/jt"
	"golang.org/x/image/draw"
	"image"
	"testing"
)

func TestColorStuff(t *testing.T) {
	j := jt.New(t)
	img := readImage("resources/balloons.jpg")
	img = CheckOkWith(img.AsDefaultType())
	j.AssertMessage(img.ToJson())
}

func TestReadJpg(t *testing.T) {
	j := jt.New(t)
	img := readYCbCrImage()
	j.AssertMessage(img.ToJson())
}

func TestConvertImageFormat(t *testing.T) {
	j := jt.New(t)
	img := readYCbCrImage()
	img2 := CheckOkWith(img.AsType(jimg.TypeNRGBA))
	j.AssertMessage(img2.ToJson())
}

func imageFit(j jt.JTest, sourceSize IPoint, targetSize IPoint) {

	mp := NewJSMap()
	mp.PutNumberedKey("Source size", sourceSize)
	mp.PutNumberedKey("Target size", targetSize)

	for i := 0; i <= 100; i += 10 {
		t := float64(i) / 100.0

		m2 := NewJSMap()
		m2.PutNumberedKey("t", t)

		scaleFactor, scaledRect := jimg.FitRectToRect(RectWithLocationAndSize(IPointZero, sourceSize), RectWithLocationAndSize(IPointZero, targetSize),
			t)

		m2.PutNumberedKey("scale", scaleFactor)
		m2.PutNumberedKey("scaled rect", scaledRect.ToJson())
		mp.PutNumbered(m2)
	}
	j.AssertMessage(mp.String())
}

func TestImageFitPortraitToLandscape(t *testing.T) {
	j := jt.Newz(t)
	imageFit(j, IPointWith(1800, 2400), IPointWith(600, 500))
}

func TestImageFitLandscapeToPortrait(t *testing.T) {
	j := jt.New(t)
	imageFit(j, IPointWith(1000, 600), IPointWith(500, 1200))
}

func TestImageFitEqual(t *testing.T) {
	j := jt.New(t)
	imageFit(j, IPointWith(1000, 600), IPointWith(1000, 600))
}

func TestImageFitSimilar(t *testing.T) {
	j := jt.New(t)
	imageFit(j, IPointWith(1000, 600), IPointWith(200, 120))
}

func pt(x int, y int) image.Point {
	return image.Point{X: x, Y: y}
}

func rect(x int, y int, w int, h int) image.Rectangle {
	return image.Rectangle{
		Min: pt(x, y), Max: pt(x+w, y+h)}
}

func TestPlotIntoImage(t *testing.T) {
	j := jt.Newz(t)
	_ = j

	srcImage := readImage("resources/0.jpg")
	srcImage = CheckOkWith(srcImage.AsType(jimg.TypeNRGBA))
	srcSize := srcImage.Size()
	Pr(srcSize)

	dogSize := pt(689, 694)
	_ = dogSize
	dstBounds := rect(0, 0, dogSize.X, 232)
	dst := image.NewRGBA(dstBounds)
	_ = dst.Pix
	ourdst := RectWithImageRect(dstBounds)

	Todo("It leaves an alpha channel which is a bit misleading...")
	Todo("Feature to convert alpha pixels to purple or something")

	_, r := jimg.FitRectToRect(RectWithSize(srcSize), ourdst, 1.0)

	jimg.SetPurple(dst)
	Todo("strange black band")

	Pr("target size:", ourdst.Size)

	//do unit test on TestImageFit()
	Pr("scaled source rect:", r)

	// Draw with scaling (and appropriate cropping?)

	sr := rect(0, 0, srcSize.X, srcSize.Y)
	Todo("investigate Over vs Src")

	tr := r.ToImageRectangle()
	Pr("target rect end:", tr.Size().Y)
	Pr("src rect end   :", sr.Size().Y)

	Pr("target rect:", tr)
	Pr("source rect:", sr)

	draw.BiLinear.Scale(dst, tr, srcImage.Image(), sr, draw.Over, nil)
	//draw.ApproxBiLinear.Scale(dst, tr, srcImage.Image(), sr, draw.Over, nil)

	dstImage := jimg.JImageOf(dst)
	writeImg(dstImage, "_SKIP_"+t.Name()+".png")
}

func readImage(filename string) jimg.JImage {
	p := NewPathM(filename)
	return CheckOkWith(jimg.DecodeImage(p.ReadBytesM()))
}

func readYCbCrImage() jimg.JImage {
	return readImage("resources/balloons.jpg")
}

func writeImg(img jimg.JImage, filename string) {
	p := NewPathM(filename)
	by := CheckOkWith(img.EncodePNG())
	p.WriteBytesM(by)
}
