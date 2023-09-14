package webapp

import (
	. "github.com/jpsember/golang-base/webserv"
)

type Page interface {
	// Note: go doesn't support covariant return types, so this must return Page, not some concrete implementation of it
	Construct(s Session) Page
	Generate()
	Name() string
	Session() Session
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
