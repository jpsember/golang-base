package webapp

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/webserv"
)

type ManagerPageStruct struct {
	sess         Session
	parentWidget Widget
}

type ManagerPage = *ManagerPageStruct

func NewManagerPage(sess Session, parentWidget Widget) ManagerPage {
	t := &ManagerPageStruct{
		sess:         sess,
		parentWidget: parentWidget,
	}
	return t
}

func (p ManagerPage) Generate() {
	SetWidgetDebugRendering()

	m := p.sess.WidgetManager()
	m.With(p.parentWidget)

	AddDevPageLabel(p.sess, "ManagerPage")

	// Row of buttons at top.
	m.Open()
	{
		m.Listener(p.newAnimalListener).Label("New Animal").AddButton()
	}
	m.Close()

	// Scrolling list of animals for this manager.
	m.Open()
	Todo("?Scrolling list of manager's animals")
	m.Close()

}

func (p ManagerPage) newAnimalListener(sess Session, widget Widget) {

	if Todo("CreateAnimalPage") {

	} else {
		//
		//sp := CreateAnimalPage(sess, p.parentWidget)
		//sp.Generate()
	}

}
