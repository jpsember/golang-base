package webapp

import (
	. "github.com/jpsember/golang-base/base"
	"github.com/jpsember/golang-base/webapp/gen/webapp_data"
	"github.com/jpsember/golang-base/webserv"
)

func RandomAnimal() webapp_data.AnimalBuilder {
	r := webserv.HTMLRand.Rand()
	a := webapp_data.NewAnimal()
	a.SetName(RandomText(r, 20, false))
	a.SetSummary(RandomText(r, 300, false))
	a.SetDetails(RandomText(r, 2000, true))
	a.SetCampaignTarget(int32((r.Int63n(10) + 2) * 50 * DollarsToCurrency))
	a.SetCampaignBalance(r.Int31n(a.CampaignTarget()))
	return a
}

func GenerateRandomAnimals() {
	for i := 0; i < 100; i++ {
		anim := RandomAnimal()
		db := Db()
		db.AddAnimal(anim)
		Pr("added animal:", INDENT, anim)
	}
}
