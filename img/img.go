package img

import (
	"bytes"
	"image"
	"image/draw"
	"image/png"
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
	size      IPoint
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

	Todo("discard coordinate system, e.g., bounds not at zero?")
	CheckNotNil(image)
	t := &JImageStruct{
		image: image,
	}
	return t
}

func (ji JImage) Image() image.Image {
	return ji.image
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

func (ji JImage) Size() IPoint {
	if ji.size == IPointZero {
		b := ji.image.Bounds()
		ji.size = IPointWith(b.Dx(), b.Dy())
	}
	return ji.size
}

func (ji JImage) ToJson() JSMap {
	m := NewJSMap()
	m.Put("", "JImage")
	m.Put("type", int(ji.Type()))
	m.Put("size", ji.Size())
	return m
}

func GetImageInfo(image image.Image) JSMap {
	ji := JImageOf(image)
	return ji.ToJson()
}

func (ji JImage) AsType(desiredType JImageType) (JImage, error) {
	var result JImage
	errstring := "unsupported image type"
	if ji.Type() == desiredType {
		result = ji
	} else {
		var m draw.Image
		switch desiredType {
		case TypeNRGBA:
			m = image.NewNRGBA(image.Rect(0, 0, ji.Size().X, ji.Size().Y))
		}
		if m != nil {
			draw.Draw(m, m.Bounds(), ji.Image(), image.Point{}, draw.Src)
			result = JImageOf(m)
		}
	}
	if result == nil {
		return nil, Error(errstring)
	} else {
		return result, nil
	}
}

func (ji JImage) ToPNG() ([]byte, error) {
	if ji.Type() != TypeNRGBA {
		return nil, Error("Cannot convert to PNG", ji.ToJson())
	}

	var bb bytes.Buffer
	err := png.Encode(&bb, ji.Image())
	Todo("wrap an error here?")
	return bb.Bytes(), err
}
