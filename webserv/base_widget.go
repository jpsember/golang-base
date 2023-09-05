package webserv

import (
	. "github.com/jpsember/golang-base/base"
)

// The simplest concrete Widget implementation
type BaseWidgetObj struct {
	BaseId        string
	Bounds        Rect
	listener      WidgetListener
	hidden        bool
	disabled      bool
	staticContent any
	idHashcode    int
}

type BaseWidget = *BaseWidgetObj

func NewBaseWidget(id string) BaseWidget {
	t := &BaseWidgetObj{}
	t.BaseId = id
	return t
}

func (w BaseWidget) String() string {
	return "<" + w.BaseId + ">"
}

func (w BaseWidget) Id() string {
	return w.BaseId
}

func (w BaseWidget) Listener() WidgetListener {
	return w.listener
}

func (w BaseWidget) SetListener(listener WidgetListener) {
	if w.listener != nil {
		BadState("Widget", w.Id(), "already has a listener")
	}
	w.listener = listener
}

func (w BaseWidget) ClearChildren() {
	NotImplemented()
}

func (w BaseWidget) Base() BaseWidget {
	return w
}

var emptyChildrenList = NewArray[Widget]().Lock()

func (w BaseWidget) Children() *Array[Widget] {
	return emptyChildrenList
}

func (w BaseWidget) SetStaticContent(content any) {
	w.staticContent = content
}

func (w BaseWidget) StaticContent() any {
	return w.staticContent
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

func (w BaseWidget) AddChild(c Widget, manager WidgetManager) {
	NotSupported("AddChild not supported")
}

func (w BaseWidget) ReceiveValue(sess Session, value string) {
	Pr("Ignoring ReceiveValue for widget:", w.BaseId, "value:", Quoted(value))
}

var emptyWidgetList = make([]Widget, 0)

func (w BaseWidget) GetChildren() []Widget {
	return emptyWidgetList
}

func (w BaseWidget) IdSummary() string {
	if w.BaseId == "" {
		return `(no id)`
	}
	return `Id: ` + w.BaseId
}

func (w BaseWidget) RenderTo(m MarkupBuilder, state JSMap) {
	m.A(`<div id='`, w.BaseId, `'></div>`)
}

func (w BaseWidget) AuxId() string {
	return w.BaseId + ".aux"
}
