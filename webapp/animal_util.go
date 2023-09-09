package webapp

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/webapp/gen/webapp_data"
)

func RandomAnimal(r JSRand) AnimalBuilder {
	r = NullToRand(r)
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
		anim := RandomAnimal(rnd)
		CreateAnimal(anim)
		Pr("added animal:", INDENT, anim)
	}
}

func AssignBlobName(b BlobBuilder) {
	if b.Name() == "" {
		Todo("We should continue to do this until we find an unused blob, with appropriate lock")
		b.SetName(string(GenerateBlobName()))
	}
}
