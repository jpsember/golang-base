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

func NewNewCard(widgetId string, animal Animal, viewButtonListener ButtonWidgetListener, buttonLabel string) NewCard {
	Pr("constructing new card")
	w := NewCardObj{}
	w.InitBase(widgetId)
	w.animal = animal
	w.buttonListener = viewButtonListener
	w.buttonLabel = buttonLabel

	Todo("!instead of passing around WidgetManager, maybe pass around Sessions, which contain the wm?")
	c := &w
	return c
}

func (w NewCard) AddChildren(m WidgetManager) {
	Pr("adding children to new card")
	m.OpenContainer(w)
	Todo("!We want the id to be tied to the parent widget id, and resolve somehow")
	m.Id("anonid1").Size(SizeTiny).AddHeading()
	m.Id("anonid2").AddText()
	m.Close()
	Pr("done adding children")
}

func (w NewCard) AddChild(c Widget, manager WidgetManager) {
	Todo("!How does this differ from the container_widget method?")
	w.children = append(w.children, c)
}

func (w NewCard) RenderTo(s Session, m MarkupBuilder) {

	// Open a bootstrap card

	cardActionPrefix := ""
	animal := ReadAnimalIgnoreError(3)

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
				// Render the name as the first child
				RenderWidget(w.children[0], s, m)
			}
			m.CloseTag()

			// Render the second child
			m.OpenTag(`p class="card-text" style="font-size:75%;"`)
			{
				RenderWidget(w.children[1], s, m)
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

			//if buttonLabel != "" {
			//	m.Comments("right-justified button")
			//	m.OpenTag(`div class="row"`)
			//	{
			//		m.OpenTag(`div class="d-grid justify-content-md-end"`)
			//		{
			//			buttonId := buttonActionPrefix + IntToString(animal.Id())
			//			RenderButton(s, m, buttonId, buttonId, true, buttonLabel, SizeSmall, AlignRight, 0)
			//		}
			//		m.CloseTag()
			//	}
			//	m.CloseTag()
			//}
		}
		m.CloseTag()
	}
	m.CloseTag()
}