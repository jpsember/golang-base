package webserv

import (
	. "github.com/jpsember/golang-base/base"
)

type ListWidgetStruct struct {
	BaseWidgetObj
	list                 ListInterface
	itemWidget           Widget
	pagePrefix           string
	WithPageControls     bool
	cachedStateProviders map[int]JSMap
	itemPrefix           string
	listItemListener     ListWidgetListener
}
type ListWidget = *ListWidgetStruct

type ListWidgetListener func(sess Session, widget *ListWidgetStruct, elementId int, args []string) error

// Construct a ListWidget.
//
// itemWidget : this is a widget that will be rendered for each displayed item
func NewListWidget(id string, list ListInterface, itemWidget Widget, listener ListWidgetListener) ListWidget {
	Todo("!Have option to wrap list items in a clickable div")
	CheckArg(itemWidget != nil, "No itemWidget given")
	w := ListWidgetStruct{
		list:             list,
		itemWidget:       itemWidget,
		WithPageControls: true,
		listItemListener: listener,
	}
	w.InitBase(id)
	w.itemPrefix = id + ":"
	w.SetLowListener(w.listListenWrapper)
	w.pagePrefix = id + ".page_"

	// If there's an item listener, add it to the item widget; this is so when the item is rendered, it ends up
	// calling the ListWidget's listener
	if listener != nil {
		itemWidget.SetLowListener(w.listListenWrapper)
	}

	return &w
}

func (w ListWidget) RenderTo(s Session, m MarkupBuilder) {
	debug := false
	pr := PrIf("ListWidget.RenderTo", debug)
	pr("ListWidget.RenderTo; id", QUO, w.Id(), "itemPrefix:", QUO, w.itemPrefix)

	m.TgOpen(`div id=`).A(QUO, w.Id()).TgContent()

	m.Comment("ListWidget")

	// Discard any previously cached state providers, and
	// cache those we are about to construct (so we don't ask client
	// to construct them unnecessarily).
	w.cachedStateProviders = make(map[int]JSMap)
	if w.WithPageControls {
		w.renderPagination(s, m)
	}

	m.TgOpen(`div class="row"`).TgContent()
	{
		elementIds := w.list.GetPageElements()
		pr("rendering page #:", w.list.CurrentPage(), "element ids:", elementIds)

		// While rendering this list's items, we will set the state provider to one for each item.
		pr("item prefix:", w.itemPrefix)

		CheckState(len(s.stack) == 2, "Expected two items on state stack: default item, plus one for list renderer")

		for _, id := range elementIds {

			elementIdStr := w.itemPrefix + IntToString(id) + ":"

			// We want each rendered widget to have a unique id, so include "<element id>:" as a *rendering* prefix

			Alert("!Calling PushIdPrefix to include prefix with each of the list item's widgets")
			s.PushIdPrefix(elementIdStr)

			sp := w.constructStateProvider(s, id)
			//sp = NewStateProvider(elementIdStr, sp.State)

			pr(VERT_SP, "pushing state provider:", sp)
			s.PushStateMap(sp)

			// If we push the state provider AFTER the id prefix, it doesn't work! Why?
			// Note that we are not calling RenderWidget(), which would not draw anything since the
			// list item widget has been marked as detached

			// Periods are used to separate widget id from context
			Todo("Update the click prefix; do we even need it?")
			s.PushClickPrefix(elementIdStr)

			if debug {
				pr("stacked state:", INDENT, s.StateStackToJson())
			}
			x := m.Len()
			w.itemWidget.RenderTo(s, m)
			pr("rendered item, markup:", INDENT, m.String()[x:])

			s.PopClickPrefix()

			s.PopStateMap()

			s.PopIdPrefix()
		}
	}
	m.TgClose()

	if w.WithPageControls {
		w.renderPagination(s, m)
	}

	m.TgClose()
}

func (w ListWidget) listListenWrapper(sess Session, widget Widget, value string, args []string) (any, error) {
	pr := PrIf("list_widget.LowLevel listener", false)
	pr(VERT_SP, "value:", QUO, value, "args:", args, "caller:", Caller())
	pr("stack size:", len(sess.stack))
	b := widget.(ListWidget)

	// See if this is an event from the page controls
	if b.handlePagerClick(sess, value) {
		pr("...page controls handled it")
		return nil, nil
	}

	var elementId int
	elementId, remainder, err := ExtractIntFromListenerArgs(args, 0, -1)
	if err == nil {
		if w.listItemListener == nil {
			err = Error("No list item listener for widget", QUO, w.Id())
		} else {
			err = w.listItemListener(sess, w, elementId, remainder)
		}
	}
	return nil, err

}

func (w ListWidget) constructStateProvider(s Session, elementId int) JSMap {
	pr := PrIf("list_widget.constructStateProvider", false)
	cached := w.cachedStateProviders[elementId]
	if cached == nil {
		pv := w.list.ItemStateProvider(s, elementId)
		cached = pv
		pr("constructed:", cached)
		w.cachedStateProviders[elementId] = cached
	}
	return cached
}

// ------------------------------------------------------------------------------------
// Pagination
// ------------------------------------------------------------------------------------

func (w ListWidget) renderPagination(s Session, m MarkupBuilder) {
	np := w.list.TotalPages()

	if np < 2 {
		return
	}

	windowSize := MinInt(np, 5)
	windowStart := Clamp(w.list.CurrentPage()-windowSize/2, 0, np-windowSize)
	windowStop := Clamp(windowStart+windowSize, 0, np-1)

	m.TgOpen(`div class="row"`).TgContent()
	{
		m.TgOpen(`nav aria-label="Page navigation"`).TgContent()
		{
			m.TgOpen(`ul class="pagination d-flex justify-content-center"`).TgContent()
			{
				w.renderPagePiece(s, m, `&lt;&lt;`, 0, true)
				w.renderPagePiece(s, m, `&lt;`, w.list.CurrentPage()-1, true)

				for i := windowStart; i <= windowStop; i++ {
					w.renderPagePiece(s, m, IntToString(i+1), i, false)
				}
				w.renderPagePiece(s, m, `&gt;`, w.list.CurrentPage()+1, true)
				w.renderPagePiece(s, m, `&gt;&gt;`, np-1, true)
			}
			m.TgClose()
		}
		m.TgClose()
	}
	m.TgClose()
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
		m.A(`" onclick="jsButton('`, s.PrependId(w.pagePrefix), targetPage, `')`)
	}
	m.A(`">`, label, `</a></li>`, CR)
}

// Process a possible pagniation control event.
func (w ListWidget) handlePagerClick(sess Session, message string) bool {
	pr := PrIf("", false)
	pr("handlePagerClick, message:", message, "pagePrefix:", w.pagePrefix)
	if page_str, f := TrimIfPrefix(message, "page_"); f {
		pr("page_str:", page_str)
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
			w.Repaint()
			break
		}
		return true
	}
	return false
}
