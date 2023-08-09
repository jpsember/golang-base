package jimg_test

import (
	. "github.com/jpsember/golang-base/base"
	"github.com/jpsember/golang-base/jimg"
	"github.com/jpsember/golang-base/jt"
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

	tf := jimg.NewImageFit()
	tf.Strategy = jimg.CROP
	tf.TargetSize = IPointWith(100, 200)

	tf.WithSourceSize(IPointWith(120, 202))
	j.AssertMessage(tf.TargetRect())
}

func readYCbCrImage() jimg.JImage {
	p := NewPathM("resources/balloons.jpg")
	bytes := p.ReadBytesM()
	img, err := jimg.DecodeImage(bytes)
	CheckOk(err)
	return img
}
