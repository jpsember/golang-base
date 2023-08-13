package webserv

import (
	. "github.com/jpsember/golang-base/base"
)

// The interface that all widgets must support
type Widget interface {
	Base() BaseWidget
	RenderTo(m MarkupBuilder, state JSMap)
	AddChild(c Widget, manager WidgetManager)
	GetChildren() []Widget
}

type WidgetMap = map[string]Widget

const MaxColumns = 12

type WidgetSize int

const (
	SizeDefault WidgetSize = iota
	SizeMicro
	SizeTiny
	SizeSmall
	SizeMedium
	SizeLarge
	SizeHuge
)
