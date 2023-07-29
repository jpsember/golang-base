package webserv

import (
	. "github.com/jpsember/golang-base/base"
)

type WidgetListener func(sess any, widget Widget)

// The simplest concrete Widget implementation
type BaseWidgetObj struct {
	Id            string
	Bounds        Rect
	Listener      WidgetListener
	hidden        bool
	disabled      bool
	staticContent any
	idHashcode    int
}

type BaseWidget = *BaseWidgetObj

func (w BaseWidget) GetBaseWidget() BaseWidget {
	return w
}

func (w BaseWidget) SetStaticContent(content any) {
	w.staticContent = content
}

func (w BaseWidget) StaticContent() any {
	return w.staticContent
}

func (w BaseWidget) WriteValue(v JSEntity) {
	NotImplemented("WriteValue")
}

func (w BaseWidget) Visible() bool {
	return !w.hidden
}

func (w BaseWidget) SetVisible(v bool) {
	w.hidden = !v
}

func (w BaseWidget) Enabled() bool {
	return !w.disabled
}

func (w BaseWidget) SetEnabled(s bool) {
	w.disabled = !s
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

// Deprecated.  Just use w.Id
func (w BaseWidget) GetId() string {
	return w.Id
}

func (w BaseWidget) RenderTo(m MarkupBuilder, state JSMap) {
	m.A("No RenderTo() for id: ")
	m.A(w.Id)
}

func (w BaseWidget) AuxId() string {
	return w.GetId() + ".aux"
}

func (w BaseWidget) IdHashcode() int {
	if w.idHashcode == 0 {
		b := []byte(w.Id)
		sum := 0
		for _, x := range b {
			sum += int(x)
		}
		w.idHashcode = MaxInt(sum&0xffff, 1)
	}
	return w.idHashcode
}
