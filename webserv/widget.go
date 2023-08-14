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

// This general type of listener can serve as a validator as well
type WidgetListener func(sess Session, widget Widget) error

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

func WidgetId(widget Widget) string {
	return widget.Base().Id
}

func WidgetErrorCount(root Widget, state JSMap) int {
	count := 0
	return auxWidgetErrorCount(count, root, state)
}

func auxWidgetErrorCount(count int, w Widget, state JSMap) int {
	problemId := w.Base().Id + ".problem"
	if state.OptString(problemId, "") != "" {
		count++
	}
	for _, child := range w.GetChildren() {
		count = auxWidgetErrorCount(count, child, state)
	}
	return count
}
