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
	Listener         ListWidgetListener
}

type ListWidgetListener func(sess Session, widget *ListWidgetStruct, itemId int, args string)

func listListenWrapper(sess Session, widget Widget, value string) (any, error) {
	pr := PrIf(false)
	pr("listListenWrapper, value:", value)

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
		val, err := ParseInt(value[0:c])
		pr("remainder:", remainder, "value:", val, "err:", err)
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

type ListWidget = *ListWidgetStruct

// Construct a ListWidget.
//
// itemWidget : this is a widget that will be rendered for each displayed item
func NewListWidget(m WidgetManager, id string, list ListInterface, itemWidget Widget) ListWidget {
	Todo("!Have option to wrap list items in a clickable div")
	CheckArg(itemWidget != nil, "No itemWidget given")

	w := ListWidgetStruct{
		list:             list,
		itemWidget:       itemWidget,
		WithPageControls: true,
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
		m.A(`" onclick="jsButton('`, s.baseIdPrefix+w.pagePrefix, targetPage, `')`)
	}
	m.A(`">`, label, `</a></li>`, CR)
}

func (w ListWidget) RenderTo(s Session, m MarkupBuilder) {
	pr := PrIf(false)
	pr("ListWidget.RenderTo")
	m.Comment("ListWidget")

	m.TgOpen(`div id=`).A(QUOTED, w.BaseId).TgContent()

	//m.OpenTag(`div id="`, w.BaseId, `"`)

	if w.WithPageControls {
		w.renderPagination(s, m)
	}

	{
		m.TgOpen(`div class="row"`).TgContent()
		{
			elementIds := w.list.GetPageElements()
			pr("rendering page num:", w.list.CurrentPage(), "element ids:", elementIds)

			// While rendering this list's items, we will replace any existing default state provider with
			// the list's one.  Save the current default state provider here, for later restoration.
			savedStateProvider := s.baseStateProvider
			savedBaseIdPrefix := s.baseIdPrefix
			for _, id := range elementIds {
				m.Comment("----------------- rendering list item with id:", id)

				s.baseStateProvider = w.list.ItemStateProvider(s, id)

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
		m.TgClose()
	}

	if w.WithPageControls {
		w.renderPagination(s, m)
	}

	m.TgClose()
}

// Process a possible pagniation control event.
func (w ListWidget) handlePagerClick(sess Session, message string) bool {
	pr := PrIf(false)
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
