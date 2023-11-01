package webserv

import (
	. "github.com/jpsember/golang-base/base"
	"math"
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

type ListWidgetListener func(sess Session, widget *ListWidgetStruct, elementId int, args WidgetArgs) error

// Construct a ListWidget.
//
// itemWidget : this is a widget that will be rendered for each displayed item
func NewListWidget(id string, list ListInterface, itemWidget Widget, listener ListWidgetListener) ListWidget {
	Todo("Have option to wrap list items in a clickable div")
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
	w.pagePrefix = w.itemPrefix + "page:"

	// If there's an item listener, add a mock listener to the item widget, so that when the item is rendered,
	// it will actually end up calling the list's listener instead
	if listener != nil {
		Todo("Refactor this somehow, maybe just a boolean flag?")
		itemWidget.SetLowListener(mockLowListener)
	}

	return &w
}

var mockLowListener = func(sess Session, widget Widget, value string, args WidgetArgs) (any, error) {
	Die("shouldn't actually get called")
	return nil, nil
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

			// We want each rendered widget to have a unique id, and also a way to tie the widget to a particular list
			// item, so construct a suitable prefix

			nestedWidgetsIdPrefix := w.itemPrefix + IntToString(id) + ":"
			s.PushIdPrefix(nestedWidgetsIdPrefix)

			sp := w.constructStateProvider(s, id)

			pr(VERT_SP, "pushing state provider:", sp)
			s.PushStateMap(sp)

			// If we push the state provider AFTER the id prefix, it doesn't work! Why?
			// Note that we are not calling RenderWidget(), which would not draw anything since the
			// list item widget has been marked as detached

			if debug {
				pr("stacked state:", INDENT, s.StateStackToJson())
			}
			w.itemWidget.RenderTo(s, m)

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

func (w ListWidget) listListenWrapper(sess Session, widget Widget, value string, args WidgetArgs) (any, error) {
	pr := PrIf("list_widget.LowLevel listener", true)
	pr(VERT_SP, "value:", QUO, value, "args:", args, VERT_SP)

	pr("stack size:", len(sess.stack))
	b := widget.(ListWidget)

	// See if this is an event from the page controls
	if b.handlePagerClick(sess, args) {
		pr("...page controls handled it")
		return nil, nil
	}

	valid, elementId := args.ReadIntWithinRange(0, math.MaxInt32)
	var err error
	if !valid {
		err = Error("Failed to read element id from", args)
	} else {
		if w.listItemListener == nil {
			BadState("No list item listener for widget", QUO, w.Id())
		} else {
			err = w.listItemListener(sess, w, elementId, args)
		}
	}
	return nil, err
}

func (w ListWidget) constructStateProvider(s Session, elementId int) JSMap {
	pr := PrIf("list_widget.constructStateProvider", false)
	cached := w.cachedStateProviders[elementId]
	if cached == nil {
		pv := w.list.ItemStateMap(s, elementId)
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
		m.A(`" onclick="jsButton('`, w.pagePrefix, targetPage, `')`)
	}
	m.A(`">`, label, `</a></li>`, CR)
}

func ReadArgIf(args []string, strValue string) (bool, []string) {
	if len(args) != 0 && args[0] == strValue {
		return true, args[1:]
	}
	return false, args
}

func ReadIntArgIf(args []string, minValue int, maxValue int) (bool, []string, int) {
	if len(args) != 0 {
		value, err := ParseInt(args[0])
		if err == nil && value >= minValue && value < maxValue {
			return true, args[1:], value
		}
	}
	return false, args, -1
}

// Process a possible pagniation control event.
func (w ListWidget) handlePagerClick(sess Session, args WidgetArgs) bool {
	pr := PrIf("handlePagerClick", false)
	pr("handlePagerClick, args:", args)
	var result bool
	var pageNumber int
	result = args.ReadIf("page")
	if result {
		result, pageNumber = args.ReadIntWithinRange(0, w.list.TotalPages())
		if result && pageNumber != w.list.CurrentPage() {
			w.list.SetCurrentPage(pageNumber)
			w.Repaint()
		}
	}
	return result
}
