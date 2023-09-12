package webserv

import (
	. "github.com/jpsember/golang-base/base"
)

// A Widget that displays editable text
type ListWidgetStruct struct {
	BaseWidgetObj
	list              ListInterface
	renderer          ListItemRenderer
	currentPageNumber int
}

type ListWidgetListener func(sess Session, widget ListWidget) error

type ListWidget = *ListWidgetStruct

func NewListWidget(id string, list ListInterface, renderer ListItemRenderer, listener ListWidgetListener) ListWidget {
	if renderer == nil {
		Alert("<1No renderer for list, using default")
		renderer = defaultRenderer
	}
	w := ListWidgetStruct{
		list:     list,
		renderer: renderer,
	}
	w.BaseId = id
	return &w
}

func (w ListWidget) RenderTo(s Session, m MarkupBuilder) {
	pr := PrIf(false)

	Todo("Add paging controls")
	m.Comment("ListWidget")
	m.OpenTag(`div id="`, w.BaseId, `"`)
	m.OpenTag(`div class="row"`)
	{
		elementIds := w.list.GetPageElements(w.currentPageNumber)
		pr("rendering page num:", w.currentPageNumber, "element ids:", elementIds)
		for _, id := range elementIds {
			m.Comment("--------------------------- rendering id:", id)
			w.renderer(w, id, m)
		}
	}
	m.CloseTag()
	m.CloseTag()
}

func defaultRenderer(widget ListWidget, elementId int, m MarkupBuilder) {
	m.OpenTag(`div class="col-sm-16" style="background-color:` + DebugColor(elementId) + `"`)
	m.WriteString("default list render, Id:" + IntToString(elementId))
	m.CloseTag()
}
