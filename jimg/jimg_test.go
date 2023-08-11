package jimg_test

import (
	"fmt"
	. "github.com/jpsember/golang-base/base"
	"github.com/jpsember/golang-base/jimg"
	"github.com/jpsember/golang-base/jt"
	"golang.org/x/image/draw"
	"image"
	"strings"
	"testing"
)

func TestColorStuff(t *testing.T) {
	j := jt.New(t)
	j.AssertMessage(readYCbCrImage().ToJson())
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

func TestImageFitPortraitToLandscape(t *testing.T) {
	j := jt.New(t)
	auxPlotIntoImage(j, "portrait.jpg", IPointWith(600, 500), 0, 1, .5, .5, .5, .5)
}

func TestImageFitLandscapeToPortrait(t *testing.T) {
	j := jt.New(t)
	auxPlotIntoImage(j, "landscape.jpg", IPointWith(500, 1200), 0, 1, .5, .5, .5, .5)
}

func TestImageFitEqual(t *testing.T) {
	j := jt.New(t)
	auxPlotIntoImage(j, "landscape.jpg", IPointWith(500, 400), .5, .5, .5, .5, .5, .5)
}

func TestImageFitSimilar(t *testing.T) {
	j := jt.New(t)
	auxPlotIntoImage(j, "landscape.jpg", IPointWith(750, 600), .5, .5, .5, .5, .5, .5)
}

func TestImageFitPadVsCrop(t *testing.T) {
	j := jt.New(t)
	auxPlotIntoImage(j, "landscape.jpg", IPointWith(300, 800), 0, 1, .5, .5, .5, .5)
}

func TestImageFitHorzBias(t *testing.T) {
	j := jt.New(t)
	auxPlotIntoImage(j, "landscape.jpg", IPointWith(500, 1200),
		1, 1, -1, 1, 0, 0)
}

func TestImageFitVertBias(t *testing.T) {
	j := jt.New(t)
	auxPlotIntoImage(j, "portrait.jpg", IPointWith(1200, 500),
		1, 1, 0, 0, -1, 1)
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

func interp(t float64, v0 float64, v1 float64) float64 {
	return v1*t + (1-t)*v0
}

func fstr(value float64) string {
	return fmt.Sprintf("%5.2f", value)

}

func auxPlotIntoImage(j jt.JTest, imageName string, dstSize IPoint,
	padVsCropBiasMin float64, padVsCropBiasMax float64,
	horzBiasMin float64, horzBiasMax float64,
	vertBiasMin float64, vertBiasMax float64) {

	mp := NewJSMap()

	srcImage := CheckOkWith(readImage("resources/" + imageName).AsType(jimg.TypeNRGBA))
	srcSize := srcImage.Size()

	mp.Put("src size", srcSize)

	if j.Verbose() {
		deleteTemporaryImages()
	}

	for pass := 0; pass <= 4; pass++ {
		factor := float64(pass) / 4.0

		m2 := NewJSMap()
		mp.PutNumberedKey("pass", m2)

		padVsCrop := interp(factor, padVsCropBiasMin, padVsCropBiasMax)
		horzBias := interp(factor, horzBiasMin, horzBiasMax)
		vertBias := interp(factor, vertBiasMin, vertBiasMax)

		m2.Put("pad vs crop", fstr(padVsCrop))
		m2.Put("horz bias", fstr(horzBias))
		m2.Put("vert bias", fstr(vertBias))

		_, r := FitRectToRect(srcSize, dstSize, padVsCrop, horzBias, vertBias)
		m2.Put("target rect", r)

		if j.Verbose() {
			dst := image.NewNRGBA(RectWithSize(dstSize).ToImageRectangle())

			//sr := rect(0, 0, srcSize.X, srcSize.Y)
			Todo("investigate Over vs Src")

			tr := r.ToImageRectangle()
			draw.BiLinear.Scale(dst, tr, srcImage.Image(), srcImage.Image().Bounds(), draw.Over, nil)
			//draw.ApproxBiLinear.Scale(dst, tr, srcImage.Image(), sr, draw.Over, nil)

			dstImage := jimg.JImageOf(dst)
			dstImage.SetTransparentPurple()

			writeImg(dstImage, "_SKIP_"+strings.TrimPrefix(j.Name(), "Test")+"_"+IntToString(pass)+".png")
		}
	}
	j.AssertMessage(mp)
}

var imagesDeletedFlag bool

// Delete all png, jpg files in the current directory that have the prefix "_SKIP_".
func deleteTemporaryImages() {
	if imagesDeletedFlag {
		return
	}
	imagesDeletedFlag = true
	for _, f := range NewDirWalk(CurrentDirectory()).IncludeExtensions("png", "jpg").FilesRelative() {
		if strings.HasPrefix(f.String(), "_SKIP_") {
			f.DeleteFileM()
		}
	}
}
