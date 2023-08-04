package webapp

import (
	"github.com/jpsember/golang-base/base"
	"github.com/jpsember/golang-base/webapp/gen/webapp_data"
	"github.com/jpsember/golang-base/webserv"
)

func RandomAnimal() webapp_data.Animal {
	r := webserv.HTMLRand.Rand()
	a := webapp_data.NewAnimal()
	a.SetId(r.Int63n(50000) + 1)
	a.SetName(base.RandomText(r, 20, false))
	a.SetSummary(base.RandomText(r, 300, false))
	a.SetDetails(base.RandomText(r, 2000, true))
	a.SetCampaignTarget(int32((r.Int63n(10) + 2) * 50 * DollarsToCurrency))
	a.SetCampaignBalance(r.Int31n(a.CampaignTarget()))
	return a
}
