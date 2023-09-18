package webserv

import (
	"fmt"
	. "github.com/jpsember/golang-base/base"
)

var DebugUIFlag = false

// Function for supplying state to widgets during rendering.
type WidgetStateProvider func(s *SessionStruct, widgetId string) any

// The interface that all widgets must support.  Widgets can embed the BaseWidget struct to
// supply default implementations.
type Widget interface {
	Id() string
	LowListener() LowLevelWidgetListener
	Enabled() bool
	Visible() bool
	// This should not be called directly; rather, RenderWidget() to handle invisible widgets properly
	RenderTo(s *SessionStruct, m MarkupBuilder)
	Children() []Widget
	AddChild(c Widget, manager WidgetManager)
	AddChildren(manager WidgetManager) // Add any child widgets
	SetColumns(columns int)            // Set the number of columns the widget occupies in its row
	Columns() int                      // Get the number of columns the widget occupies in its row
	StateProvider() WidgetStateProvider
	SetStateProvider(p WidgetStateProvider)
	fmt.Stringer
}

const WidgetIdPage = "page"

// This general type of listener can serve as a validator as well
// type WidgetListener func(sess Session, widget Widget)
type LowLevelWidgetListener func(sess Session, widget Widget, value string) (string, error)

type WidgetMap = map[string]Widget

const MaxColumns = 12

type WidgetSize int

const (
	SizeDefault WidgetSize = iota
	SizeMicro
	SizeTiny
	SizeSmall
	SizeMedium
	SizeLarge
	SizeHuge
)

type WidgetAlign int

const (
	AlignDefault WidgetAlign = iota
	AlignCenter
	AlignLeft
	AlignRight
)

func WidgetErrorCount(root Widget, state JSMap) int {
	count := 0
	return auxWidgetErrorCount(count, root, state)
}

func auxWidgetErrorCount(count int, w Widget, state JSMap) int {
	problemId := WidgetIdWithProblem(w.Id())
	if state.OptString(problemId, "") != "" {
		count++
	}
	for _, child := range w.Children() {
		count = auxWidgetErrorCount(count, child, state)
	}
	return count
}

func WidgetIdWithProblem(id string) string {
	CheckArg(id != "")
	return id + ".problem"
}

// Call w.RenderTo(...) iff the widget is visible, otherwise render an empty div with the widget's id.
func RenderWidget(w Widget, s Session, m MarkupBuilder) {
	if !w.Visible() {
		m.A(`<div id='`, w.Id(), `'></div>`).Cr()
	} else {
		w.RenderTo(s, m)
	}
}

func ReadStateStringForId(s Session, w Widget, id string) string {
	Todo("All these global functions are troubling... make this a session function?")
	p := w.StateProvider()
	obj := p(s, id)
	if obj == nil {
		return ""
	}
	switch x := obj.(type) {
	case string:
		return x
	case JSEntity:
		return x.AsString()
	default:
		Alert("#50<1State for widget id", id, "was not a string:", Info(obj))
		return "???"
	}

}

func ReadStateString(s Session, w Widget) string {
	return ReadStateStringForId(s, w, w.Id())
}

func ReadStateBoolean(s Session, w Widget) bool {
	p := w.StateProvider()
	obj := p(s, w.Id())
	if obj == nil {
		return false
	}
	if x, ok := obj.(bool); ok {
		return x
	}
	Alert("#50<1State for widget id", w.Id(), "was not a bool:", Info(obj))
	return false
}

// Reads an integer from a WidgetStateProvider (one that is not necessarily tied to a widget).
func ReadIntFromProvider(s Session, id string, p WidgetStateProvider) int {
	obj := p(s, id)
	if obj == nil {
		return 0
	}
	switch x := obj.(type) {
	case JSEntity:
		return int(x.AsInteger())
	default:
		Alert("#50<1State for widget id", id, "was not a string:", Info(obj))
		return 0
	}

}
