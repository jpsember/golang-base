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

const animal_field_prefix = "card."

const (
	vi_title = iota
	vi_summary
)

type NewCard = *NewCardObj

func NewNewCard(widgetId string, animal Animal, viewButtonListener ButtonWidgetListener, buttonLabel string) NewCard {
	Pr("constructing new card")
	w := NewCardObj{}
	w.InitBase(widgetId)
	CheckArg(animal.Id() != 0, "no animal")
	w.animal = animal
	w.buttonListener = viewButtonListener
	w.buttonLabel = buttonLabel

	Todo("!instead of passing around WidgetManager, maybe pass around Sessions, which contain the wm?")
	c := &w
	return c
}

func (w NewCard) AddChildren(m WidgetManager) {
	pr := PrIf(false)
	pr("adding children to new card")

	// Construct a WidgetStateProvider to access this particular animal's data

	jsmap := w.animal.ToJson().AsJSMap()
	m.PushStateProvider(anim_state_prefix, jsmap)
	pr("pushing state provider, prefix:", anim_state_prefix, "map:", jsmap)

	Todo("ability to push id prefix for subsequent id() calls")

	m.OpenContainer(w)
	m.Id(anim_state_prefix + "name").Size(SizeTiny).AddHeading()
	m.Id(anim_state_prefix + "summary").AddText()
	if w.buttonLabel != "" {
		m.Align(AlignRight).Size(SizeSmall).Label(w.buttonLabel).AddButton(w.buttonListener)
	}
	Todo("Add button somehow")
	m.Close()

	m.PopStateProvider()

	Pr("done adding children")
}

func (w NewCard) AddChild(c Widget, manager WidgetManager) {
	Todo("!How does this differ from the container_widget method?")
	w.children = append(w.children, c)
}

func (w NewCard) RenderTo(s Session, m MarkupBuilder) {

	ci := 0
	cimax := len(w.children)

	// Open a bootstrap card

	cardActionPrefix := ""
	animal := w.animal

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
				RenderWidget(w.children[ci], s, m)
				ci++
			}
			m.CloseTag()

			// Render the second child
			m.OpenTag(`p class="card-text" style="font-size:75%;"`)
			{
				RenderWidget(w.children[ci], s, m)
				ci++
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

			if ci < cimax {
				m.Comments("right-justified button")
				m.OpenTag(`div class="row"`)
				{
					m.OpenTag(`div class="d-grid justify-content-md-end"`)
					RenderWidget(w.children[ci], s, m)
					ci++
					//{
					//	buttonId := buttonActionPrefix + IntToString(animal.Id())
					//	RenderButton(s, m, buttonId, buttonId, true, buttonLabel, SizeSmall, AlignRight, 0)
					//}
					m.CloseTag()
				}
				m.CloseTag()
				ci++
			}
			if w.buttonLabel != "" {
				//m.Comments("right-justified button")
				//m.OpenTag(`div class="row"`)
				//{
				//	m.OpenTag(`div class="d-grid justify-content-md-end"`)
				//	{
				//		buttonId := buttonActionPrefix + IntToString(animal.Id())
				//		RenderButton(s, m, buttonId, buttonId, true, buttonLabel, SizeSmall, AlignRight, 0)
				//	}
				//	m.CloseTag()
				//}
				//m.CloseTag()
			}
		}
		m.CloseTag()
	}
	m.CloseTag()
}
