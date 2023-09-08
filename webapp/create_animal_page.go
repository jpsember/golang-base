package webapp

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/webserv"
)

type CreateAnimalPageStruct struct {
	sess         Session
	parentWidget Widget
}

type CreateAnimalPage = *CreateAnimalPageStruct

func NewCreateAnimalPage(sess Session, parentWidget Widget) CreateAnimalPage {

	t := &CreateAnimalPageStruct{
		sess:         sess,
		parentWidget: parentWidget,
	}
	return t
}

func (p CreateAnimalPage) Generate() {
	SetWidgetDebugRendering()

	m := p.sess.WidgetManager()
	m.With(p.parentWidget)

	AddDevPageLabel(p.sess, "CreateAnimalPage")

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
