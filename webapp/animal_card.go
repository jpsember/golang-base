package webapp

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/webapp/gen/webapp_data"
	. "github.com/jpsember/golang-base/webserv"
)

type AnimalCardStruct struct {
	BaseWidgetObj
	cardListener   CardWidgetListener
	buttonListener CardWidgetListener
	buttonLabel    string
	children       []Widget
}

type AnimalCard = *AnimalCardStruct

// Note: If card is a list item, the widget's Animal() might not be accurate!
type CardWidgetListener func(sess Session, widget AnimalCard, arg string)

func NewAnimalCard(m WidgetManager, cardListener CardWidgetListener, buttonLabel string, buttonListener CardWidgetListener) AnimalCard {
	Todo("!Not sure we will need card buttons")

	widgetId := m.ConsumeOptionalPendingId()

	// If a button is requested, it must have a listener
	CheckArg((buttonLabel == "") == (buttonListener == nil))

	w := AnimalCardStruct{
		cardListener:   cardListener,
		buttonLabel:    buttonLabel,
		buttonListener: buttonListener,
	}
	Todo("!any way of simplifying the LowListener boilerplate here and in other widgets? Using templates perhaps?")
	w.LowListen = w.lowLevelListener // Only has an effect if cardListener != nil
	w.InitBase(widgetId)
	return &w
}

func (w AnimalCard) lowLevelListener(sess Session, widget Widget, value string) (any, error) {
	pr := PrIf("cardListenWrapper", false)
	pr("calling listener for id", QUO, w.Id(), "value", QUO, value)
	Todo("!Is the listener 'value' necessary?")
	if w.cardListener != nil {
		w.cardListener(sess, w, value)
	}
	return nil, nil
}

func (w AnimalCard) ourButtonListener(sess Session, widget Widget, arg string) {
	Pr("ourButtonListener called...")
	w.buttonListener(sess, w, arg)
}

const animalCardItemPrefix = "animal_item:"

func (w AnimalCard) AddChildren(m WidgetManager) {
	pr := PrIf("", false)
	pr("adding children to new card")

	m.OpenContainer(w)

	Todo("!Make animalCardItemPrefix a parameter in case we want multiple lists")
	m.PushIdPrefix(animalCardItemPrefix)
	{
		m.Id(Animal_Name).Size(SizeTiny).AddHeading()
		m.Id(Animal_Summary).AddText()
	}
	if w.buttonLabel != "" {
		m.Align(AlignRight).Size(SizeSmall).Label(w.buttonLabel).AddButton(w.ourButtonListener)
	}
	m.PopIdPrefix()
	m.Close()

	pr("done adding children")
}

func (w AnimalCard) AddChild(c Widget, manager WidgetManager) {
	w.children = append(w.children, c)
}

func (w AnimalCard) RenderTo(s Session, m MarkupBuilder) {
	ci := 0
	cimax := len(w.children)

	// Open a bootstrap card
	m.Comments("Animal Card")

	m.TgOpen(`div class="card bg-light mb-3"`).Style(`width:14em`).TgContent()
	{

		Todo("Use an image widget to render the photo")

		//imgUrl := "unknown"
		//var photoId int
		////photoId := animal.PhotoThumbnail()
		//if photoId == 0 {
		//	Alert("#50Animal has no photo")
		//} else {
		//	imgUrl = SharedWebCache.GetBlobURL(photoId)
		//}
		//
		//// If there's a card listener, treat the image as a big button returning the card's id
		//clickArg := ""
		//if w.cardListener != nil {
		//	clickId := s.PrependId(w.Id())
		//	clickArg = ` onclick="jsButton('` + clickId + `')"`
		//}
		//
		//m.Comment("animal image")
		//m.A(`<img src="`, imgUrl, `" alt="animal image" `, clickArg)
		//
		//PlotImageSizeMarkup(s, m, IPointZero) //AnimalPicSizeNormal.ScaledBy(0.4))
		//
		//m.A(`>`).Cr()

		// Display title and brief summary
		m.Comments("title and summary")
		m.TgOpen(`div class="card-body" style="max-height:8em; padding-top:.5em;  padding-bottom:.2em;"`).TgContent()
		{
			Todo("!This used to be h6, but due to validation problem, changed to div:")
			m.TgOpen(`div class="card-title"`).TgContent()
			{
				// Render the name as the first child
				RenderWidget(w.children[ci], s, m)
				ci++
			}
			m.TgClose()

			// Render the second child
			Alert("!this used to be a <p>, but to fix validation is now <div>")
			m.TgOpen(`div class="card-text"`).Style(`font-size:75%;`).TgContent()
			{
				RenderWidget(w.children[ci], s, m)
				ci++
			}
			m.TgClose()
		}
		m.TgClose()

		m.Comments(`Progress towards goal, controls`)
		var campaignBalance, campaignTarget int

		Todo("add widgets for the campaign balance")

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
				m.A(ESCAPED, CurrencyToString(campaignBalance)+` raised of `+CurrencyToString(campaignTarget)+` goal`)
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
