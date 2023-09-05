package webapp

import (
	. "github.com/jpsember/golang-base/base"
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
