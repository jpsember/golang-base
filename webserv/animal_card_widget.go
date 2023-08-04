package webserv

import (
	. "github.com/jpsember/golang-base/base"
)

// A Widget that displays editable text
type AnimalCardWidgetObj struct {
	BaseWidgetObj
	animalId string
	children *Array[Widget]
}

type AnimalCardWidget = *AnimalCardWidgetObj

func OpenAnimalCardWidget(m WidgetManager, baseId string, animalId string, viewButtonListener WidgetListener) {
	widget := newAnimalCardWidget(baseId, animalId)
	m.OpenContainer(widget)
	// Create a button within this card
	m.Id(baseId + "_view").Text(`View`).Listener(viewButtonListener).Size(SizeSmall).AddButton()
	m.Close()
}

func newAnimalCardWidget(widgetId string, animalId string) AnimalCardWidget {
	w := AnimalCardWidgetObj{}
	w.GetBaseWidget().Id = widgetId
	w.animalId = animalId
	w.children = NewArray[Widget]()
	return &w
}

var picCounter = 0

func (w AnimalCardWidget) RenderTo(m MarkupBuilder, state JSMap) {

	if !w.Visible() {
		m.RenderInvisible(w)
		return
	}

	// Open a bootstrap card

	m.Comments("AnimalCardWidget").OpenTag(`div class="card bg-light mb-3 animal-card"`)
	{

		// Display an image
		picCounter++
		imgUrl := IntToString(MyMod(picCounter, 3)) + ".jpg"
		Todo("!add support for image based on particular animal")
		m.Comment("animal image").VoidTag(`img class="card-img-top" src="`, imgUrl, `"`)
		//m.Pr(`<img class="card-img-top" src="0.jpg">`).Cr()

		// Display title and brief summary
		m.Comments("title and summary").
			OpenTag(`div class="card-body" style="max-height:8em; padding-top:.5em;  padding-bottom:.2em;"`)
		{

			Todo("!display animal name")

			m.A(`<h6 class="card-title">Roscoe</h6>`).Cr()

			m.A(`<p class="card-text" style="font-size:75%;">This boxer cross came 
                           to us with skin issues and needs additional treatment.  
                           She is on the mend though!</p>`).Cr()
		}
		m.CloseTag()

		m.Comments(`Progress towards goal, controls`).OpenTag(`div class="card-body"`)
		{
			m.Comments("progress-container").OpenTag(`div class="progress-container"`)
			{
				m.Comment("Plot grey in background, full width").OpenCloseTag(`div class="progress-bar-bgnd"`)
				m.Comment("Plot bar graph in foreground, partial width").OpenCloseTag(`div class="progress-bar" style="width: 35%;"`)
			}
			m.CloseTag()
			m.OpenTag(`div class="progress-text"`)
			{
				m.Pr("$120 raised of $250 goal")
			}
			m.CloseTag()
			m.Comments("right-justified button").OpenTag(`div class="row"`)
			{
				m.OpenTag(`div class="d-grid justify-content-md-end"`)
				{
					// Add the single child widget (a view button)
					ch := w.GetChildren()
					CheckState(len(ch) == 1, "expected single 'view' button widget")
					vb := ch[0]
					Todo("!Add ability to add style = 'width:100%; font-size:75%;' to the child button")
					Todo("!add:  <button class='btn btn-primary btn-sm'> to button")
					vb.RenderTo(m, state)
				}
				m.CloseTag()
			}
			m.CloseTag()
		}
		m.CloseTag()
	}
	m.CloseTag()
}

func (w AnimalCardWidget) GetChildren() []Widget {
	return w.children.Array()
}

const maxChildren = 1

func (w AnimalCardWidget) AddChild(c Widget, manager WidgetManager) {
	CheckState(w.children.Size() < maxChildren)
	w.children.Add(c)
}
