package webserv

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/json"
	"unsafe"
)

// The interface that all widgets must support
type Widget interface {
	GetId() string
	WriteValue(v JSEntity)
	ReadValue() JSEntity
}

// The simplest concrete Widget implementation
type BaseWidgetObj struct {
	Id string
}

type BaseWidget = *BaseWidgetObj

func (w BaseWidget) WriteValue(v JSEntity) {
	NotImplemented("WriteValue")
}

func (w BaseWidget) ReadValue() JSEntity {
	NotImplemented("ReadValue")
	return JBoolFalse
}

func (w BaseWidget) GetId() string {
	return w.Id
}

// A concrete Widget that can contain others
type ContainerWidgetObj struct {
	BaseWidgetObj
}

type ContainerWidget = *ContainerWidgetObj

func NewContainerWidget() ContainerWidget {
	w := ContainerWidgetObj{}
	return &w
}

func (c ContainerWidget) AddChild(w Widget, gc GridCell) {
	Todo("add child to container")
}

func Verify() {
	var w Widget

	c := ContainerWidgetObj{}
	w = &c

	Pr("size of c:", unsafe.Sizeof(c))

	Pr("size of w:", unsafe.Sizeof(w))
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
