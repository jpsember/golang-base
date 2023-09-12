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
	buttonListener ButtonWidgetListener
	buttonLabel    string
}

type AnimalCardWidget = *AnimalCardWidgetObj

func NewAnimalCardWidget(widgetId string, animal Animal, viewButtonListener ButtonWidgetListener, buttonLabel string) AnimalCardWidget {
	w := AnimalCardWidgetObj{}
	w.Base().BaseId = widgetId
	w.animal = animal
	w.buttonListener = viewButtonListener
	w.buttonLabel = buttonLabel
	return &w
}

func (w AnimalCardWidget) RenderTo(s Session, m MarkupBuilder) {
	if !w.Visible() {
		m.RenderInvisible(w)
		return
	}
	RenderAnimalCard(s, w.animal, m, w.buttonLabel, "amnimal_id_")
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

func RenderAnimalCard(s Session, animal Animal, m MarkupBuilder, buttonLabel string, actionPrefix string) {

	// Open a bootstrap card

	m.Comments("AnimalCardWidget")
	m.OpenTag(`div class="card bg-light mb-3 animal-card" style="width:14em"`)
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

		PlotImageSizeMarkup(s, m, IPointZero) //AnimalPicSizeNormal.ScaledBy(0.4))

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
						buttonId := actionPrefix + IntToString(animal.Id())
						RenderButton(s, m, buttonId, buttonId, true, buttonLabel, SizeSmall, AlignRight, 0)
					}
					m.CloseTag()
				}
				m.CloseTag()
			}
		}
		m.CloseTag()
	}
	m.CloseTag()
}
