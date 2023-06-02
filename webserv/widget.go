package webserv

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/json"
)

type ClientValueObj struct {
	values  []string
	problem string
}

type ClientValue = *ClientValueObj

func MakeClientValue(values []string) ClientValue {
	c := ClientValueObj{
		values: values,
	}
	return &c
}

func (c ClientValue) SetProblem(message ...any) ClientValue {
	if c.problem == "" {
		c.problem = "Problem with ajax request: " + ToString(message...)
	}
	return c
}

func (c ClientValue) GetString() string {
	if c.problem == "" {
		if len(c.values) == 1 {
			return c.values[0]
		}
		c.SetProblem("Expected single string")
	}
	return ""
}

// The interface that all widgets must support
type Widget interface {
	GetId() string
	WriteValue(v JSEntity)
	ReadValue() JSEntity
	RenderTo(m MarkupBuilder, state JSMap)
	GetBaseWidget() BaseWidget
	AddChild(c Widget, manager WidgetManager)
	LayoutChildren(manager WidgetManager)
	GetChildren() []Widget
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
