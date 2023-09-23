package webserv

import (
	. "github.com/jpsember/golang-base/base"
	"strings"
)

// A Widget that displays editable text
type ListWidgetStruct struct {
	BaseWidgetObj
	list             ListInterface
	itemWidget       Widget
	pagePrefix       string
	WithPageControls bool
	listener         ListWidgetListener
}

type ListWidgetListener func(sess Session, widget *ListWidgetStruct, itemId int, args string)

func listListenWrapper(sess Session, widget Widget, value string) (any, error) {
	pr := PrIf("", true)
	pr("listListenWrapper, value:", Quoted(value), "caller:", Caller())

	b := widget.(ListWidget)

	// See if this is an event from the page controls
	if b.handlePagerClick(sess, value) {
		pr("...page controls handled it")
		return nil, nil
	}

	// This is presumably something like <element id> '.' <remainder>
	itemId := -1
	c := strings.IndexByte(value, '.')
	remainder := ""
	if c > 0 {
		remainder = value[c+1:]
		valStr := value[0:c]
		val, err := ParseInt(valStr)
		if err == nil {
			Todo("!Verify that the parsed value matches an id in the list")
		}
		pr("remainder:", remainder, "value:", val, "err:", err)
		if err != nil {
			Alert("#50 trouble parsing int from:", value)
		} else {
			itemId = int(val)
		}

		// Look for a widget (presumably within the original ListItem widget) with the extracted id
		sourceId := remainder
		sourceWidget := sess.WidgetManager().Opt(sourceId)
		if sourceWidget == nil {
			Alert("#50Can't find source widget with id", Quoted(sourceId), "; original value:", Quoted(value))
		} else {
			// Forward the message to that widget
			Todo("How do we distinguish between value actions (like text fields) and button presses?")
			newVal, err := sourceWidget.LowListener()(sess, sourceWidget, valStr)
			Pr("sourceWidget lowlistener returned:", newVal, err)
			return newVal, err
		}
	}

	if b.listener == nil {
		Alert("#50No ListListener registered; itemId:", itemId, "args:", remainder)
	} else {
		b.listener(sess, b, itemId, remainder)
	}
	return "", nil
}

type ListWidget = *ListWidgetStruct

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
		listener:         listener,
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
	pr := PrIf("ListWidget.RenderTo", true)
	pr("ListWidget.RenderTo")
	m.Comment("ListWidget")

	m.TgOpen(`div id=`).A(QUOTED, w.Id()).TgContent()

	if w.WithPageControls {
		w.renderPagination(s, m)
	}

	{
		m.TgOpen(`div class="row"`).TgContent()
		{
			elementIds := w.list.GetPageElements()
			pr("rendering page #:", w.list.CurrentPage(), "element ids:", elementIds)

			// While rendering this list's items, we will replace any existing default state provider with
			// the list's one.  Save the current default state provider here, for later restoration.

			savedStateProvider := s.BaseStateProvider()

			Todo("What is savedBaseIdPrefix used for?  Can't this just be the BaseStateProvider?")
			//savedBaseIdPrefix := s.baseIdPrefix

			pr(VERT_SP, "saved pv:", INDENT, savedStateProvider)

			Todo("I think we want to send list *item* messages to the list item widget's handler, with message to include the list item id")

			for _, id := range elementIds {
				m.Comment("----------------- rendering list item with id:", id)

				pv := w.list.ItemStateProvider(s, id)
				pr("itm pv:", INDENT, pv)
				Todo("what is the prefix returned by the list item state provider?")
				np := NewStateProvider(w.Id()+"."+IntToString(id)+"."+savedStateProvider.Prefix, pv.State)
				pr("replacing with:", INDENT, np)

				// This doesn't quite work... clicking on the image is not having any effect.
				// The heading and summary strings appear ok though.
				s.SetBaseStateProvider(np)

				// When rendering list items, any ids should be mangled in such a way that
				//  a) ids remain distinct, even if we are rendering the same widget for each row; and
				//  b) when responding to click events and the like, we can figure out which list, and
				//      element within the list, generated the event.

				// The item widgets will have this id structure:
				//
				// [id of containing ListWidget].[id of item].[session.baseIdPrefix (?what for?)]
				//
				//s.baseIdPrefix = w.Id() + "." + IntToString(id) + "." + savedBaseIdPrefix
				w.itemWidget.RenderTo(s, m)
			}
			// Restore the default state provider to what it was before we rendered the items.
			s.SetBaseStateProvider(savedStateProvider)
		}
		m.TgClose()
	}

	if w.WithPageControls {
		w.renderPagination(s, m)
	}

	m.TgClose()
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
			sess.Repaint(w)
			break
		}
		return true
	}
	return false
}
