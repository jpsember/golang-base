package webserv

import (
	. "github.com/jpsember/golang-base/base"
)

// The simplest concrete Widget implementation
type BaseWidgetObj struct {
	baseId string

	LowListen     LowLevelWidgetListener
	staticContent any
	size          WidgetSize
	align         WidgetAlign
	columns       int
	stateProvider JSMap
	bitFlags      int
}

const (
	wflagHidden = 1 << iota
	wflagTrace
	wflagDisabled
	wflagDetached
	wflagRepaint
)

type BaseWidget = *BaseWidgetObj

func (w BaseWidget) isFlag(flag int) bool {
	return (w.bitFlags & flag) != 0
}

func (w BaseWidget) ClearRepaint() {
	w.setOrClearFlag(wflagRepaint, false)
}

func (w BaseWidget) Repaint() {
	changed := w.setOrClearFlag(wflagRepaint, true)
	if DebugWidgetRepaint && changed {
		Pr("Repaint:", w.Id(), INDENT, Callers(1, 3))
	}
}
func (w BaseWidget) IsRepaint() bool {
	return w.isFlag(wflagRepaint)
}

func (w BaseWidget) Trace() bool { return w.isFlag(wflagTrace) }
func (w BaseWidget) setOrClearFlag(flag int, set bool) bool {
	oldVal := w.bitFlags
	newVal := oldVal
	if set {
		newVal |= flag
	} else {
		newVal &= ^flag
	}
	w.bitFlags = newVal
	return newVal != oldVal
}

func (w BaseWidget) SetTrace(flag bool) {
	changed := w.setOrClearFlag(wflagTrace, flag)
	if flag && changed {
		Alert("#50<1Setting trace on widget:", w.Id())
	}
}

func (w BaseWidget) InitBase(id string) {
	w.baseId = id
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
	return "<" + w.Id() + ">"
}

func (w BaseWidget) ValidationValue(s Session) (string, bool) {
	return "", false
}

func (w BaseWidget) Id() string {
	return w.baseId
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
	return !w.isFlag(wflagHidden)
}

func (w BaseWidget) Detached() bool {
	return w.isFlag(wflagDetached)
}

func (w BaseWidget) SetVisible(v bool) {
	Todo("?These flags don't cause widget to be plotted, which they ought to be (if their status is changing)")
	w.setOrClearFlag(wflagHidden, !v)
}

func (w BaseWidget) SetDetached(v bool) {
	w.setOrClearFlag(wflagDetached, v)
}

func (w BaseWidget) Enabled() bool {
	return !w.isFlag(wflagDisabled)
}

func (w BaseWidget) SetEnabled(s bool) {
	w.setOrClearFlag(wflagDisabled, !s)
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
	if w.Id() == "" {
		return `(no id)`
	}
	return `Id: ` + w.Id()
}

func (w BaseWidget) RenderTo(s Session, m MarkupBuilder) {
	m.A(`<div id='`, w.Id(), `'></div>`)
}

func (w BaseWidget) AuxId() string {
	return w.Id() + ".aux"
}

func (w BaseWidget) setStateProvider(p JSMap) {
	if w.isFlag(wflagTrace) {
		if p != nil {
			w.Log("setStateProvider:", p, Caller())
		}
	}
	w.stateProvider = p
}

func (w BaseWidget) Log(args ...any) {
	if w.isFlag(wflagTrace) {
		Pr("{"+w.Id()+"}: ", ToString(args...))
	}
}

func (w BaseWidget) StateProvider() JSMap {
	return w.stateProvider
}
