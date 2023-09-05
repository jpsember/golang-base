package webapp

import (
	. "github.com/jpsember/golang-base/base"
	"github.com/jpsember/golang-base/webapp/gen/webapp_data"
	. "github.com/jpsember/golang-base/webserv"
	"strings"
)

type AnimalFeedPageStruct struct {
	sess         Session
	parentWidget Widget
}

type AnimalFeedPage = *AnimalFeedPageStruct

func NewAnimalFeedPage(sess Session, parentWidget Widget) AnimalFeedPage {
	t := &AnimalFeedPageStruct{
		sess:         sess,
		parentWidget: parentWidget,
	}
	return t
}

func (p AnimalFeedPage) Generate() {
	SetWidgetDebugRendering()

	m := p.sess.WidgetManager()
	m.With(p.parentWidget)

	alertWidget = NewAlertWidget("sample_alert", AlertInfo)
	alertWidget.SetVisible(false)
	m.Add(alertWidget)

	m.Size(SizeTiny).Label("AnimalFeedPage").AddHeading()

	//heading := NewHeadingWidget("header_text", 1)
	//m.Add(heading)

	m.Col(4)
	for i := 1; i < 12; i++ {
		anim, err := webapp_data.ReadAnimal(i)
		if err != nil {
			Pr("what do we do with unexpected errors?", INDENT, err)
		}
		if anim == nil {
			continue
		}
		cardId := "animal_" + IntToString(int(anim.Id()))
		OpenAnimalCardWidget(m, cardId, anim, buttonListener)
	}

	m.Col(4)
	m.Label("uniform delta").AddText()
	m.Col(8)
	m.Id("x58").Label(`X58`).Listener(buttonListener).AddButton().SetEnabled(false)

	m.Col(2).AddSpace()
	m.Col(3).AddSpace()
	m.Col(3).AddSpace()
	m.Col(4).AddSpace()

	m.Col(6)
	m.Listener(birdListener)
	m.Label("Bird").Id("bird")
	m.AddInput()

	m.Col(6)
	m.Open()
	m.Id("x59").Label(`Label for X59`).Listener(checkboxListener).AddCheckbox()
	m.Id("x60").Label(`With fruit`).Listener(checkboxListener).AddSwitch()
	m.Close()

	m.Col(4)
	m.Id("launch").Label(`Launch`).Listener(buttonListener).AddButton()

	m.Col(8)
	m.Label(`Sample text; is 5 < 26? A line feed
"Quoted string"
Multiple line feeds:


   an indented final line`)
	m.AddText()

	m.Col(4)
	m.Listener(zebraListener)
	m.Label("Animal").Id("zebra").AddInput()
}

func birdListener(s Session, widget Widget) {
	Todo("?can we have sessions produce listener functions with appropriate handling of sess any?")
	newVal := s.GetValueString()
	s.SetWidgetProblem(widget, nil)
	s.State.Put(widget.Id(), newVal)
	Todo("!do validation as a global function somewhere")
	if newVal == "parrot" {
		s.SetWidgetProblem(widget, "No parrots, please!")
	}
	s.WidgetManager().Repaint(widget)
}

func zebraListener(s Session, widget Widget) {

	// Get the requested new value for the widget
	newVal := s.GetValueString()

	// Store this as the new value for this widget within the session state map
	s.State.Put(widget.Id(), newVal)

	// Increment the alert class, and update its message
	alertWidget.Class = (alertWidget.Class + 1) % AlertTotal

	alertWidget.SetVisible(newVal != "")

	s.State.Put(alertWidget.BaseId,
		strings.TrimSpace(newVal+" "+
			RandomText(myRand.Rand(), 55, false)))
	s.WidgetManager().Repaint(alertWidget)
}

func buttonListener(s Session, widget Widget) {
	wid := s.GetWidgetId()
	newVal := "Clicked: " + wid

	// Increment the alert class, and update its message
	alertWidget.Class = (alertWidget.Class + 1) % AlertTotal
	alertWidget.SetVisible(true)

	s.State.Put(alertWidget.BaseId,
		strings.TrimSpace(newVal))
	s.WidgetManager().Repaint(alertWidget)
}

func checkboxListener(s Session, widget Widget) {
	wid := s.GetWidgetId()

	// Get the requested new value for the widget
	newVal := s.GetValueBoolean()

	Todo("It is safe to not check if there was a RequestProblem, as any state changes will still go through validation...")

	s.State.Put(wid, newVal)
	// Repainting isn't necessary, as the web page has already done this
}
