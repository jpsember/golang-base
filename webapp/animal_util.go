package webapp

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/webapp/gen/webapp_data"
)

func RandomAnimal(r JSRand, managers []User) AnimalBuilder {
	CheckArg(len(managers) != 0)
	r = NullToRand(r)
	a := NewAnimal()
	a.SetName(RandomText(r, 20, false))
	a.SetSummary(RandomText(r, Ternary(false, 300, 20), false))
	a.SetDetails(RandomText(r, Ternary(false, 2000, 20), true))
	a.SetCampaignTarget((r.Intn(10) + 2) * 50 * DollarsToCurrency)
	a.SetCampaignBalance(r.Intn(a.CampaignTarget()))
	a.SetManagerId(managers[r.Intn(len(managers))].Id())

	if SamplePhotoBlobIdCount > 0 {
		a.SetPhotoThumbnail(SamplePhotoBlobIdStart + r.Intn(SamplePhotoBlobIdCount))
	}
	Pr("random animal, photo:", a.PhotoThumbnail(), "sample count:", SamplePhotoBlobIdCount)
	
	//
	//{
	//	// Copy one of our sample photos to use as this animal's photo
	//	d := SharedDemoPhotos
	//	nms := d.ScaledPhotoNames()
	//	if len(nms) == 0 {
	//		Alert("?No demo photos found; animal(s) won't have photos")
	//	} else {
	//		i := r.Intn(len(nms))
	//		b := NewBlob()
	//		pth := d.scaledPhotosDir().JoinM(nms[i])
	//		b.SetData(pth.ReadBytesM())
	//		AssignBlobName(b)
	//		b2 := CheckOkWith(CreateBlob(b))
	//		a.SetPhotoThumbnail(b2.Id())
	//	}
	//}
	//Todo("Issue #59: add random photo")
	return a
}

func HasAnimals() bool {
	return CheckOkWith(ReadAnimal(1)) != DefaultAnimal
}

func AssignBlobName(b BlobBuilder) {
	if b.Name() == "" {
		Todo("We should continue to do this until we find an unused blob, with appropriate lock")
		b.SetName(string(GenerateBlobName()))
	}
}

var NoSuchAnimalError = Error("No such animal found")

func ReadActualAnimal(id int) (Animal, error) {
	result, err := ReadAnimal(id)
	if err == nil && result.Id() == 0 {
		err = NoSuchAnimalError
	}
	return result, err
}

func ReadAnimalIgnoreError(id int) Animal {
	anim := DefaultAnimal
	anim2, err := ReadAnimal(id)
	if !ReportIfError(err, "failed to read animal") {
		anim = anim2
	}
	return anim
}

func ReadUserIgnoreError(id int) User {
	user := DefaultUser
	user2, err := ReadUser(id)
	if !ReportIfError(err, "failed to read user") {
		user = user2
	}
	return user
}

func AnimalExistsAndIsActive(id int) bool {
	anim := ReadAnimalIgnoreError(id)
	return anim.Id() != 0
}

func AnimalExistsAndIsActiveForManager(animalId int, managerId int) bool {
	anim := ReadAnimalIgnoreError(animalId)
	return anim.Id() != 0 && anim.ManagerId() == managerId
}
