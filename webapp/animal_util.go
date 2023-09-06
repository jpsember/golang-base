package webapp

import (
	. "github.com/jpsember/golang-base/base"
	"github.com/jpsember/golang-base/jimg"
	. "github.com/jpsember/golang-base/webapp/gen/webapp_data"
	"github.com/jpsember/golang-base/webserv"
)

func RandomAnimal() AnimalBuilder {
	r := webserv.HTMLRand.Rand()
	a := NewAnimal()
	a.SetName(RandomText(r, 20, false))
	a.SetSummary(RandomText(r, Ternary(false, 300, 20), false))
	a.SetDetails(RandomText(r, Ternary(false, 2000, 20), true))
	a.SetCampaignTarget(int((r.Intn(10) + 2) * 50 * DollarsToCurrency))
	a.SetCampaignBalance(r.Intn(a.CampaignTarget()))
	Todo("Issue #59: add random photo")
	return a
}

func HasAnimals() bool {
	return CheckOkWith(ReadAnimal(1)) != DefaultAnimal
}

func GenerateRandomAnimals() {
	Alert("Generating some random animals")
	for i := 0; i < 100; i++ {
		anim := RandomAnimal()
		CreateAnimal(anim)
		Pr("added animal:", INDENT, anim)
	}
}

type DemoPhotosStruct struct {
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

func (d DemoPhotos) ReadSamples() {
	var inspectionDir Path
	if false && Alert("Writing inspection images") {
		dsk := HomeDirM().JoinM("Desktop")
		CheckState(dsk.IsDir())
		inspectionDir = dsk.JoinM("_animal_inspection_")
		inspectionDir.MkDirsM()
	}

	dir := FindProjectDirM().JoinM("sample_photos")
	if !dir.Exists() {
		Alert("!No such directory:", dir)
		return
	}
	dirTarget := FindProjectDirM().JoinM("sample_photos_scaled")
	dirTarget.MkDirsM()

	w := NewDirWalk(dir).IncludeExtensions("jpg", "jpeg", "png")
	for _, f := range w.Files() {

		Pr("file:", f)
		nm := f.TrimExtension().Base()
		targetFile := dirTarget.JoinM(nm + ".jpg")
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
			if inspectionDir.NonEmpty() {
				x := inspectionDir.JoinM(nm + ".jpg")
				x.WriteBytesM(CheckOkWith(scaled.ToJPEG()))
			}
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
