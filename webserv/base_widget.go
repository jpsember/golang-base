package webserv

import (
	. "github.com/jpsember/golang-base/base"
)

// The simplest concrete Widget implementation
type BaseWidgetObj struct {
	BaseId string
	Bounds Rect

	LowListen     LowLevelWidgetListener
	trace         bool
	hidden        bool
	disabled      bool
	staticContent any
	idHashcode    int
	size          WidgetSize
	align         WidgetAlign
	columns       int
	stateProvider WidgetStateProvider
	//clickListener ClickListener
}

type BaseWidget = *BaseWidgetObj

func (w BaseWidget) Trace() bool { return w.trace }

func (w BaseWidget) SetTrace(flag bool) {
	if flag && !w.trace {
		Alert("#50<1Setting trace on widget:", w.Id())
	}
	w.trace = flag
}

func (w BaseWidget) InitBase(id string) {
	w.BaseId = id
}

func NewBaseWidget(id string) BaseWidget {
	t := &BaseWidgetObj{}
	t.InitBase(id)
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

// Base widgets have no children
func (w BaseWidget) Children() []Widget {
	return nil
}

// Base widgets don't add any children
func (w BaseWidget) AddChildren(m WidgetManager) {
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

func (w BaseWidget) RemoveChild(c Widget) {
	NotSupported("RemoveChild not supported")
}

func (w BaseWidget) SetColumns(columns int) {
	w.columns = columns
}

func (w BaseWidget) Columns() int { return w.columns }

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

func (w BaseWidget) SetStateProvider(p WidgetStateProvider) {
	if w.trace {
		if p != nil {
			w.Log("SetStateProvider:", p, Caller())
		}
	}
	w.stateProvider = p
}

func (w BaseWidget) Log(args ...any) {
	if w.trace {
		Pr("{"+w.Id()+"}: ", ToString(args...))
	}
}

func (w BaseWidget) StateProvider() WidgetStateProvider {
	return w.stateProvider
}

//func (w BaseWidget) GetClickListener() ClickListener {
//	return w.clickListener
//}
//
//func (w BaseWidget) SetClickListener(c ClickListener) {
//	w.clickListener = c
//}
