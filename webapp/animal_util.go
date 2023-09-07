package webapp

import (
	. "github.com/jpsember/golang-base/base"
	"github.com/jpsember/golang-base/jimg"
	. "github.com/jpsember/golang-base/webapp/gen/webapp_data"
	"math/rand"
)

func RandomAnimal(r *rand.Rand) AnimalBuilder {
	if r == nil {
		r = NewJSRand().Rand()
	}
	a := NewAnimal()
	a.SetName(RandomText(r, 20, false))
	a.SetSummary(RandomText(r, Ternary(false, 300, 20), false))
	a.SetDetails(RandomText(r, Ternary(false, 2000, 20), true))
	a.SetCampaignTarget(int((r.Intn(10) + 2) * 50 * DollarsToCurrency))
	a.SetCampaignBalance(r.Intn(a.CampaignTarget()))

	{
		// Copy one of our sample photos to use as this animal's photo
		d := SharedDemoPhotos
		nms := d.ScaledPhotoNames()
		if len(nms) == 0 {
			Alert("?No demo photos found; animal(s) won't have photos")
		} else {
			i := r.Intn(len(nms))
			b := NewBlob()
			pth := d.scaledPhotosDir().JoinM(nms[i])
			b.SetData(pth.ReadBytesM())
			AssignBlobName(b)
			b2 := CheckOkWith(CreateBlob(b))
			a.SetPhotoThumbnail(b2.Id())
		}
	}
	Todo("Issue #59: add random photo")
	return a
}

func HasAnimals() bool {
	return CheckOkWith(ReadAnimal(1)) != DefaultAnimal
}

func GenerateRandomAnimals() {
	Alert("Generating some random animals")
	rnd := NewJSRand()
	for i := 0; i < 30; i++ {
		anim := RandomAnimal(rnd.Rand())
		CreateAnimal(anim)
		Pr("added animal:", INDENT, anim)
	}
}

type DemoPhotosStruct struct {
	scaledPhotoNames []string
	scaledPhotoDir   Path
}

type DemoPhotos = *DemoPhotosStruct

func newDemoPhotos() DemoPhotos {
	t := &DemoPhotosStruct{}
	t.init()
	return t
}

var SharedDemoPhotos DemoPhotos = newDemoPhotos()

func (d DemoPhotos) init() {

}

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
			scaled := ScaleImageToSize(img, targetSize)

			targetFile.WriteBytesM(CheckOkWith(scaled.ToJPEG()))
			break
		}
		if err != nil {
			Alert("Trouble decoding:", f.Base(), err, INDENT, img)
		}
	}
}

func ScaleImageToSize(img jimg.JImage, targetSize IPoint) jimg.JImage {
	img = img.AsDefaultTypeM()
	_, targetRect := FitRectToRect(img.Size(), targetSize, 1.0, 0, -.5)
	return img.ScaledToRect(targetSize, targetRect)
}

func AssignBlobName(b BlobBuilder) {
	if b.Name() == "" {
		b.SetName(string(GenerateBlobId()))
	}
}

func TrimBlob(b Blob) Blob {
	CheckNotNil(b)
	if len(b.Data()) > 50 {
		b2 := b.Build().ToBuilder()
		b2.SetData(DefaultBlob.Data())
		b = b2.Build()
	}
	return b
}
