package webapp

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/webapp/gen/webapp_data"
	. "github.com/jpsember/golang-base/webserv"
)

// A Widget that displays editable text
type AnimalCardWidgetObj struct {
	BaseWidgetObj
	animal         Animal
	children       *Array[Widget]
	buttonListener ButtonWidgetListener
	buttonLabel    string
}

type AnimalCardWidget = *AnimalCardWidgetObj

func OpenAnimalCardWidget(m WidgetManager, baseId string, animal Animal, viewButtonListener ButtonWidgetListener, buttonLabel string) {
	widget := newAnimalCardWidget(baseId, animal, viewButtonListener, buttonLabel)
	m.OpenContainer(widget)
	Todo("what is the point of having a container if it has nothing?")
	//// Create a button within this card
	//m.Id(baseId + "_view").Label(`View`).Size(SizeSmall).AddButton(viewButtonListener)
	m.Close()
}

func newAnimalCardWidget(widgetId string, animal Animal, viewButtonListener ButtonWidgetListener, buttonLabel string) AnimalCardWidget {
	w := AnimalCardWidgetObj{}
	w.Base().BaseId = widgetId
	w.animal = animal
	w.children = NewArray[Widget]()
	w.buttonListener = viewButtonListener
	w.buttonLabel = buttonLabel
	return &w
}

func (w AnimalCardWidget) RenderTo(s Session, m MarkupBuilder) {
	if !w.Visible() {
		m.RenderInvisible(w)
		return
	}
	//// Add the single child widget (a view button)
	//ch := w.GetChildren()
	//CheckState(len(ch) == 1, "expected single 'view' button widget")
	//vb := ch[0]
	RenderAnimalCard(s, w.animal, m, w.buttonLabel)
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

func RenderAnimalCard(s Session, animal Animal, m MarkupBuilder, buttonLabel string) {

	// Open a bootstrap card

	m.Comments("AnimalCardWidget")
	m.OpenTag(`div class="card bg-light mb-3 animal-card"`)
	{
		imgUrl := "unknown"
		photoId := animal.PhotoThumbnail()
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
		m.Comments("title and summary")
		m.OpenTag(`div class="card-body" style="max-height:8em; padding-top:.5em;  padding-bottom:.2em;"`)
		{

			m.OpenTag(`h6 class="card-title"`)
			{
				m.Escape(animal.Name())
			}
			m.CloseTag()

			m.OpenTag(`p class="card-text" style="font-size:75%;"`)
			{
				m.Escape(animal.Summary())
			}
			m.CloseTag()
		}
		m.CloseTag()

		m.Comments(`Progress towards goal, controls`)
		m.OpenTag(`div class="card-body"`)
		{
			m.Comments("progress-container")
			m.OpenTag(`div class="progress-container"`)
			{
				m.Comment("Plot grey in background, full width").OpenCloseTag(`div class="progress-bar-bgnd"`)
				m.Comment("Plot bar graph in foreground, partial width").OpenCloseTag(`div class="progress-bar" style="width: 35%;"`)
			}
			m.CloseTag()
			m.OpenTag(`div class="progress-text"`)
			{
				m.Escape(CurrencyToString(animal.CampaignBalance()) + ` raised of ` + CurrencyToString(animal.CampaignTarget()) + ` goal`)
			}
			m.CloseTag()

			if buttonLabel != "" {
				m.Comments("right-justified button")
				m.OpenTag(`div class="row"`)
				{
					m.OpenTag(`div class="d-grid justify-content-md-end"`)
					{
						Todo("Figure out how to create a button on-the-fly, at render time?")

						buttonId := "animal_id_" + IntToString(animal.Id())

						// Adding py-3 here to put some vertical space between button and other widgets
						m.A(`<div class='py-3' id='`, buttonId, `'>`)
						//m.DoIndent()
						{
							m.A(`<button class='btn btn-primary `)

							//if w.Align() == AlignRight {
							m.A(`float-end `)
							//}

							Todo("add size support")
							//if w.size != SizeDefault {
							//	m.A(MapValue(btnTextSize, w.size))
							//}
							m.A(`'`)
							//}
							//
							//if !w.Enabled() {
							//	m.A(` disabled`)
							//}

							m.A(` onclick='jsButton("`, buttonId, `")'>`)
							m.Escape(buttonLabel)
							m.A(`</button>`)
							m.Cr()
						}
						//m.DoOutdent()
						m.A(`</div>`)
					}
					//
					//
					//if button != nil {
					//	vb := button
					//
					//	// Add the single child widget (a view button)
					//
					//	Todo("!Add ability to add style = 'width:100%; font-size:75%;' to the child button")
					//	Todo("!add:  <button class='btn btn-primary btn-sm'> to button")
					//	Todo("assuming session doesn't need to be sent here")
					//	RenderWidget(vb, s, m)
					//}
					m.CloseTag()
				}
				m.CloseTag()
			}
		}
		m.CloseTag()
	}
	m.CloseTag()
}
