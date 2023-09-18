package webapp

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/webapp/gen/webapp_data"
	. "github.com/jpsember/golang-base/webserv"
)

// A Widget that displays editable text
type NewCardObj struct {
	BaseWidgetObj
	animal           Animal
	buttonListener   ButtonWidgetListener
	buttonLabel      string
	cardActionPrefix string
	children         []Widget
}

type NewCard = *NewCardObj

func NewNewCard(widgetId string, animal Animal, viewButtonListener ButtonWidgetListener, buttonLabel string, cardActionPrefix string) NewCard {
	Todo("Add explicit listener for the card action (somehow)")
	Pr("constructing new card")
	CheckArg(animal.Id() != 0, "no animal")
	w := NewCardObj{
		animal:           animal,
		buttonListener:   viewButtonListener,
		buttonLabel:      buttonLabel,
		cardActionPrefix: cardActionPrefix,
	}
	w.InitBase(widgetId)
	Todo("!instead of passing around WidgetManager, maybe pass around Sessions, which contain the wm?")
	return &w
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
	m.Close()

	m.PopStateProvider()

	pr("done adding children")
}

func (w NewCard) AddChild(c Widget, manager WidgetManager) {
	w.children = append(w.children, c)
}

func (w NewCard) RenderTo(s Session, m MarkupBuilder) {
	ci := 0
	cimax := len(w.children)

	Todo("I think the button vs card presses are getting jumbled up.  Maybe move the card press to the image only")
	// Open a bootstrap card

	animal := w.animal

	m.Comments("Animal Card")

	m.OpenTag(`div class="card bg-light mb-3" style="width:14em"`)
	{
		imgUrl := "unknown"
		photoId := animal.PhotoThumbnail()
		if photoId == 0 {
			Alert("!Animal has no photo")
		} else {
			imgUrl = SharedWebCache.GetBlobURL(photoId)
		}

		clickArg := ""
		if w.cardActionPrefix != "" {
			clickArg = ` onclick="jsButton('` + w.cardActionPrefix + IntToString(animal.Id()) + `')"`
		}
		m.Comment("animal image")
		m.A(`<img src="`, imgUrl, `" `, clickArg)

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
					m.CloseTag()
				}
				m.CloseTag()
				ci++
			}
		}
		m.CloseTag()
	}
	m.CloseTag()
}
