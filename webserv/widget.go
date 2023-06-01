package webserv

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/json"
)

// The interface that all widgets must support
type Widget interface {
	GetId() string
	WriteValue(v JSEntity)
	ReadValue() JSEntity
	RenderTo(m MarkupBuilder, state JSMap)
	GetBaseWidget() BaseWidget
	AddChild(c Widget, manager WidgetManager)
	LayoutChildren(manager WidgetManager)
}

// The simplest concrete Widget implementation
type BaseWidgetObj struct {
	Id     string
	Bounds Rect
}

type BaseWidget = *BaseWidgetObj

func (w BaseWidget) GetBaseWidget() BaseWidget {
	return w
}

func (w BaseWidget) WriteValue(v JSEntity) {
	NotImplemented("WriteValue")
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

type LabelWidgetObj struct {
	BaseWidgetObj
	LineCount  int
	Text       string
	Size       int
	Monospaced bool
	Alignment  int
}

type LabelWidget = *LabelWidgetObj

func NewLabelWidget() LabelWidget {
	return &LabelWidgetObj{}
}

type PanelWidgetObj struct {
	BaseWidgetObj
}

type PanelWidget = *PanelWidgetObj

func NewPanelWidget() PanelWidget {
	return &PanelWidgetObj{}
}
