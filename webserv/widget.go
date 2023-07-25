package webserv

import (
	. "github.com/jpsember/golang-base/base"
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
	// These should be in the base widget
	SetEnabled(s bool)
	Enabled() bool
}

type WidgetMap = map[string]Widget

const MaxColumns = 12
