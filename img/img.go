package img

import (
	"bytes"
	"image"
)

import (
	. "github.com/jpsember/golang-base/base"
	_ "image/jpeg"
	// Package image/jpeg is not used explicitly in the code below,
	// but is imported for its initialization side-effect, which allows
	// image.Decode to understand JPEG formatted images. Uncomment these
	// two lines to also understand GIF and PNG images:
	// _ "image/gif"
	_ "image/png"
)

func DecodeImage(imgbytes []byte) (JImage, error) {

	img, format, err := image.Decode(bytes.NewReader(imgbytes))
	var jmg JImage
	if err == nil {
		jmg = JImageOf(img)
	}
	Pr("format:", format)
	return jmg, err
}

type JImageStruct struct {
	image     image.Image
	imageType JImageType
}

type JImage = *JImageStruct

type JImageType int

const (
	typeUnitialized JImageType = iota
	TypeRGBA
	TypeNRGBA
	TypeCMYK
	TypeYCbCr
	TypeUnknown = -1
)

func JImageOf(image image.Image) JImage {
	CheckNotNil(image)
	t := &JImageStruct{
		image: image,
	}
	return t
}

func (ji JImage) Type() JImageType {

	if ji.imageType == typeUnitialized {
		ty := ji.imageType
		switch ji.image.(type) {
		case *image.RGBA:
			ty = TypeRGBA
		case *image.NRGBA:
			ty = TypeNRGBA
		case *image.CMYK:
			ty = TypeCMYK
		case *image.YCbCr:
			ty = TypeYCbCr
		default:
			Pr("Color model:", ji.image.ColorModel())
			ty = TypeUnknown
		}
		ji.imageType = ty
	}
	return ji.imageType
}

func (ji JImage) ToJson() JSMap {
	m := NewJSMap()
	m.Put("", "JImage")
	m.Put("type", int(ji.Type()))
	return m
}
