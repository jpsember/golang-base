package img_test

import (
	. "github.com/jpsember/golang-base/base"
	"github.com/jpsember/golang-base/jt"
	"image"
	"os"
	"testing"

	// Package image/jpeg is not used explicitly in the code below,
	// but is imported for its initialization side-effect, which allows
	// image.Decode to understand JPEG formatted images. Uncomment these
	// two lines to also understand GIF and PNG images:
	// _ "image/gif"
	// _ "image/png"
	_ "image/jpeg"
)

func TestReadJpg(t *testing.T) {
	j := jt.New(t) // Use Newz to regenerate hash

	r := CheckOkWith(os.Open("resources/0.jpg"))
	defer r.Close()

	img, format, err := image.Decode(r)
	CheckOk(err)
	Pr("format:", format)
	Pr("bounds:", img.Bounds())
	j.AssertMessage("none")
}
