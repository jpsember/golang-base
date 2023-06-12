package webserv

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/json"
	"net/http"
	"sync"
)

type Session = *SessionStruct

type SessionStruct struct {
	Id string

	// For storing an application Oper, for example
	AppData any

	// widget representing the entire page; nil if not constructed yet
	PageWidget Widget
	// Lock for making request handling thread safe; we synchronize a particular session's requests
	Mutex sync.RWMutex
	// JSMap containing widget values, other user session state
	State JSMap

	widgetManager WidgetManager
	repaintSet *Set[string]

	// Current request variables
	responseWriter http.ResponseWriter
	request        *http.Request
	// If nonempty, problem detected with current request
	requestProblem string
	widgetIds      []string
	widgetValues   []string
}

func NewSession() Session {
	s := SessionStruct{
		State: NewJSMap(),
	}
	Todo("Restore user session from filesystem/database")
	return &s
}

// Get WidgetManager for this session, creating one if necessary
func (s Session) WidgetManager() WidgetManager {
	if s.widgetManager == nil {
		s.widgetManager = NewWidgetManager()
	}
	return s.widgetManager
}

// Get widget map from the WidgetManager.
func (s Session) widgetMap() WidgetMap {
	return s.WidgetManager().widgetMap
}

func (s Session) ToJson() *JSMapStruct {
	m := NewJSMap()
	m.Put("id", s.Id)
	return m
}

// Mark a widget for repainting.
func (s Session) Repaint(w Widget) {
	pr := PrIf(true)

	id := w.GetId()

	// There are two repaint-related variables for each widget:
	//  a) dirty:      the widget needs repainting
	//  b) dirtyBelow  some widget within the widget's subtree needs repainting
	//
	// If widget's dirty flag is already set, return.  Otherwise:
	//
	// Set every widget in the subtree as dirty
	// Mark all ancestors (parent, parent's parent, etc) as dirtyBelow
	//
	val, exists := s.repaintMap[id]
	pr("repaint widget", id, "value:", val, "exists:", exists)
	if exists {
		return
	}
	pr("marking as DIRTY")
	s.repaintMap[id] = REPAINT_DIRTY
	s.markDescendentsDirtyBelow(w)
	s.markAncestorsDirtyAbove(w)
}

func (s Session) markAncestorsDirtyAbove(w Widget) {
	pr := PrIf(true)
	pr("marking descendents dirty below", w.GetId())
	for _, c := range w.GetChildren() {
		id := c.GetId()
		_, exists := s.repaintMap[id]
		if !exists {
			pr("marking child as dirty below:", id)

func (s Session) markDescendentsDirtyBelow(w Widget) {
	pr := PrIf(true)
	pr("marking descendents dirty below", w.GetId())
	for _, c := range w.GetChildren() {
		id := c.GetId()
		_, exists := s.repaintMap[id]
		if !exists {
			pr("marking child as dirty below:", id)
			s.repaintMap[id] = REPAINT_DIRTY_BELOW
			s.markDescendentsDirtyBelow(w)
		}
	}
}

const REPAINT_DIRTY = 1 << 0
const REPAINT_DIRTY_BELOW = 1 << 1

func ParseSession(source JSEntity) Session {
	var s = source.(*JSMapStruct)
	var n = NewSession()
	n.Id = s.OptString("id", "")
	return n
}

// Prepare for serving a client request from this session's user. Acquire a lock on this session.
func (s Session) HandleAjaxRequest(w http.ResponseWriter, req *http.Request) {
	defer s.discardRequest()
	s.Mutex.Lock()
	s.responseWriter = w
	s.request = req
	s.repaintMap = make(map[string]byte)
	s.requestProblem = ""
	s.parseAjaxRequest(req)
	s.processClientMessage()
	s.sendAjaxResponse()
}

func (s Session) parseAjaxRequest(req *http.Request) {
	// At present, the ajax request parameters are of the form
	//  /ajax?w=<widget id>&v=<widget value>
	//
	v := req.URL.Query()
	// A url can contain multiple values for a parameter, though we
	// will expected just one.
	s.widgetValues, _ = v[clientKeyValue]
	s.widgetIds, _ = v[clientKeyWidget]
}

func (s Session) processClientMessage() {
	// At present, we will assume that the request consists of a single widget id, and perhaps a single value
	// for that widget
	//
	widget := s.GetWidget()
	if !s.Ok() {
		return
	}
	listener := widget.GetBaseWidget().Listener
	if listener == nil {
		s.SetProblem("no listener for id", widget.GetId())
		return
	}
	listener(s, widget)
}

func (s Session) processRepaintFlags(w Widget, refmap JSMap) {
	id := w.GetId()
	pr := PrIf(true)
	pr("processRepaintFlags for widget:", id)
	repaintCode, found := s.repaintMap[id]
	if !found {
		pr("...no repaint necessary")
		return
	}
	pr("...repaint flag:", repaintCode)
	if repaintCode == REPAINT_DIRTY {
		pr("...repainting this one")
		m := NewMarkupBuilder()
		w.RenderTo(m, s.State)
		refmap.Put(w.GetId(), m.String())
	}
	for _, c := range w.GetChildren() {
		pr("...processing child of", id)
		s.processRepaintFlags(c, refmap)
	}
}

// Send Ajax response back to client.
func (s Session) sendAjaxResponse() {
	if !s.Ok() {
		return
	}
	pr := PrIf(true)

	jsmap := NewJSMap()

	// refmap will be the map sent to the client with the widgets
	refmap := NewJSMap()

	s.processRepaintFlags(s.PageWidget, refmap)
	//// Determine which widgets need repainting
	//if s.repaintMap.Size() != 0 {
	//	jsmap.Put("w", refmap)
	//
	//	painted := NewSet[string]()
	//
	//	for k, _ := range s.repaintMap.WrappedMap() {
	//		w := wm[k]
	//		addSubtree(painted, w)
	//	}
	//
	//	// Do a depth first search of the widget map, sending widgets that have been marked for painting
	//	stack := NewArray[string]()
	//	stack.Add(s.PageWidget.GetId())
	//	for stack.NonEmpty() {
	//		widgetId := stack.Pop()
	//		widget := wm[widgetId]
	//		if painted.Contains(widgetId) {
	//			m := NewMarkupBuilder()
	//			widget.RenderTo(m, s.State)
	//			refmap.Put(widgetId, m.String())
	//		}
	//		for _, child := range widget.GetChildren() {
	//			stack.Add(child.GetId())
	//		}
	//	}
	//}

	pr("sending back to Ajax caller:", INDENT, jsmap)
	content := jsmap.CompactString()

	s.responseWriter.Write([]byte(content))
}

func addSubtree(target *Set[string], w Widget) {
	id := w.GetId()
	// If we've already added this to the list, do nothing
	if target.Contains(id) {
		return
	}
	target.Add(id)
	for _, c := range w.GetChildren() {
		addSubtree(target, c)
	}
}

// Discard state added to session to serve a request; release session lock.
func (s Session) discardRequest() {
	problem := s.GetProblem()
	if problem != "" {
		Pr("Problem processing client message:", INDENT, problem)
	}
	s.responseWriter = nil
	s.request = nil
	s.requestProblem = ""
	s.widgetValues = nil
	s.widgetIds = nil
	// Empty the map
	Pr("discarding request, clearing repaintMap")
	//s.repaintMap = nil
	s.Mutex.Unlock()
}

func (s Session) SetProblem(message ...any) Session {
	if s.requestProblem == "" {
		s.requestProblem = "Problem with ajax request: " + ToString(message...)
	}
	return s
}

func (s Session) GetProblem() string {
	return s.requestProblem
}

func (s Session) Ok() bool {
	return s.requestProblem == ""
}

// Read request's (single) widget id
func (s Session) GetWidgetId() string {
	if s.widgetIds == nil || len(s.widgetIds) != 1 {
		s.SetProblem("Unable to get widget id")
		return ""
	}
	return s.widgetIds[0]
}

// Read request's widget value as a string
func (s Session) GetValueString() string {
	if s.widgetValues == nil || len(s.widgetValues) != 1 {
		s.SetProblem("Unable to get widget value")
		return ""
	}
	return s.widgetValues[0]
}

func (s Session) GetWidget() Widget {
	widgetId := s.GetWidgetId()
	if s.Ok() {
		widget, ok := s.widgetMap()[widgetId]
		if ok {
			return widget
		}
		s.SetProblem("no widget found with id", widgetId)
	}
	return nil
}
