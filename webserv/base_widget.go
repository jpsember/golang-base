package webserv

import (
	. "github.com/jpsember/golang-base/base"
)

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

func NewBaseWidget(id string) BaseWidget {
	t := &BaseWidgetObj{}
	t.Id = id
	return t
}

func (w BaseWidget) Base() BaseWidget {
	return w
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

func (w BaseWidget) RenderTo(m MarkupBuilder, state JSMap) {
	m.A(`<div id='`, w.Id, `'></div>`)
}

func (w BaseWidget) AuxId() string {
	return w.Id + ".aux"
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
