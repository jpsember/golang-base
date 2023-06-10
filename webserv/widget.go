package webserv

import (
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
	GetChildren() []Widget
}

type WidgetMap = map[string]Widget

const MaxColumns = 12
