package webserv

import (
	. "github.com/jpsember/golang-base/base"
	"math"
)

type ListWidgetStruct struct {
	BaseWidgetObj
	list                 ListInterface
	itemWidget           Widget
	itemPrefix           string // We want each item widget and its subwidgets to have a unique id
	pagePrefix           string
	WithPageControls     bool
	cachedStateProviders map[int]JSMap
	currentElement       int
}
type ListWidget = *ListWidgetStruct

// Construct a ListWidget.
//
// itemWidget : this is a widget that will be rendered for each displayed item
func NewListWidget(id string, list ListInterface, itemWidget Widget) ListWidget {
	CheckArg(itemWidget != nil, "No itemWidget given")
	w := ListWidgetStruct{
		list:             list,
		itemWidget:       itemWidget,
		WithPageControls: true,
		currentElement:   -1,
	}
	w.InitBase(id)
	w.itemPrefix = id + ":"
	w.pagePrefix = w.itemPrefix + "page:"
	w.SetLowListener(w.lowLevelListener)
	return &w
}

func (w ListWidget) CurrentElement() int {
	x := w.currentElement
	if x < 0 {
		BadState("ListWidget has no current element")
	}
	return x
}

func (w ListWidget) RenderTo(s Session, m MarkupBuilder) {
	debug := false
	pr := PrIf("ListWidget.RenderTo", debug)
	pr("ListWidget.RenderTo; id", QUO, w.Id())

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
		CheckState(len(s.stack) == 2, "Expected two items on state stack: default item, plus one for list renderer")

		for _, id := range elementIds {

			// We want each rendered subwidget to have a unique id, and also a way to tie the widget to an element
			s.PushIdPrefix(w.itemPrefix + IntToString(id) + ":")

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

func (w ListWidget) lowLevelListener(sess Session, widget Widget, value string, args WidgetArgs) (any, error) {
	pr := PrIf("list_widget.lowLevelListener", false)
	pr(VERT_SP, "value:", QUO, value, "args:", args, VERT_SP)

	pr("stack size:", len(sess.stack))
	b := widget.(ListWidget)

	// See if this is an event from the page controls
	if b.handlePagerClick(sess, args) {
		pr("...page controls handled it")
		return nil, nil
	}

	valid, elementId := args.ReadIntWithinRange(0, math.MaxInt32)
	if !valid {
		return nil, Error("Failed to read element id from", args)
	}
	pr("list element id:", elementId)

	// If there are additional arguments, see if there is a widget id prefix within them
	auxWidget := args.FindWidgetIdAsPrefix(sess)

	if auxWidget != nil {
		auxListener := auxWidget.LowListener()
		if auxListener == nil {
			Alert("#50No low-level listener for widget:", QUO, auxWidget.Id(), "within list:", QUO, w.Id())
			return nil, nil
		}
		w.currentElement = elementId
		auxListener(sess, auxWidget, "", args)
		w.currentElement = -1
		return nil, nil
	} else {
		pr("!!! no auxilliary widget to get listener")
		return nil, nil
	}
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
