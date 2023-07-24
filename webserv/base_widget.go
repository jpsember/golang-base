package webserv

import (
	. "github.com/jpsember/golang-base/base"
)

type WidgetListener func(sess any, widget Widget)

// The simplest concrete Widget implementation
type BaseWidgetObj struct {
	Id       string
	Bounds   Rect
	Listener WidgetListener
	hidden   bool
}

type BaseWidget = *BaseWidgetObj

func (w BaseWidget) GetBaseWidget() BaseWidget {
	return w
}

func (w BaseWidget) WriteValue(v JSEntity) {
	NotImplemented("WriteValue")
}

func (w BaseWidget) SetVisible(v bool) {
	w.hidden = !v
}

func (w BaseWidget) ReadValue() JSEntity {
	NotImplemented("ReadValue")
	return JBoolFalse
}

func (w BaseWidget) AddChild(c Widget, manager WidgetManager) {
	NotSupported("AddChild not supported")
}

func (w BaseWidget) LayoutChildren(manager WidgetManager) {
	NotSupported("LayoutChildren not supported")
}

func (w BaseWidget) ReceiveValue(sess Session, value string) {
	Pr("Ignoring ReceiveValue for widget:", w.Id, "value:", Quoted(value))
}

var emptyWidgetList = make([]Widget, 0)

func (w BaseWidget) GetChildren() []Widget {
	return emptyWidgetList
}

func (w BaseWidget) IdSummary() string {
	if w.Id == "" {
		return `(no id)`
	}
	return `Id: ` + w.Id
}

func (w BaseWidget) IdComment() string {
	return WrapWithinComment(w.IdSummary())
}

func (w BaseWidget) GetId() string {
	return w.Id
}

func (w BaseWidget) RenderTo(m MarkupBuilder, state JSMap) {
	m.A("BaseWidget, id: ")
	m.A(w.Id)
}

func (w BaseWidget) Visible() bool {
	return !w.hidden
}
