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

func imageFit(j jt.JTest, sourceSize IPoint, targetSize IPoint) {
	tf := jimg.NewImageFit()
	tf.SourceSize = sourceSize
	tf.TargetSize = targetSize

	mp := NewJSMap()
	mp.PutNumberedKey("Source size", sourceSize)
	mp.PutNumberedKey("Target size", targetSize)

	for i := 0; i <= 100; i++ {
		j := float64(i) / 100.0

		cropFactor := (.28 * (1.0 - j)) + (.29 * j)
		//cropFactor := float64(i) / 100.0
		padFactor := 1.0 - cropFactor
		if cropFactor == padFactor {
			continue
		}
		m2 := NewJSMap()
		m2.PutNumberedKey("crop", cropFactor)
		m2.PutNumberedKey("pad", padFactor)

		tf.FactorCropH = cropFactor
		tf.FactorCropV = cropFactor
		tf.FactorPadH = padFactor
		tf.FactorPadV = padFactor

		tf.Optimize()

		m2.PutNumberedKey("scale", tf.ScaleFactor)
		m2.PutNumberedKey("scaled src", tf.ScaledSourceRect.ToJson())
		mp.PutNumbered(m2)
	}
	j.AssertMessage(mp.String())
}

func TestImageFit(t *testing.T) {
	j := jt.New(t)
	imageFit(j, IPointWith(1800, 2400), IPointWith(600, 500))
}

func TestImageFit2(t *testing.T) {
	j := jt.Newz(t)
	imageFit(j, IPointWith(1000, 600), IPointWith(2000, 3000))
}

func TestImageFit3(t *testing.T) {
	j := jt.Newz(t)

	sourceSize := IPointWith(2000, 1000)
	targetSize := IPointWith(600, 500)

	tf := jimg.NewImageFit()
	tf.SourceSize = sourceSize
	tf.TargetSize = targetSize

	mp := NewJSMap()
	mp.PutNumberedKey("Source size", sourceSize)
	mp.PutNumberedKey("Target size", targetSize)

	cropFactor := 0.1
	padFactor := 0.9

	tf.FactorCropH = cropFactor
	tf.FactorCropV = cropFactor

	tf.FactorPadH = padFactor
	tf.FactorPadV = padFactor

	mp.PutNumberedKey("crop", cropFactor)
	mp.PutNumberedKey("pad", padFactor)

	tf.Optimize()

	mp.PutNumberedKey("scale", tf.ScaleFactor)
	mp.PutNumberedKey("scaled src", tf.ScaledSourceRect.ToJson())
	j.AssertMessage(mp.String())
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
		fit.TargetSize = ourdst.Size
		fit.SourceSize = srcSize
		fit.Optimize()
		r := fit.ScaledSourceRect

		Pr("target size:", ourdst.Size)

		//do unit test on TestImageFit()
		Pr("scaled source rect:", r)
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
