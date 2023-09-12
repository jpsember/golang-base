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

func (w ListWidget) renderPagination(s Session, m MarkupBuilder) {
	np := w.list.TotalPages()

	if np < 2 {
		return
	}

	m.Comment("Rendering pagination, number of pages:", np)
	pagePrefix := w.Id() + ".page_"

	windowSize := MinInt(np, 5)
	windowStart := Clamp(w.list.CurrentPage()-windowSize/2, 0, np-windowSize+1)

	m.OpenTag(`div class="row"`)
	{
		m.OpenTag(`nav aria-label="Page navigation"`)
		{
			m.OpenTag(`ul class="pagination"`)

			{

				// "Previous"
				m.A(`<li class="page-item"><a class="page-link`)
				if w.list.CurrentPage() == 0 {
					m.A(` disabled`)
				} else {
					m.A(`" onclick="jsButton('`, pagePrefix, `0')"`)
				}
				m.A(`">&lt;&lt;</a></li>`, CR)

				// Window elements
				{
					for i := windowStart; i < windowStart+windowSize; i++ {
						m.A(`<li class="page-item"><a class="page-link`)
						if w.list.CurrentPage() == i {
							m.A(` active`)
						}
						m.A(`" onclick="jsButton('`, pagePrefix, i, `')"`)
						m.A(`>`, 1+i, `</a></li>`).Cr()
					}
				}

				// "Next"
				//

				{
					m.A(`<li class="page-item"><a class="page-link`)
					if w.list.CurrentPage() == np-1 {
						m.A(` disabled`)
					} else {
						m.A(`" onclick="jsButton('`, pagePrefix, np-1, `')`)
					}
					m.A(`">&gt;&gt;</a></li>`, CR)
				}

			}
			m.CloseTag()
		}
		m.CloseTag()
	}
	m.CloseTag()
	m.Comment("done Rendering pagination")
}

func (w ListWidget) RenderTo(s Session, m MarkupBuilder) {
	pr := PrIf(false)

	m.Comment("ListWidget")

	m.OpenTag(`div id="`, w.BaseId, `"`)

	w.renderPagination(s, m)

	if !Alert("skipping items") {
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
	}

	m.CloseTag()
}

func defaultRenderer(widget ListWidget, elementId int, m MarkupBuilder) {
	m.OpenTag(`div class="col-sm-16" style="background-color:` + DebugColor(elementId) + `"`)
	m.WriteString("default list render, Id:" + IntToString(elementId))
	m.CloseTag()
}
