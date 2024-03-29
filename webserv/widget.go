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
	SetLowListener(listener LowLevelWidgetListener)
	// Get value as a string, or return false if no validation required. This should be the same value that
	// can be passed to the LowLevelWidgetListener.  This allows us to validate the widget by simulating
	// what any listeners would have done with values received via AJAX calls.
	ValidationValue(s *SessionStruct) (string, bool)
	Enabled() bool
	Visible() bool
	Detached() bool
	// This should not be called directly; rather, RenderWidget() to handle invisible and detached widgets properly
	RenderTo(s *SessionStruct, m MarkupBuilder)
	Children() []Widget
	AddChild(c Widget, manager WidgetManager)
	RemoveChild(c Widget)
	AddChildren(manager WidgetManager) // Add any child widgets
	SetColumns(columns int)            // Set the number of columns the widget occupies in its row
	Columns() int                      // Get the number of columns the widget occupies in its row
	StateProvider() JSMap
	setStateProvider(p JSMap)
	SetVisible(bool)
	SetDetached(bool)
	SetTrace(bool)
	Repaint()
	ClearRepaint()
	IsRepaint() bool
	Trace() bool
	Log(args ...any) // Logs messages if tracing is set for this widget
}

const WidgetIdPage = "page"

type LowLevelWidgetListener func(sess Session, widget Widget, value string, args WidgetArgs) (optNewWidgetValue any, err error)

type ClickWidgetListener func(sess *SessionStruct, widget Widget, args WidgetArgs)

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

func widgetProblemKey(widgetId string) string {
	return widgetId + ".problem"
}

// Call w.RenderTo(...) iff the widget is visible, otherwise render an empty div with the widget's id.
func RenderWidget(w Widget, s Session, m MarkupBuilder) {
	if w.Detached() {
		return
	}
	if !w.Visible() {
		w.Log("RenderWidget, not visible;")
		i := m.Len()
		m.TgOpen(`div id=`).A(QUO, s.PrependId(w.Id())).TgContent().TgClose()
		w.Log("Markup:", INDENT, m.String()[i:])
	} else {
		w.RenderTo(s, m)
	}
}
