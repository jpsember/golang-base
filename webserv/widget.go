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
		//sb.A(`<div class="row">`)
		for _, child := range w.Children.Array() {
			bw := child.GetBaseWidget()
			b := &bw.Bounds
			CheckArg(b.IsValid())
			//&child.Bounds
			if b.Location.Y > prevRect.Location.Y {
				if prevRect.Location.Y >= 0 {
					m.Cr()
					m.CloseHtml("div", "end of row")
				}
				m.Cr()
				m.OpenHtml(`div class="row"`, `start of row`)
				m.Cr()
			}
			prevRect = *b
			m.OpenHtml(`div class="col-sm-`+IntToString(b.Size.W)+`"`, `child`)
			child.RenderTo(m)
			//renderViewHelper(sess, sb, child)
			m.CloseHtml(`div`, `child`)
		}
		m.Cr()
		m.CloseHtml("div", "row")
		m.Cr()
	}

	Pr("done render to")
	//
	//
	//
	//m.A(`<div class="row">`).Cr()
	//
	//
	//
	//for _, c := range w.Children.Array() {
	//	m.A(`<div class="col-sm">`).Cr()
	//	c.RenderTo(m)
	//	m.CloseHtml("div", "child")
	//}
	//m.CloseHtml("div", "row")
	//m.CloseHtml("div", "container")
}

//func RenderChildren(children *Array[Widget], m MarkupBuilder) {
//
//	//// We need to keep track of whether we are rendering a row of more than one view
//	//wrapInCol := view.Bounds.Size.W != 12
//	//if wrapInCol {
//	//	sb.A(`<div class="col-sm-`)
//	//	sb.A(strconv.Itoa(view.Bounds.Size.W))
//	//	sb.A(`">`)
//	//}
//
//	if children.NonEmpty() {
//		// We will assume all child views are in grid order
//		// We will also assume that they define some number of rows, where each row is completely full
//		prevRect := RectWith(-1, -1, 0, 0)
//		//sb.A(`<div class="row">`)
//		for _, child := range children.Array() {
//			b := &child.Bounds
//			if b.Location.Y > prevRect.Location.Y {
//				if prevRect.Location.Y >= 0 {
//					sb.CloseHtml("div", "row")
//				}
//				sb.OpenHtml(`div class="row"`, ``)
//			}
//			prevRect = *b
//			sb.OpenHtml(`div class="col-sm-`+IntToString(b.Size.W), `child`)
//			renderViewHelper(sess, sb, child)
//			sb.CloseHtml(`div`, `child`)
//		}
//		sb.CloseHtml("div", "row")
//	}
//
//
//}

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
