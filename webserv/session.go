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
	// widget representing the entire page; nil if not constructed yet
	PageWidget Widget
	// Lock for making request handling thread safe; we synchronize a particular session's requests
	Mutex sync.RWMutex
	// JSMap containing widget values, other user session state
	State JSMap
	// Map of widgets for this session
	WidgetMap  map[string]Widget
	repaintMap *Set[string]
	// TODO: we might want the repaintMap to be ephemeral, only alive while serving the request
	// We also might want to have a singleton, global widget map, since the state is stored here in the session

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
		State:      NewJSMap(),
		repaintMap: NewSet[string](),
	}
	Todo("Restore user session from filesystem/database")
	return &s
}

func (s Session) ToJson() *JSMapStruct {
	m := NewJSMap()
	m.Put("id", s.Id)
	return m
}

// Mark a widget for repainting.
func (s Session) Repaint(id string) {
	s.repaintMap.Add(id)
}

func ParseSession(source JSEntity) Session {
	var s = source.(*JSMapStruct)
	var n = NewSession()
	n.Id = s.OptString("id", "")
	return n
}

// Prepare for serving a client request from this session's user. Acquire a lock on this session.
func (s Session) HandleAjaxRequest(w http.ResponseWriter, req *http.Request) {
	defer s.discardRequest()
	s.prepareRequestVars(w, req)
	s.processClientMessage()
	s.sendAjaxResponse()
}

func (s Session) prepareRequestVars(w http.ResponseWriter, req *http.Request) {
	s.Mutex.Lock()
	s.responseWriter = w
	s.request = req
	s.requestProblem = ""
	v := req.URL.Query()
	s.widgetValues, _ = v[clientKeyValue]
	s.widgetIds, _ = v[clientKeyWidget]
}

func (s Session) processClientMessage() {
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

// Send Ajax response back to client.
func (s Session) sendAjaxResponse() {
	if !s.Ok() {
		return
	}
	pr := PrIf(false)

	jsmap := NewJSMap()

	// TODO: there might be a more efficient way to do the repainting

	// Determine which widgets need repainting
	if s.repaintMap.Size() != 0 {
		// refmap will be the map sent to the client with the widgets
		refmap := NewJSMap()
		jsmap.Put("w", refmap)

		painted := NewSet[string]()

		for k, _ := range s.repaintMap.WrappedMap() {
			w := s.WidgetMap[k]
			addSubtree(painted, w)
		}

		// Do a depth first search of the widget map, sending widgets that have been marked for painting
		stack := NewArray[string]()
		stack.Add(s.PageWidget.GetId())
		for stack.NonEmpty() {
			widgetId := stack.Pop()
			widget := s.WidgetMap[widgetId]
			if painted.Contains(widgetId) {
				m := NewMarkupBuilder()
				widget.RenderTo(m, s.State)
				refmap.Put(widgetId, m.String())
			}
			for _, child := range widget.GetChildren() {
				stack.Add(child.GetId())
			}
		}
	}

	content := jsmap.CompactString()

	pr("sending back to Ajax caller:", INDENT, content)
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
		widget, ok := s.WidgetMap[widgetId]
		if ok {
			return widget
		}
		s.SetProblem("no widget found with id", widgetId)
	}
	return nil
}
