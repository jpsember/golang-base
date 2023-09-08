package webapp

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/webserv"
)

type CreateAnimalPageStruct struct {
	BasicPage
}

type CreateAnimalPage = *CreateAnimalPageStruct

func NewCreateAnimalPage(sess Session, parentWidget Widget) AbstractPage {
	t := &CreateAnimalPageStruct{
		NewBasicPage(sess, parentWidget),
	}
	return t
}

func (p CreateAnimalPage) Generate() {
	SetWidgetDebugRendering()

	m := p.session.WidgetManager()
	m.With(p.parentPage)

	AddDevPageLabel(p.session, "CreateAnimalPage")

	Todo("")
	//// Row of buttons at top.
	//m.Open()
	//{
	//	m.Listener(p.newAnimalListener).Label("New Animal").AddButton()
	//}
	//m.Close()
	//
	//// Scrolling list of animals for this manager.
	//m.Open()
	//Todo("?Scrolling list of manager's animals")
	//m.Close()

}

//func (p CreateAnimalPage) newAnimalListener(sess Session, widget Widget) {
//
//	if Todo("CreateAnimalPage") {
//
//	} else {
//		//
//		//sp := CreateAnimalPage(sess, p.parentWidget)
//		//sp.Generate()
//	}
//
//}
