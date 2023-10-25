package webserv

import (
	. "github.com/jpsember/golang-base/base"
)

// A Widget that displays editable text
type ListWidgetStruct struct {
	BaseWidgetObj
	list                 ListInterface
	itemWidget           Widget
	pagePrefix           string
	WithPageControls     bool
	cachedStateProviders map[int]WidgetStateProvider
}

func (w ListWidget) listListenWrapper(sess Session, widget Widget, value string) (any, error) {
	pr := PrIf("list_widget.LowLevel listener", true)
	pr(VERT_SP, "value:", QUO, value, "caller:", Caller())
	pr("stack size:", len(sess.stack))
	b := widget.(ListWidget)

	// See if this is an event from the page controls
	if b.handlePagerClick(sess, value) {
		pr("...page controls handled it")
		return nil, nil
	}

	// We expect a value to be <element id> ['.' <remainder>]*

	elementIdStr, remainder := ExtractFirstDotArg(value)
	if elementIdStr != "" {
		elementId, err := ParseInt(elementIdStr)
		pr("remainder:", remainder, "value:", elementId, "err:", err)
		if err != nil {
			Alert("#50 trouble parsing int from:", value)
			return nil, Error("trouble parsing int from:", QUO, value)
		}
		Todo("!Verify that the parsed value matches an id in the list")

		// Look for a widget (presumably within the original ListItem widget) with the extracted id.
		// If the value is "xxx.yyy.zzz" and we don't find such a widget, look for "xxx.yyy" and pass "zzz" as the value

		var sourceWidget Widget
		var sourceId string
		sourceId, remainder = ExtractFirstDotArg(remainder)
		if sourceId != "" {
			sourceWidget = sess.Opt(sourceId)
		}

		if sourceWidget == nil {
			Alert("#50Can't find source widget(s) for:", Quoted(sourceId), "; original value:", Quoted(value))
			return nil, Error("can't find widget with id:", QUO, sourceId, "value was:", QUO, value)
		}
		// Forward the message to that widget
		Todo("!How do we distinguish between value actions (like text fields) and button presses?")
		// Set up the same state provider that we did when rendering the widget
		sess.PushStateProvider(w.constructStateProvider(sess, elementId))
    Todo("This is causing a crash, as we are sometimes *changing pages* during this call, which re-initializes the stack and whatnot")
		sess.ProcessWidgetValue(sourceWidget, remainder, elementId)
		sess.PopStateProvider()
		// Fall through to return nil, nil
	}
	return nil, nil
}

type ListWidget = *ListWidgetStruct

// Construct a ListWidget.
//
// itemWidget : this is a widget that will be rendered for each displayed item
func NewListWidget(id string, list ListInterface, itemWidget Widget) ListWidget {
	Todo("!Have option to wrap list items in a clickable div")
	CheckArg(itemWidget != nil, "No itemWidget given")
	w := ListWidgetStruct{
		list:             list,
		itemWidget:       itemWidget,
		WithPageControls: true,
	}
	w.InitBase(id)
	w.LowListen = w.listListenWrapper
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

func (w ListWidget) RenderTo(s Session, m MarkupBuilder) {

	pr := PrIf("ListWidget.RenderTo", false)
	pr("ListWidget.RenderTo; id", w.Id())

	m.TgOpen(`div id=`).A(QUO, w.Id()).TgContent()

	m.Comment("ListWidget")

	// Discard any previously cached state providers, and
	// cache those we are about to construct (so we don't ask client
	// to construct them unnecessarily).
	w.cachedStateProviders = make(map[int]WidgetStateProvider)
	if w.WithPageControls {
		w.renderPagination(s, m)
	}

	{
		m.TgOpen(`div class="row"`).TgContent()
		{
			elementIds := w.list.GetPageElements()
			pr("rendering page #:", w.list.CurrentPage(), "element ids:", elementIds)

			// While rendering this list's items, we will set the state provider to one for each item.
			pr("item prefix:", w.list.ItemPrefix())

			CheckState(len(s.stack) == 2, "Expected two items on state stack: default item, plus one for list renderer")

			for _, id := range elementIds {
				sp := w.constructStateProvider(s, id)
				pr("pushing state provider:", sp)
				s.PushStateProvider(sp)

				elementIdStr := IntToString(id)
				// We want each rendered widget to have a unique id, so include "<element id>:" as a *rendering* prefix
				s.PushIdPrefix(elementIdStr + ":")
				// If we push the state provider AFTER the id prefix, it doesn't work! Why?
				// Note that we are not calling RenderWidget(), which would not draw anything since the
				// list item widget has been marked as detached

				// Periods are used to separate widget id from context
				s.PushClickPrefix(w.Id() + "." + elementIdStr + ".")

				x := m.Len()
				w.itemWidget.RenderTo(s, m)
				pr("rendered item, markup:", INDENT, m.String()[x:])

				s.PopClickPrefix()
				s.PopIdPrefix()
				s.PopStateProvider()
			}
		}
		m.TgClose()
	}

	if w.WithPageControls {
		w.renderPagination(s, m)
	}

	m.TgClose()
}

func (w ListWidget) constructStateProvider(s Session, elementId int) WidgetStateProvider {
	cached := w.cachedStateProviders[elementId]
	if cached == nil {
		pv := w.list.ItemStateProvider(s, elementId)
		cached = NewStateProvider(w.list.ItemPrefix(), pv.State)
		w.cachedStateProviders[elementId] = cached
	}
	return cached
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
