package webserv

import (
	. "github.com/jpsember/golang-base/base"
)

// A Widget that displays editable text
type ListWidgetStruct struct {
	BaseWidgetObj
	list              ListInterface
	currentPageNumber int
}

type ListWidgetListener func(sess Session, widget ListWidget) error

type ListWidget = *ListWidgetStruct

func NewListWidget(id string, list ListInterface, listener ListWidgetListener) ListWidget {
	w := ListWidgetStruct{
		list: list,
	}
	w.BaseId = id
	return &w
}

func (w ListWidget) RenderTo(m MarkupBuilder, state JSMap) {
	Todo("The RenderInvisible could be handled elsewhere")

	pr := PrIf(true)

	m.Comment("ListWidget")
	m.OpenTag(`div id="`, w.BaseId, `"`)
	{
		elementIds := w.list.GetPageElements(w.currentPageNumber)
		pr("page num:", w.currentPageNumber, "element ids:", elementIds)

		for i, id := range elementIds {
			m.OpenTag(`div class="col-sm-16"`)

			m.WriteString("hello, this is : " + IntToString(i) + ", id " + IntToString(id))

			m.CloseTag()
		}

	}
	m.CloseTag()

}
