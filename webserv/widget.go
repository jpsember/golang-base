package webserv

import (
	. "github.com/jpsember/golang-base/base"
)

// The interface that all widgets must support
type Widget interface {
	Base() BaseWidget
	RenderTo(m MarkupBuilder, state JSMap)
	Children() *Array[Widget]
	AddChild(c Widget, manager WidgetManager)
	ClearChildren()
	Id() string
}

// This general type of listener can serve as a validator as well
type WidgetListener func(sess Session, widget Widget)

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

func WidgetErrorCount(root Widget, state JSMap) int {
	count := 0
	return auxWidgetErrorCount(count, root, state)
}

func auxWidgetErrorCount(count int, w Widget, state JSMap) int {
	problemId := WidgetIdWithProblem(w.Base().BaseId)
	if state.OptString(problemId, "") != "" {
		count++
	}
	for _, child := range w.Children().Array() {
		count = auxWidgetErrorCount(count, child, state)
	}
	return count
}

func WidgetIdWithProblem(id string) string {
	CheckArg(id != "")
	return id + ".problem"
}
