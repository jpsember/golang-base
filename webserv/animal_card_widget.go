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
	if !w.Visible() {
		m.RenderInvisible(w)
		return
	}
	m.OpenHtml(`div class="card bg-light mb-3 animal-card"`, "AnimalCardWidget")
	Pr(`
<div class="card bg-light mb-3" style="max-width:16em;">

        <img class="card-img-top" src="_SKIP_0.jpg">

        <div class="card-body" style="max-height:8em; padding-top:.5em;  padding-bottom:.2em;">
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



        <div class="row">
          <div class="col-sm-7">
          </div>
          <div class="col-sm-5">
            <div id='view'>
              <button class='btn btn-primary' style="width:100%; font-size:75%;" >View</button>
            </div>
          </div>
        </div>


        </div>
        </div> <!-- card-body -->

      </div> <!-- card -->
`)

	m.A(`<div id="`)
	m.A(w.Id)
	m.A(`">`)

	m.DoIndent()

	problemId := w.Id + ".problem"
	problemText := state.OptString(problemId, "")
	if false && Alert("always problem") {
		problemText = "sample problem information"
	}
	hasProblem := problemText != ""

	labelHtml := w.Label
	if labelHtml != nil {
		m.Comment("Label")
		m.A(`<label class="form-label" style="font-size:70%">`).Cr()
		m.Escape(labelHtml)
		m.A(`</label>`).Cr()
	}

	m.Comment("Input")
	m.A(`<input class="form-control`)
	if hasProblem {
		m.A(` border-danger border-3`) // Adding border-3 makes the text shift a bit on error, maybe not desirable
	}
	m.A(`" type="text" id="`)
	m.A(w.Id)
	m.A(`.aux" value="`)
	value := WidgetStringValue(state, w.Id)
	m.Escape(value)
	m.A(`" onchange='jsVal("`)
	m.A(w.Id)
	m.A(`")'>`).Cr()

	if hasProblem {
		m.Comment("Problem")
		m.A(`<div class="form-text`)
		m.A(` text-danger" style="font-size:  70%">`)
		m.Escape(problemText).A(`</div>`).Cr()
	}

	m.DoOutdent()

	m.A(`</div>`)
	m.CloseHtml(`div`, `AnimalCardWidget`)
	m.Cr()
}
