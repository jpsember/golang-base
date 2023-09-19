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
	cardListener   CardWidgetListener
	buttonListener CardWidgetListener
	buttonLabel    string
	children       []Widget
	childIdPrefix  string
}

type NewCard = *NewCardObj

type CardWidgetListener func(sess Session, widget NewCard)

func cardListenWrapper(sess Session, widget Widget, value string) (string, error) {
	b := widget.(NewCard)
	b.cardListener(sess, b)
	return "", nil
}

func (w NewCard) Animal() Animal {
	return w.animal
}

func NewNewCard(widgetId string, animal Animal, cardListener CardWidgetListener, buttonLabel string, buttonListener CardWidgetListener) NewCard {
	// An id of zero can be used for constructing a template (e.g., list item widget)
	//	CheckArg(animal.Id() != 0, "no animal")
	CheckArg((buttonLabel == "") == (buttonListener == nil))
	w := NewCardObj{
		animal:         animal,
		cardListener:   cardListener,
		buttonLabel:    buttonLabel,
		buttonListener: buttonListener,
	}
	Todo("!any way of simplifying the LowListener boilerplate here and in other widgets? Using templates perhaps?")
	w.LowListen = cardListenWrapper // Only has an effect if cardListener != nil
	w.InitBase(widgetId)
	Todo("!instead of passing around WidgetManager, maybe pass around Sessions, which contain the wm?")
	return &w
}

func (w NewCard) ourButtonListener(sess Session, widget Widget) {
	Pr("ourButtonListener called...")
	w.buttonListener(sess, w)
}

func (w NewCard) AddChildren(m WidgetManager) {
	pr := PrIf(false)
	pr("adding children to new card")

	// Determine a unique prefix for this card's fields, in case we are rendering multiple cards
	// (and not in a list)
	w.childIdPrefix = m.AllocateAnonymousId("card_children.")
	m.OpenContainer(w)
	m.PushIdPrefix(w.childIdPrefix)
	m.Id("name").Size(SizeTiny).AddHeading()
	m.Id("summary").AddText()
	if w.buttonLabel != "" {
		m.Align(AlignRight).Size(SizeSmall).Label(w.buttonLabel).AddButton(w.ourButtonListener)
	}
	m.PopIdPrefix()
	m.Close()
	pr("done adding children")
}

func (w NewCard) AddChild(c Widget, manager WidgetManager) {
	w.children = append(w.children, c)
}

func (w NewCard) SetAnimal(anim Animal) {
	w.animal = anim
}

func (w NewCard) StateProviderFunc() ListItemStateProvider {
	return w.BuildStateProvider
}

func (w NewCard) BuildStateProvider(sess Session, widget ListWidget, elementId int) (string, JSMap) {
	anim := ReadAnimalIgnoreError(elementId)
	CheckState(anim.Id() != 0, "no animal specified")
	w.animal = anim
	return w.childIdPrefix, anim.ToJson().AsJSMap()
}

func (w NewCard) RenderTo(s Session, m MarkupBuilder) {
	ci := 0
	cimax := len(w.children)

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

		// If there's a card listener, treat the image as a big button returning the card's id
		clickArg := ""
		if w.cardListener != nil {
			clickArg = ` onclick="jsButton('` + w.Id() + `')"`
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

			// If there's a button, render it

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
