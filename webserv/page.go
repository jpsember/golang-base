package webserv

import (
	. "github.com/jpsember/golang-base/base"
)

type Page interface {
	// Note: go doesn't support covariant return types, so this must return Page, not some concrete implementation of it
	Name() string
	Session() Session
	Construct(s Session, args ...any) Page
	Generate()
	Request(s Session, parser PathParse) Page
}

type PageDevLabelRenderer func(s Session, p Page)

var DevLabelRenderer PageDevLabelRenderer

// Some common boilerplate that is typically some of the first code that
// Generate() would otherwise execute.
func GenerateHeader(page Page) WidgetManager {
	var _ = Pr
	//SetWidgetDebugRendering()
	s := page.Session()
	m := s.WidgetManager()
	m.With(s.PageWidget)
	if DevLabelRenderer != nil {
		DevLabelRenderer(s, page)
	}
	Todo("We must also include the arguments, if any... but how?")
	Todo("Set browser expression to url expr?")
	//s.SetURLExpression(page.Name())
	return m
}
