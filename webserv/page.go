package webserv

import (
	. "github.com/jpsember/golang-base/base"
)

type Page interface {
	Name() string
	Args() []string // The additional arguments that would show up in the url (e.g., edit/17), args would be ["17"]
	// Attempt to construct a new page with the specified args; return nil if args aren't valid
	Construct(s Session, args PageArgs) Page
	GenerateWidgets(s Session)
}

type PageDevLabelRenderer func(s Session, page Page)

var DevLabelRenderer PageDevLabelRenderer

// Some common boilerplate that is typically some of the first code that
// GenerateWidgets() would otherwise execute.
func GenerateHeader(s Session, p Page) WidgetManager {
	Todo("!We could merge Construct(...) and GenerateWidgets(...)")
	var _ = Pr
	//SetWidgetDebugRendering()
	CheckState(s != nil)
	m := s.WidgetManager()
	m.With(s, s.PageWidget)
	if DevLabelRenderer != nil {
		DevLabelRenderer(s, p)
	}
	return m
}
