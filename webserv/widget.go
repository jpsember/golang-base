package webserv

import (
	"fmt"
	. "github.com/jpsember/golang-base/base"
)

var DebugUIFlag = false

// The interface that all widgets must support.  Widgets can embed the BaseWidget struct to
// supply default implementations.
type Widget interface {
	fmt.Stringer
	Id() string
	LowListener() LowLevelWidgetListener
	Enabled() bool
	Visible() bool
	// This should not be called directly; rather, RenderWidget() to handle invisible widgets properly
	RenderTo(s *SessionStruct, m MarkupBuilder)
	Children() []Widget
	AddChild(c Widget, manager WidgetManager)
	RemoveChild(c Widget)
	AddChildren(manager WidgetManager) // Add any child widgets
	SetColumns(columns int)            // Set the number of columns the widget occupies in its row
	Columns() int                      // Get the number of columns the widget occupies in its row
	StateProvider() WidgetStateProvider
	SetStateProvider(p WidgetStateProvider)
	SetVisible(bool)
	SetTrace(bool)
	Trace() bool
	Log(args ...any) // Logs messages if tracing is set for this widget
}

const WidgetIdPage = "page"

type LowLevelWidgetListener func(sess Session, widget Widget, value string) (optNewWidgetValue any, err error)

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
		Todo("!is it ok to render this as a void tag? Apparently not")
		w.Log("RenderWidget, not visible;")
		i := m.Len()
		m.TgOpen(`div id=`).A(QUOTED, w.Id()).TgContent().TgClose()
		w.Log("Markup:", INDENT, m.String()[i:])
		//m.A(`<div id=`,QUOTED, w.Id(), `></div>`).Cr()
	} else {
		w.RenderTo(s, m)
	}
}
