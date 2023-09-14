package webapp

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/webserv"
)

type PageConstructFunc = func(s Session) Page
type Page interface {
	GetBasicPage() BasicPage
	Constructor() func(s Session) Page
	Generate(s Session)
}

type PageGenerateFunc func()
type BasicPageStruct struct {
	Session  Session
	PageName string
	//Generate PageGenerateFunc
}

type BasicPage = *BasicPageStruct

func InitPage(pg BasicPage, name string, sess Session) {
	Todo("!Move BasicPage to webserv package")
	CheckArg(name != "")
	pg.PageName = name
	pg.Session = sess
}

// Some common boilerplate that is typically some of the first code that
// Generate() would otherwise execute.
func (p BasicPage) GenerateHeader() WidgetManager {
	//SetWidgetDebugRendering()
	s := p.Session
	m := s.WidgetManager()
	m.With(s.PageWidget)
	AddDevPageLabel(s, p.PageName)
	s.SetURLExpression(p.PageName)
	return m
}
