package webapp

import (
	. "github.com/jpsember/golang-base/base"
	"github.com/jpsember/golang-base/jimg"
)

type DemoPhotosStruct struct {
	scaledPhotoNames []string
	scaledPhotoDir   Path
}

type DemoPhotos = *DemoPhotosStruct

func newDemoPhotos() DemoPhotos {
	t := &DemoPhotosStruct{}
	return t
}

var SharedDemoPhotos DemoPhotos = newDemoPhotos()

func (d DemoPhotos) ScaledPhotoNames() []string {
	if d.scaledPhotoNames == nil {
		w := NewDirWalk(d.scaledPhotosDir()).IncludeExtensions("jpg")
		var result []string
		for _, x := range w.FilesRelative() {
			result = append(result, x.AsNonEmptyString())
		}
		d.scaledPhotoNames = result
	}
	return d.scaledPhotoNames
}

func (d DemoPhotos) scaledPhotosDir() Path {
	if d.scaledPhotoDir.Empty() {
		dirTarget := FindProjectDirM().JoinM("sample_photos_scaled")
		dirTarget.MkDirsM()
		d.scaledPhotoDir = dirTarget
	}
	return d.scaledPhotoDir
}

// Read all images (jpeg, png) in project_config/sample_photos, scale to our standard size,
// and write to project_config/sample_photos_scaled.
func (d DemoPhotos) ReadSamples() {
	dir := FindProjectDirM().JoinM("sample_photos")
	if !dir.Exists() {
		Alert("!No such directory:", dir)
		return
	}

	w := NewDirWalk(dir).IncludeExtensions("jpg", "jpeg", "png")

	for _, f := range w.Files() {

		nm := f.TrimExtension().Base()
		targetFile := d.scaledPhotosDir().JoinM(nm + ".jpg")
		if targetFile.Exists() {
			continue
		}

		var err error
		var img jimg.JImage

		for {
			img, err = jimg.DecodeImage(f.ReadBytesM())
			if err != nil {
				break
			}
			img, err = img.AsDefaultType()
			if err != nil {
				break
			}

			// Scale the image to our desired portrait size.
			// Later we will support different sizes.
			targetSize := IPointWith(800, 1200)
			scaled := img.AsDefaultTypeM().ScaleToSize(targetSize)

			targetFile.WriteBytesM(CheckOkWith(scaled.ToJPEG()))
			break
		}
		if err != nil {
			Alert("Trouble decoding:", f.Base(), err, INDENT, img)
		}
	}
}
