package webserv

import (
	. "github.com/jpsember/golang-base/base"
)

// The simplest concrete Widget implementation
type BaseWidgetObj struct {
	BaseId string
	Bounds Rect

	LowListen     LowLevelWidgetListener
	hidden        bool
	disabled      bool
	staticContent any
	idHashcode    int
	size          WidgetSize
	align         WidgetAlign
}

type BaseWidget = *BaseWidgetObj

func NewBaseWidget(id string) BaseWidget {
	t := &BaseWidgetObj{}
	t.BaseId = id
	return t
}

func (w BaseWidget) LowListener() LowLevelWidgetListener {
	return w.LowListen
}

func (w BaseWidget) String() string {
	return "<" + w.BaseId + ">"
}

func (w BaseWidget) Id() string {
	return w.BaseId
}

func (w BaseWidget) SetSize(size WidgetSize) {
	w.size = size
}

func (w BaseWidget) SetAlign(align WidgetAlign) {
	w.align = align
}

func (w BaseWidget) Align() WidgetAlign {
	return w.align
}

func (w BaseWidget) Size() WidgetSize {
	return w.size
}

func (w BaseWidget) Base() BaseWidget {
	return w
}

func (w BaseWidget) Children() []Widget {
	return nil
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

func (w BaseWidget) GetChildren() []Widget {
	return nil
}

func (w BaseWidget) IdSummary() string {
	if w.BaseId == "" {
		return `(no id)`
	}
	return `Id: ` + w.BaseId
}

func (w BaseWidget) RenderTo(s Session, m MarkupBuilder) {
	m.A(`<div id='`, w.BaseId, `'></div>`)
}

func (w BaseWidget) AuxId() string {
	return w.BaseId + ".aux"
}
