package webserv

import (
	. "github.com/jpsember/golang-base/base"
)

type Page interface {
	// Note: go doesn't support covariant return types, so this must return Page, not some concrete implementation of it
	Name() string
	Args() []any // The additional arguments that would show up in the url (e.g., edit/17), args would be [17]
	Session() Session
	// Attempt to construct a new page with the specified args; return nil if args aren't valid
	Construct(s Session, args ...any) Page
	Generate()
}

var EmptyPageArgs = []any{}

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

func asInt(arg any) (int, error) {
	switch k := arg.(type) {
	case string:
		return ParseInt2(k)
	case int:
		return k, nil
	default:
		return -1, NotIntError
	}

}

var NotIntError = Error("Not an integer")

func ParsePageIntArg(args []any, index int) int {
	if PageArgExists(args, index) {
		val := args[index]
		intVal, _ := asInt(val)
		return intVal
	}
	return -1
}

func PageArgExists(args []any, index int) bool {
	return index >= 0 && index < len(args)
}
