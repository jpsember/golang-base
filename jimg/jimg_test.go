package jimg_test

import (
	. "github.com/jpsember/golang-base/base"
	"github.com/jpsember/golang-base/jimg"
	"github.com/jpsember/golang-base/jt"
	"golang.org/x/image/draw"
	"image"
	"testing"
)

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

func TestImageFit(t *testing.T) {
	j := jt.Newz(t)

	sourceSize := IPointWith(689, 694)
	targetSize := IPointWith(100, 200)

	tf := jimg.NewImageFit()
	tf.Strategy = jimg.LETTERBOX
	tf.TargetSize = targetSize

	tf.WithSourceSize(sourceSize)
	j.AssertMessage(tf.TargetRect())
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
	dstBounds := rect(0, 0, dogSize.X/3, dogSize.Y/3)
	dst := image.NewRGBA(dstBounds)
	ourdst := RectWithImageRect(dstBounds)

	if false {
		// Draw without scaling, but with appropriate cropping
		draw.Draw(dst, dstBounds, srcImage.Image(), pt(50, 100), draw.Src)
	} else {

		fit := jimg.NewImageFit()
		fit.Strategy = jimg.CROP
		fit.TargetSize = ourdst.Size
		fit.WithSourceSize(srcSize)
		r := fit.TargetRect()
		Pr("target size:", ourdst.Size)

		//do unit test on TestImageFit()
		Pr("target rect:", r)
		// Draw with scaling (and appropriate cropping?)

		sr := rect(0, 0, srcSize.X, srcSize.Y)
		draw.ApproxBiLinear.Scale(dst, r.ToImageRectangle(), srcImage.Image(), sr, draw.Over, nil)
	}

	dstImage := jimg.JImageOf(dst)
	writeImg(dstImage, "_SKIP_"+t.Name()+".png")

	//func Draw(dst Image, r image.Rectangle, src image.Image, sp image.Point, op Op)
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
