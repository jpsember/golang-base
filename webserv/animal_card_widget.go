package webserv

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/webapp/gen/webapp_data"
)

// A Widget that displays editable text
type AnimalCardWidgetObj struct {
	BaseWidgetObj
	animal   Animal
	children *Array[Widget]
}

type AnimalCardWidget = *AnimalCardWidgetObj

func OpenAnimalCardWidget(m WidgetManager, baseId string, animal Animal, viewButtonListener WidgetListener) {
	widget := newAnimalCardWidget(baseId, animal)
	m.OpenContainer(widget)
	// Create a button within this card
	m.Id(baseId + "_view").Text(`View`).Listener(viewButtonListener).Size(SizeSmall).AddButton()
	m.Close()
}

func newAnimalCardWidget(widgetId string, animal Animal) AnimalCardWidget {
	w := AnimalCardWidgetObj{}
	w.GetBaseWidget().Id = widgetId
	w.animal = animal
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

			m.OpenTag(`h6 class="card-title"`)
			{
				m.Escape(w.animal.Name())
			}
			m.CloseTag()

			m.OpenTag(`p class="card-text" style="font-size:75%;"`)
			{
				m.Escape(w.animal.Summary())
			}
			m.CloseTag()
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
				m.Escape(CurrencyToString(w.animal.CampaignBalance()) + ` raised of ` + CurrencyToString(w.animal.CampaignTarget()) + ` goal`)
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

type Currency = int32

func CurrencyToString(amount Currency) string {
	pr := PrIf(false)
	pr("currency to string, amount:", amount)
	j := IntToString(int(amount))
	h := len(j)
	pr("j:", j, "h:", h)
	if h < 3 {
		j = "000"[0:3-h] + j
		h = 3
		pr("adjusted, j:", j, "h:", h)
	}
	result := `$` + j[:h-2] + "." + j[h-2:]
	pr("returning:", result)
	return result
}
