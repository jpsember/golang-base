package webapp

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/webapp/gen/webapp_data"
	. "github.com/jpsember/golang-base/webserv"
)

type AnimalCardStruct struct {
	BaseWidgetObj
	cardListener   ButtonWidgetListener
	buttonListener ButtonWidgetListener
	buttonLabel    string
	children       []Widget
}

const (
	acchild_photothumbnail = iota
	acchild_name
	acchild_summary
	acchild_button
)

type AnimalCard = *AnimalCardStruct

func NewAnimalCard(m WidgetManager, cardListener ButtonWidgetListener, buttonLabel string, buttonListener ButtonWidgetListener) AnimalCard {
	Todo("!Not sure we will need card buttons")
	Todo("!The feed_item: prefix is duplicated within the card widget ids")
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

func (w AnimalCard) AddChildren(m WidgetManager) {
	pr := PrIf("", false)
	pr("adding children to new card")

	m.OpenContainer(w)

	{
		// Wrap the card listener so we can process it as a list item...?

		m.Listener(func(s Session, w2 Widget, msg string) {
			Pr("image listener within animal card")
			w.cardListener(s, w2, msg)
		})
		m.Id(Animal_PhotoThumbnail).AddImage(w.imageURLProvider)

		m.Id(Animal_Name).Size(SizeTiny).AddHeading()
		m.Id(Animal_Summary).AddText()
	}
	if w.buttonLabel != "" {
		m.Align(AlignRight).Size(SizeSmall).Label(w.buttonLabel).Listener(w.ourButtonListener).AddBtn()
	}
	m.Close()

	pr("done adding children")
}

func (w AnimalCard) AddChild(c Widget, manager WidgetManager) {
	w.children = append(w.children, c)
}

func (w AnimalCard) RenderTo(s Session, m MarkupBuilder) {

	// Open a bootstrap card
	m.Comments("Animal Card")

	m.TgOpen(`div class="card bg-light mb-3"`).Style(`width:14em`).TgContent()
	{
		RenderWidget(w.children[acchild_photothumbnail], s, m)

		// Display title and brief summary
		m.Comments("title and summary")
		m.TgOpen(`div class="card-body" style="max-height:8em; padding-top:.5em;  padding-bottom:.2em;"`).TgContent()
		{
			Todo("!This used to be h6, but due to validation problem, changed to div:")
			m.TgOpen(`div class="card-title"`).TgContent()
			RenderWidget(w.children[acchild_name], s, m)
			m.TgClose()

			// Summary
			Alert("!this used to be a <p>, but to fix validation is now <div>")
			m.TgOpen(`div class="card-text"`).Style(`font-size:75%;`).TgContent()
			RenderWidget(w.children[acchild_summary], s, m)
			m.TgClose()
		}
		m.TgClose()

		m.Comments(`Progress towards goal, controls`)
		var campaignBalance, campaignTarget int

		Todo("!add widgets for the campaign balance")

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

			if len(w.children) > acchild_button {
				m.Comments("right-justified button")
				m.TgOpen(`div`).A(` class="row"`).TgContent()
				{
					m.TgOpen(`div`).A(` class="d-grid justify-content-md-end"`).TgContent()
					RenderWidget(w.children[acchild_button], s, m)
					m.TgClose()
				}
				m.TgClose()
			}
		}
		m.TgClose()
	}
	m.TgClose()
}

func (w AnimalCard) imageURLProvider(s Session) string {
	imgUrl := "unknown"
	photoId := s.WidgetIntValue(w.children[acchild_photothumbnail])
	if photoId == 0 {
		Alert("#50Animal has no photo")
	} else {
		imgUrl = SharedWebCache.GetBlobURL(photoId)
	}
	return imgUrl
}
