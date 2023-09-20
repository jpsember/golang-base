package webserv

import (
	. "github.com/jpsember/golang-base/base"
	"strings"
)

// A Widget that displays editable text
type ListWidgetStruct struct {
	BaseWidgetObj
	list              ListInterface
	itemStateProvider ListItemStateProvider
	itemWidget        Widget
	pagePrefix        string
	WithPageControls  bool
	Listener          ListWidgetListener
}

type ListWidgetListener func(sess Session, widget *ListWidgetStruct, itemId int, args string)

func listListenWrapper(sess Session, widget Widget, value string) (string, error) {
	b := widget.(ListWidget)
	Pr("listListenWrapper, value:", value)
	Todo("parse args")

	// This is presumably something like <element id> '.' <remainder>
	itemId := -1
	c := strings.IndexByte(value, '.')
	remainder := ""
	if c > 0 {
		remainder = value[c+1:]
		val, err := ParseInt(value[0:c])
		if err != nil {
			Alert("#50 trouble parsing int from:", value)
		} else {
			itemId = int(val)
		}
	}

	if b.Listener == nil {
		Alert("#50No ListListener registered; itemId:", itemId, "args:", remainder)
	} else {
		b.Listener(sess, b, itemId, remainder)
	}
	return "", nil
}

type ListItemStateProvider func(sess Session, widget *ListWidgetStruct, elementId int) WidgetStateProvider

type ListWidget = *ListWidgetStruct

func NewListWidget(id string, list ListInterface, itemWidget Widget, itemStateProvider ListItemStateProvider) ListWidget {
	Todo("Document the fact that widgets that have their own explicit state providers won't use the one for this list, so items might not render as expected")
	Todo("If no item widget has been supplied, construct a default one")
	CheckArg(itemWidget != nil, "No itemWidget given")

	w := ListWidgetStruct{
		list:              list,
		itemWidget:        itemWidget,
		itemStateProvider: itemStateProvider,
		WithPageControls:  true,
	}
	w.InitBase(id)
	w.LowListen = listListenWrapper
	w.pagePrefix = id + ".page_"

	return &w
}

func (w ListWidget) ItemWidget() Widget {
	return w.itemWidget
}

func (w ListWidget) renderPagination(s Session, m MarkupBuilder) {
	np := w.list.TotalPages()

	if np < 2 {
		return
	}

	Todo("Who is causing a lot of comments here?")
	windowSize := MinInt(np, 5)
	windowStart := Clamp(w.list.CurrentPage()-windowSize/2, 0, np-windowSize)
	windowStop := Clamp(windowStart+windowSize, 0, np-1)

	m.OpenTag(`div class="row"`)
	{
		m.OpenTag(`nav aria-label="Page navigation"`)
		{
			m.OpenTag(`ul class="pagination d-flex justify-content-center"`)
			{
				w.renderPagePiece(s, m, `&lt;&lt;`, 0, true)
				w.renderPagePiece(s, m, `&lt;`, w.list.CurrentPage()-1, true)

				for i := windowStart; i <= windowStop; i++ {
					w.renderPagePiece(s, m, IntToString(i+1), i, false)
				}
				w.renderPagePiece(s, m, `&gt;`, w.list.CurrentPage()+1, true)
				w.renderPagePiece(s, m, `&gt;&gt;`, np-1, true)
			}
			m.CloseTag()
		}
		m.CloseTag()
	}
	m.CloseTag()
}

func (w ListWidget) renderPagePiece(s Session, m MarkupBuilder, label string, targetPage int, edges bool) {

	m.A(`<li class="page-item"><a class="page-link`)
	targetPage = Clamp(targetPage, 0, w.list.TotalPages()-1)
	if w.list.CurrentPage() == targetPage {
		if edges {
			m.A(` disabled`)
		} else {
			m.A(` active`)
		}
	} else {
		m.A(`" onclick="jsButton('`, s.baseIdPrefix+w.pagePrefix, targetPage, `')`)
	}
	m.A(`">`, label, `</a></li>`, CR)
}

func (w ListWidget) RenderTo(s Session, m MarkupBuilder) {
	pr := PrIf(false)
	pr("ListWidget.RenderTo")
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
			savedStateProvider := s.baseStateProvider
			savedBaseIdPrefix := s.baseIdPrefix
			for _, id := range elementIds {
				m.Comment("----------------- rendering list item with id:", id)

				// Get the client to return a state provider
				s.baseStateProvider = w.itemStateProvider(s, w, id)

				// When rendering list items, any ids should be mangled in such a way that
				//  a) ids remain distinct, even if we are rendering the same widget for each row; and
				//  b) when responding to click events and the like, we can figure out which list, and
				//      element within the list, generated the event.
				s.baseIdPrefix = w.Id() + "." + IntToString(id) + "." + savedBaseIdPrefix
				w.itemWidget.RenderTo(s, m)
			}
			// Restore the default state provider to what it was before we rendered the items.
			s.baseIdPrefix = savedBaseIdPrefix
			s.baseStateProvider = savedStateProvider
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
