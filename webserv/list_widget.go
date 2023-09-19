package webserv

import (
	. "github.com/jpsember/golang-base/base"
)

// A Widget that displays editable text
type ListWidgetStruct struct {
	BaseWidgetObj
	list              ListInterface
	renderer          ListItemRenderer
	itemStateProvider ListItemStateProvider
	itemWidget        Widget
	pagePrefix        string
	WithPageControls  bool
}

type ListItemStateProvider func(sess Session, widget *ListWidgetStruct, elementId int) (string, JSMap)

type ListWidgetListener func(sess Session, widget ListWidget) error

type ListWidget = *ListWidgetStruct

func NewListWidget(id string, list ListInterface, itemWidget Widget, itemStateProvider ListItemStateProvider, renderer ListItemRenderer, listener ListWidgetListener) ListWidget {

	Todo("Document the fact that widgets that have their own explicit state providers won't use the one for this list, so items might not render as expected")

	Todo("Maybe pass in a widget as the item to be rendered")
	if itemWidget == nil {
		Alert("!lists will require widets instead of renderers soon.")

		if renderer == nil {
			Alert("<1No renderer for list, using default")
			renderer = defaultRenderer
		}
	}
	w := ListWidgetStruct{
		list:              list,
		renderer:          renderer,
		itemWidget:        itemWidget,
		itemStateProvider: itemStateProvider,
		WithPageControls:  true,
	}
	w.InitBase(id)
	w.pagePrefix = id + ".page_"
	return &w
}

func (w ListWidget) renderPagination(s Session, m MarkupBuilder) {
	np := w.list.TotalPages()

	if np < 2 {
		return
	}

	windowSize := MinInt(np, 5)
	windowStart := Clamp(w.list.CurrentPage()-windowSize/2, 0, np-windowSize)
	windowStop := Clamp(windowStart+windowSize, 0, np-1)

	m.OpenTag(`div class="row"`)
	{
		m.OpenTag(`nav aria-label="Page navigation"`)
		{
			m.OpenTag(`ul class="pagination d-flex justify-content-center"`)
			{
				w.renderPagePiece(m, `&lt;&lt;`, 0, true)
				w.renderPagePiece(m, `&lt;`, w.list.CurrentPage()-1, true)

				for i := windowStart; i <= windowStop; i++ {
					w.renderPagePiece(m, IntToString(i+1), i, false)
				}
				w.renderPagePiece(m, `&gt;`, w.list.CurrentPage()+1, true)
				w.renderPagePiece(m, `&gt;&gt;`, np-1, true)
			}
			m.CloseTag()
		}
		m.CloseTag()
	}
	m.CloseTag()
}

func (w ListWidget) renderPagePiece(m MarkupBuilder, label string, targetPage int, edges bool) {
	m.A(`<li class="page-item"><a class="page-link`)
	targetPage = Clamp(targetPage, 0, w.list.TotalPages()-1)
	if w.list.CurrentPage() == targetPage {
		if edges {
			m.A(` disabled`)
		} else {
			m.A(` active`)
		}
	} else {
		m.A(`" onclick="jsButton('`, w.pagePrefix, targetPage, `')`)
	}
	m.A(`">`, label, `</a></li>`, CR)
}

func (w ListWidget) RenderTo(s Session, m MarkupBuilder) {
	pr := PrIf(false)

	oldStyle := w.renderer != nil
	m.Comment("ListWidget")

	m.OpenTag(`div id="`, w.BaseId, `"`)

	if w.WithPageControls {
		w.renderPagination(s, m)
	}

	{
		m.OpenTag(`div class="row"`)
		{
			elementIds := w.list.GetPageElements()
			pr("rendering page num:", w.list.CurrentPage(), "element ids:", elementIds)

			// While rendering this list's items, we will replace any existing default state provider with
			// the list's one.  Save the current default state provider here, for later restoration.
			savedStateProvider := s.DefaultStateProvider

			for _, id := range elementIds {
				m.Comment("--------------------------- rendering id:", id)

				if oldStyle {
					w.renderer(s, w, id, m)
					continue
				}
				// Get the client to return a state provider
				prefix, jsmap := w.itemStateProvider(s, w, id)
				Pr("got list item state provider, map:", jsmap)
				// Make it the default state provider.
				sp := NewStateProvider(prefix, jsmap)
				s.DefaultStateProvider = sp
				Todo("the state provider struct could be built earlier and initialized here each time")

				w.itemWidget.RenderTo(s, m)
			}
			// Restore the default state provider to what it was before we rendered the items.
			s.DefaultStateProvider = savedStateProvider
		}
		m.CloseTag()
	}

	if w.WithPageControls {
		w.renderPagination(s, m)
	}

	m.CloseTag()
}

// Parse a click event, and if it is aimed at us, process it and return true.  This is used by the
// pagination controls.  **Clicks on the list items are still handled by the client.**
// This
func (w ListWidget) HandleClick(sess Session, message string) bool {
	if page_str, f := TrimIfPrefix(message, w.pagePrefix); f {
		for {
			i, err := ParseInt(page_str)
			if ReportIfError(err, "handling click:", message) {
				break
			}
			targetPage := int(i)
			if targetPage < 0 || targetPage >= w.list.TotalPages() {
				Alert("#50illegal page requested;", message)
				break
			}

			if targetPage == w.list.CurrentPage() {
				break
			}
			w.list.SetCurrentPage(targetPage)
			sess.Repaint(w)
			break
		}
		return true
	}
	return false
}

func defaultRenderer(session Session, widget ListWidget, elementId int, m MarkupBuilder) {
	m.OpenTag(`div class="col-sm-16" style="background-color:` + DebugColor(elementId) + `"`)
	m.WriteString("default list render, Id:" + IntToString(elementId))
	m.CloseTag()
}
