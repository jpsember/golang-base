package webapp

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/webapp/gen/webapp_data"
	. "github.com/jpsember/golang-base/webserv"
)

// A Widget that displays editable text
type AnimalCardWidgetObj struct {
	BaseWidgetObj
	animal   Animal
	children *Array[Widget]
}

type AnimalCardWidget = *AnimalCardWidgetObj

func OpenAnimalCardWidget(m WidgetManager, baseId string, animal Animal, viewButtonListener ButtonWidgetListener) {
	widget := newAnimalCardWidget(baseId, animal)
	m.OpenContainer(widget)
	// Create a button within this card
	m.Id(baseId + "_view").Label(`View`).Size(SizeSmall).AddButton(viewButtonListener)
	m.Close()
}

func newAnimalCardWidget(widgetId string, animal Animal) AnimalCardWidget {
	w := AnimalCardWidgetObj{}
	w.Base().BaseId = widgetId
	w.animal = animal
	w.children = NewArray[Widget]()
	return &w
}

func (w AnimalCardWidget) RenderTo(s Session, m MarkupBuilder) {
	if !w.Visible() {
		m.RenderInvisible(w)
		return
	}
	// Add the single child widget (a view button)
	ch := w.GetChildren()
	CheckState(len(ch) == 1, "expected single 'view' button widget")
	vb := ch[0]
	RenderAnimalCard(s, w.animal, m, vb)
}

func (w AnimalCardWidget) GetChildren() []Widget {
	return w.children.Array()
}

const maxChildren = 1

func (w AnimalCardWidget) AddChild(c Widget, manager WidgetManager) {
	CheckState(w.children.Size() < maxChildren)
	w.children.Add(c)
}

func ReadImageIntoCache(blobId int) string {
	s := SharedWebCache
	blob := s.GetBlobWithId(blobId)
	var url string
	if blob.Id() == 0 {
		url = "missing.jpg"
	} else {
		url = "r/" + blob.Name()
	}
	return url
}

func RenderAnimalCard(s Session, w_animal Animal, m MarkupBuilder, button Widget) {

	// Open a bootstrap card

	m.Comments("AnimalCardWidget").OpenTag(`div class="card bg-light mb-3 animal-card"`)
	{
		imgUrl := "unknown"
		photoId := w_animal.PhotoThumbnail()
		if photoId == 0 {
			Alert("!Animal has no photo")
		} else {
			imgUrl = ReadImageIntoCache(photoId)
		}

		m.Comment("animal image")
		m.A(`<img class="card-jimg-top" src="`, imgUrl, `"`)
		// Changing the image size here doesn't seem to change the card size
		//m.A(` width="250" height="375"`)
		m.A(`>`).Cr()

		// Display title and brief summary
		m.Comments("title and summary").
			OpenTag(`div class="card-body" style="max-height:8em; padding-top:.5em;  padding-bottom:.2em;"`)
		{

			m.OpenTag(`h6 class="card-title"`)
			{
				m.Escape(w_animal.Name())
			}
			m.CloseTag()

			m.OpenTag(`p class="card-text" style="font-size:75%;"`)
			{
				m.Escape(w_animal.Summary())
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
				m.Escape(CurrencyToString(w_animal.CampaignBalance()) + ` raised of ` + CurrencyToString(w_animal.CampaignTarget()) + ` goal`)
			}
			m.CloseTag()
			m.Comments("right-justified button").OpenTag(`div class="row"`)
			{
				m.OpenTag(`div class="d-grid justify-content-md-end"`)
				{
					Todo("Figure out how to create a button on-the-fly, at render time?")

					if button != nil {
						vb := button

						// Add the single child widget (a view button)

						Todo("!Add ability to add style = 'width:100%; font-size:75%;' to the child button")
						Todo("!add:  <button class='btn btn-primary btn-sm'> to button")
						Todo("assuming session doesn't need to be sent here")
						vb.RenderTo(s, m)
					}
				}
				m.CloseTag()
			}
			m.CloseTag()
		}
		m.CloseTag()
	}
	m.CloseTag()
}
