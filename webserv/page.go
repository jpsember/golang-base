package webserv

import (
	. "github.com/jpsember/golang-base/base"
)

type Page interface {
	Name() string
	Args() []string // The additional arguments that would show up in the url (e.g., edit/17), args would be ["17"]
	// Attempt to construct a new page with the specified args; return nil if args aren't valid
	ConstructPage(s *SessionStruct, args PageArgs) Page
}

// Some common boilerplate that is typically some of the first code that
// generateWidgets() would otherwise execute.
func GenerateHeader(s Session, p Page) WidgetManager {
	var _ = Pr
	CheckState(s != nil)
	m := s.WidgetManager()
	m.With(s.PageWidget)
	return m
}
