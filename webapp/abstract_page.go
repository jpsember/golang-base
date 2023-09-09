package webapp

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/webserv"
)

// The interface that all pages must implement.
type AbstractPage interface {
	Generate()
}

type BasicPageStruct struct {
	session    Session
	parentPage Widget
	devLabel   string
}

type BasicPage = *BasicPageStruct

func NewBasicPage(session Session, parentPage Widget) BasicPage {
	t := &BasicPageStruct{
		session:    session,
		parentPage: parentPage,
	}
	Todo("?Not sure structs are required for pages; session, parentpage could probably be found from widget manager")
	return t
}

// Some common boilerplate that is typically some of the first code that
// Generate() would otherwise execute.
func (p BasicPage) GenerateHeader() WidgetManager {
	//SetWidgetDebugRendering()
	m := p.session.WidgetManager()
	m.With(p.parentPage)
	if p.devLabel != "" {
		AddDevPageLabel(p.session, p.devLabel)
	}
	return m
}
