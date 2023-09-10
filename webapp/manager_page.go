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
	t.devLabel = "manager_page"
	return t
}

func (p ManagerPage) Generate() {
	m := p.GenerateHeader()

	// Row of buttons at top.
	m.Open()
	{
		m.Listener(p.newAnimalListener).Label("New Animal").AddButton(nil)
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
