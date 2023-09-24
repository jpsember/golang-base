package webapp

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/webapp/gen/webapp_data"
	. "github.com/jpsember/golang-base/webserv"
)

// A Widget that displays editable text
type AnimalCardStruct struct {
	BaseWidgetObj
	cardAnimal     Animal
	cardListener   CardWidgetListener
	buttonListener CardWidgetListener
	buttonLabel    string
	children       []Widget
	ChildIdPrefix  string
}

type AnimalCard = *AnimalCardStruct

type CardWidgetListener func(sess Session, widget AnimalCard, arg string)

func cardListenWrapper(sess Session, widget Widget, value string) (any, error) {
	b := widget.(AnimalCard)
	Todo("!Is the listener 'arg' necessary?")
	b.cardListener(sess, b, value)
	Alert("#50cardListenWrapper, calling AnimalCard", b.Id())
	return nil, nil
}

func (w AnimalCard) Animal() Animal {
	return w.cardAnimal
}

func NewAnimalCard(widgetId string, animal Animal, cardListener CardWidgetListener, buttonLabel string, buttonListener CardWidgetListener) AnimalCard {
	// An id of zero can be used for constructing a template (e.g., list item widget)

	// If a button is requested, it must have a listener
	CheckArg((buttonLabel == "") == (buttonListener == nil))

	w := AnimalCardStruct{
		cardAnimal:     animal,
		cardListener:   cardListener,
		buttonLabel:    buttonLabel,
		buttonListener: buttonListener,
	}
	Todo("!any way of simplifying the LowListener boilerplate here and in other widgets? Using templates perhaps?")
	w.LowListen = cardListenWrapper // Only has an effect if cardListener != nil
	w.InitBase(widgetId)
	//w.SetTrace(true)

	Todo("!instead of passing around WidgetManager, maybe pass around Sessions, which contain the wm?")

	return &w
}

func (w AnimalCard) ourButtonListener(sess Session, widget Widget, arg string) {
	Pr("ourButtonListener called...")
	w.buttonListener(sess, w, arg)
}

func (w AnimalCard) AddChildren(m WidgetManager) {
	pr := PrIf("", false)
	pr("adding children to new card")

	// Determine a unique prefix for this card's fields.
	// Note that we do *don't* set any state providers until we know what this prefix is.  Specifically,
	// we don't create a state provider at construction time.
	w.ChildIdPrefix = m.AllocateAnonymousId("card_children:")

	// If we were given an actual animal, give this card's children a default state provider
	if w.cardAnimal.Id() != 0 {
		m.PushStateProvider(NewStateProvider(w.ChildIdPrefix, w.cardAnimal.ToJson()))
	}

	m.OpenContainer(w)
	m.PushIdPrefix(w.ChildIdPrefix)
	if !Experiment {
		c1 := m.Id("name").Size(SizeTiny).AddHeading()
		c2 := m.Id("summary").AddText()
		c1.SetTrace(false)
		c2.SetTrace(false)
	}
	if w.buttonLabel != "" {
		m.Align(AlignRight).Size(SizeSmall).Label(w.buttonLabel).AddButton(w.ourButtonListener)
	}
	m.PopIdPrefix()
	m.Close()
	if w.cardAnimal.Id() != 0 {
		m.PopStateProvider()
	}

	pr("done adding children")
}

func (w AnimalCard) AddChild(c Widget, manager WidgetManager) {
	w.children = append(w.children, c)
}

func (w AnimalCard) SetAnimal(anim Animal) {
	w.cardAnimal = anim
}

func (w AnimalCard) RenderTo(s Session, m MarkupBuilder) {
	ci := 0
	cimax := len(w.children)

	// Open a bootstrap card
	animal := w.cardAnimal
	m.Comments("Animal Card")

	m.TgOpen(`div class="card bg-light mb-3"`).Style(`width:14em`).TgContent()
	{
		imgUrl := "unknown"
		photoId := animal.PhotoThumbnail()
		if photoId == 0 {
			Alert("#50Animal has no photo")
		} else {
			imgUrl = SharedWebCache.GetBlobURL(photoId)
		}

		// If there's a card listener, treat the image as a big button returning the card's id
		clickArg := ""
		if w.cardListener != nil {
			clickId := s.PrependId(w.Id())
			clickArg = ` onclick="jsButton('` + clickId + `')"`
		}

		m.Comment("animal image")
		m.A(`<img src="`, imgUrl, `" `, clickArg)

		PlotImageSizeMarkup(s, m, IPointZero) //AnimalPicSizeNormal.ScaledBy(0.4))

		m.A(`>`).Cr()

		// Display title and brief summary
		m.Comments("title and summary")
		m.TgOpen(`div class="card-body" style="max-height:8em; padding-top:.5em;  padding-bottom:.2em;"`).TgContent()
		if !Experiment {
			m.TgOpen(`h6 class="card-title"`).TgContent()
			{
				// Render the name as the first child
				RenderWidget(w.children[ci], s, m)
				ci++
			}
			m.TgClose()

			// Render the second child
			m.TgOpen(`p class="card-text"`).Style(`font-size:75%;`).TgContent()
			{
				RenderWidget(w.children[ci], s, m)
				ci++
			}
			m.TgClose()
		}
		m.TgClose()

		m.Comments(`Progress towards goal, controls`)
		m.TgOpen(`div class="card-body"`).TgContent()
		{
			m.Comments("progress-container")
			m.TgOpen(`div class="progress-container"`).TgContent()
			{
				m.Comment("Plot grey in background, full width").TgOpen(`div class="progress-bar-bgnd"`).TgContent().TgClose()
				m.Comment("Plot bar graph in foreground, partial width").TgOpen(`div class="progress-bar"`).Style(`width: 35%;`).TgContent().TgClose()
			}
			m.TgClose()
			m.TgOpen(`div class="progress-text"`).TgContent()
			{
				m.A(ESCAPED, CurrencyToString(animal.CampaignBalance())+` raised of `+CurrencyToString(animal.CampaignTarget())+` goal`)
			}
			m.TgClose()

			// If there's a button, render it

			if ci < cimax {
				m.Comments("right-justified button")
				m.TgOpen(`div`).A(` class="row"`).TgContent()
				{
					m.TgOpen(`div`).A(` class="d-grid justify-content-md-end"`).TgContent()
					RenderWidget(w.children[ci], s, m)
					ci++
					m.TgClose()
				}
				m.TgClose()
				ci++
			}
		}
		m.TgClose()
	}
	m.TgClose()
}
