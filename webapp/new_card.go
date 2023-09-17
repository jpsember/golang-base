package webapp

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/webapp/gen/webapp_data"
	. "github.com/jpsember/golang-base/webserv"
)

// A Widget that displays editable text
type NewCardObj struct {
	BaseWidgetObj
	animal         Animal
	buttonListener ButtonWidgetListener
	buttonLabel    string
	children       []Widget
}

const (
	vi_title = iota
	vi_summary
)

type NewCard = *NewCardObj

func NewNewCard(
	wm WidgetManager, // manager to add child views
	widgetId string, animal Animal, viewButtonListener ButtonWidgetListener, buttonLabel string) NewCard {
	w := NewCardObj{}
	w.BaseId = widgetId
	w.animal = animal
	w.buttonListener = viewButtonListener
	w.buttonLabel = buttonLabel

	Todo("instead of passing around WidgetManager, maybe pass around Sessions, which contain the wm?")
	c := &w
	// save manager state, and start working with this widget
	wm.PushContainer(c)
	c.addChildren(wm)
	wm.PopContainer()
	return c
}

func (w NewCard) addChildren(m WidgetManager) {
	Todo("We want the id to be tied to the parent widget id, and resolve somehow")
	Todo("We want it to add this widget to a *new* widget map, not the current one")
	m.Id("anonid1").AddHeading()
	m.Id("anonid2").AddText()
}

func (w NewCard) RenderTo(s Session, m MarkupBuilder) {
	RenderAnimalCard(s, w.animal, m, w.buttonLabel, action_prefix_animal_card, action_prefix_animal_card)
}

func NewRenderAnimalCard(s Session, animal Animal, m MarkupBuilder, buttonLabel string, buttonActionPrefix string, cardActionPrefix string) {

	// Open a bootstrap card

	m.Comments("NewCard")
	clickArg := ""
	if cardActionPrefix != "" {
		clickArg = ` onclick="jsButton('` + cardActionPrefix + IntToString(animal.Id()) + `')"`
	}
	m.OpenTag(`div class="card bg-light mb-3" style="width:14em"`, clickArg)
	{
		imgUrl := "unknown"
		photoId := animal.PhotoThumbnail()
		if photoId == 0 {
			Alert("!Animal has no photo")
		} else {
			imgUrl = SharedWebCache.GetBlobURL(photoId)
		}

		m.Comment("animal image")
		m.A(`<img src="`, imgUrl, `"`)

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
						buttonId := buttonActionPrefix + IntToString(animal.Id())
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
