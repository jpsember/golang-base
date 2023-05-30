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
	RenderTo(m MarkupBuilder)
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

func (w BaseWidget) RenderTo(m MarkupBuilder) {
	Todo("I don't think the base widget should include its div; maybe the container should handle that?")
	m.A("BaseWidget, id: ")
	m.A(w.Id)
	//m.A(`<div id="`)
	//m.A(w.Id)
	//m.A(`">BaseWidget, id: `)
	//m.A(w.Id)
	//m.A(`</div>\n`)
}

// A concrete Widget that can contain others
type ContainerWidgetObj struct {
	BaseWidgetObj
	Children *Array[Widget]
}

type ContainerWidget = *ContainerWidgetObj

func NewContainerWidget() ContainerWidget {
	w := ContainerWidgetObj{
		Children: NewArray[Widget](),
	}
	return &w
}

func (c ContainerWidget) AddChild(w Widget, gc GridCell) {
	Pr("adding child widget, grid cell:", gc.X, gc.Y)
	c.Children.Add(w)
}

func (w ContainerWidget) RenderTo(m MarkupBuilder) {

	desc := `ContainerWidget ` + w.IdComment()
	m.OpenHtml(`p`, desc).A(desc).CloseHtml(`p`, ``)

	m.A(`<div class="row">`).CR()

	for _, c := range w.Children.Array() {
		m.A(`<div class="col-sm">`).CR()
		c.RenderTo(m)
		m.CloseHtml("div", "child")
	}
	m.CloseHtml("div", "row")
	m.CloseHtml("div", "container")
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
