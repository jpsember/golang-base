package webserv

import (
	"fmt"
	. "github.com/jpsember/golang-base/base"
)

var DebugUIFlag = false

// The interface that all widgets must support
type Widget interface {
	Id() string
	LowListener() LowLevelWidgetListener
	Enabled() bool
	Visible() bool
	// This should not be called directly; rather, RenderWidget() to handle invisible widgets properly
	RenderTo(s *SessionStruct, m MarkupBuilder)
	Children() []Widget
	AddChild(c Widget, manager WidgetManager)

	SetStaticContent(content any)
	StaticContent() any

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
