package jimg

import (
	"bufio"
	"bytes"
	. "github.com/jpsember/golang-base/base"
	"golang.org/x/image/draw"
	"image"
	_ "image/jpeg"
	"image/png"
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
	if false {
		Pr("format:", format)
	}
	return jmg, err
}

// ------------------------------------------------------------------------------------
// Some notes about colors.
// ------------------------------------------------------------------------------------

// From  color.go:

// NRGBA represents a non-alpha-premultiplied 32-bit color.
//
// type NRGBA struct {
//	 R, G, B, A uint8
// }
//
// RGBA represents a traditional 32-bit alpha-premultiplied color, having 8
// bits for each of red, green, blue and alpha.
//
// An alpha-premultiplied color component C has been scaled by alpha (A), so
// has valid values 0 <= C <= A.
//
// So we want to be dealing with NRGBA colors, which are 8-bit per component,
// and unless alpha channel is to be used, its alpha values should all be FF.
//

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

var imageTypeToStringMap = map[JImageType]string{
	TypeNRGBA:   "NRGBA",
	TypeCMYK:    "CMYK",
	TypeYCbCr:   "YCbCr",
	TypeUnknown: "Unknown",
}

func ImageTypeStr(imgType JImageType) string {
	result := imageTypeToStringMap[imgType]
	if result == "" {
		result = "???"
	}
	return result
}

var zer = image.Point{}

func JImageOf(img image.Image) JImage {
	CheckNotNil(img)
	if img.Bounds().Min != zer {
		Pr("origin of image is not at (0,0);", img.Bounds())
	}
	t := &JImageStruct{
		image: img,
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
	m.Put("type", ImageTypeStr(ji.Type()))
	m.Put("size", ji.Size())
	if ji.Type() == TypeNRGBA {
		w := ji.NRGBA()
		m.Put("opaque", w.Opaque())
		Todo("What happens if we try to encode as jpeg with some transparent pixels?")
	}
	return m
}

func GetImageInfo(image image.Image) JSMap {
	ji := JImageOf(image)
	return ji.ToJson()
}

func (ji JImage) AsDefaultType() (JImage, error) {
	return ji.AsType(TypeNRGBA)
}

func (ji JImage) AsDefaultTypeM() JImage {
	return CheckOkWith(ji.AsType(TypeNRGBA))
}

func (ji JImage) NRGBA() *image.NRGBA {
	result, ok := ji.image.(*image.NRGBA)
	if !ok {
		BadArg("<1JImage does not contain an image.NRGBA:", TypeOf(ji.image))
	}
	return result
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
	if err != nil {
		Pr("Failed to encode image as PNG")
	}
	return bb.Bytes(), err
}

func (ji JImage) ScaledTo(size IPoint) JImage {

	var targetX, targetY int

	origSize := ji.Size()
	if size.X == 0 {
		if size.Y > 0 {
			targetY = size.Y
			targetX = MaxInt(1, (origSize.X*targetY)/origSize.Y)
		}
	} else {
		if size.X > 0 {
			targetX = size.X
			targetY = MaxInt(1, (origSize.Y*targetX)/origSize.X)
		}
	}
	CheckArg(targetX > 0 && targetY > 0, "Cannot scale image of size", ji.Size(), "to", size)
	scaledImage := image.NewNRGBA(image.Rect(0, 0, targetX, targetY))
	inputImage := ji.Image()
	draw.ApproxBiLinear.Scale(scaledImage, scaledImage.Bounds(), inputImage, inputImage.Bounds(), draw.Over, nil)
	return JImageOf(scaledImage)
}

func (ji JImage) EncodePNG() ([]byte, error) {
	var err error
	var result []byte
	for {
		byteBuffer := bytes.Buffer{}
		writer := bufio.NewWriter(&byteBuffer)
		err = png.Encode(writer, ji.Image())
		if err != nil {
			break
		}
		err = writer.Flush()
		if err == nil {
			result = byteBuffer.Bytes()
		}
		break
	}
	return result, err
}

var purple = []byte{
	0xc6, 0x64, 0xed, 0xff,
}

func (ji JImage) SetTransparentPurple() {
	img := ji.NRGBA()
	pix := img.Pix
	h := len(pix)
	for i := 0; i < h; i += 4 {
		if pix[i+3] != 0xff {
			for j, val := range purple {
				pix[i+j] = val
			}
		}
	}
}
