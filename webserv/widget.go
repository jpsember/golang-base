package webserv

import (
	"fmt"
	. "github.com/jpsember/golang-base/base"
)

// The interface that all widgets must support
type Widget interface {
	Id() string
	Listener() WidgetListener
	SetListener(WidgetListener)

	Enabled() bool
	RenderTo(m MarkupBuilder, state JSMap)
	Children() *Array[Widget]
	AddChild(c Widget, manager WidgetManager)
	ClearChildren()

	SetStaticContent(content any)
	StaticContent() any

	fmt.Stringer
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

type WidgetAlign int

const (
	AlignDefault WidgetAlign = iota
	AlignCenter
	AlignLeft
	AlignRight
)

func WidgetErrorCount(root Widget, state JSMap) int {
	count := 0
	return auxWidgetErrorCount(count, root, state)
}

func auxWidgetErrorCount(count int, w Widget, state JSMap) int {
	problemId := WidgetIdWithProblem(w.Id())
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
