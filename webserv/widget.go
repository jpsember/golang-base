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
	GetBaseWidget() BaseWidget
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

func (w BaseWidget) RenderTo(m MarkupBuilder) {
	m.A("BaseWidget, id: ")
	m.A(w.Id)
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
	w.GetBaseWidget().Bounds = RectWith(gc.X, gc.Y, gc.Width, 1)
	c.Children.Add(w)
}

func (w ContainerWidget) RenderTo(m MarkupBuilder) {

	desc := `ContainerWidget ` + w.IdSummary()
	m.OpenHtml(`p`, desc).A(desc).CloseHtml(`p`, ``)

	if w.Children.NonEmpty() {
		// We will assume all child views are in grid order
		// We will also assume that they define some number of rows, where each row is completely full
		prevRect := RectWith(-1, -1, 0, 0)
		for _, child := range w.Children.Array() {
			bw := child.GetBaseWidget()
			b := &bw.Bounds
			CheckArg(b.IsValid())
			if b.Location.Y > prevRect.Location.Y {
				if prevRect.Location.Y >= 0 {
					m.CloseHtml("div", "end of row")
					m.Br()
				}
				m.Br()
				m.OpenHtml(`div class="row"`, `start of row`)
				m.Cr()
			}
			prevRect = *b
			m.OpenHtml(`div class="col-sm-`+IntToString(b.Size.W)+`"`, `child`)
			child.RenderTo(m)
			m.CloseHtml(`div`, `child`)
		}
		m.CloseHtml("div", "row")
		m.Br()
	}
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
