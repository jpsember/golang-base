package webapp

import (
	. "github.com/jpsember/golang-base/webserv"
)

type Page interface {
	// Note: go doesn't support covariant return types, so this must return Page, not some concrete implementation of it
	Name() string
	Session() Session
	Construct(s Session, args ...any) Page
	Generate()
	Request(s Session, parser PathParse) Page
}

// Some common boilerplate that is typically some of the first code that
// Generate() would otherwise execute.
func GenerateHeader(page Page) WidgetManager {
	//SetWidgetDebugRendering()
	s := page.Session()
	m := s.WidgetManager()
	m.With(s.PageWidget)
	AddDevPageLabel(s, page.Name())
	s.SetURLExpression(page.Name())
	return m
}
