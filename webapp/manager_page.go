package webapp

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/webserv"
)

type ManagerPageStruct struct {
	BasicPage
}

type ManagerPage = *ManagerPageStruct

func NewManagerPage(sess Session, parentWidget Widget) ManagerPage {
	t := &ManagerPageStruct{
		NewBasicPage(sess, parentWidget),
	}
	return t
}

func (p ManagerPage) Generate() {
	SetWidgetDebugRendering()

	m := p.session.WidgetManager()
	m.With(p.parentPage)

	AddDevPageLabel(p.session, "ManagerPage")

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
	NewCreateAnimalPage(sess, p.parentPage).Generate()
}
