package webserv

import (
	. "github.com/jpsember/golang-base/base"
)

// A Widget that displays editable text
type AnimalCardWidgetObj struct {
	BaseWidgetObj
	animalId string
}

type AnimalCardWidget = *AnimalCardWidgetObj

func NewAnimalCardWidget(widgetId string, aId string) AnimalCardWidget {
	w := AnimalCardWidgetObj{}
	w.GetBaseWidget().Id = widgetId
	w.animalId = aId
	return &w
}

func (w AnimalCardWidget) RenderTo(m MarkupBuilder, state JSMap) {

	Todo("This is another type of container widget, have to figure that out")

	if !w.Visible() {
		m.RenderInvisible(w)
		return
	}
	m.OpenHtml(`div class="card bg-light mb-3 animal-card"`, "AnimalCardWidget")

	m.Pr(`<img class="card-img-top" src="0.jpg">`).Cr()
	m.Pr(` <div class="card-body" style="max-height:8em; padding-top:.5em;  padding-bottom:.2em;">
          <h6 class="card-title">Roscoe</h6>
          <p class="card-text" style="font-size:75%;">This boxer cross came to us with skin issues and needs additional treatment.  She is on the mend though!</p>
        </div>


        <div class="card-body">

          <div class="progress-container">
            <!-- Plot grey in background, full width -->
            <div class="progress-bar-bgnd"></div>

            <!-- Plot bar graph in foreground, partial width -->
            <div class="progress-bar" style="width: 35%;"></div>

          </div>

          <div class="progress-text">
            $120 raised of $250 goal
          </div>

 </div>
`).Cr()
	m.CloseHtml(`div`, `AnimalCardWidget`)
	m.Cr()
}
