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

	Todo("If we are generating a new page, we shouldn't try to store the error in the old one")
	// Row of buttons at top.
	m.Open()
	{
		m.Label("New Animal").AddButton(p.newAnimalListener)
	}
	m.Close()

	Todo("ability to store some user-specific data types in the session other than the state")

	// Scrolling list of animals for this manager.
	m.Open()
	Todo("?Scrolling list of manager's animals")
	m.Close()

}

func (p ManagerPage) newAnimalListener(sess Session, widget Widget) error {
	NewCreateAnimalPage(sess, p.parentPage).Generate()
	return nil
}
