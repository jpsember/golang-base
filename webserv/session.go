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

	//WidgetValue    ClientValue
	//WidgetId       ClientValue
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

// Mark a widget for repainting
func (s Session) Repaint(id string) {
	s.repaintMap.Add(id)
}

func ParseSession(source JSEntity) Session {
	var s = source.(*JSMapStruct)
	var n = NewSession()
	n.Id = s.OptString("id", "")
	return n
}

// Prepare for serving a client request from this session's user
func (s Session) OpenRequest(w http.ResponseWriter, req *http.Request) {
	s.responseWriter = w
	s.request = req
	s.requestProblem = ""
	v := req.URL.Query()
	s.widgetValues, _ = v[clientKeyValue]
	s.widgetIds, _ = v[clientKeyWidget]
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
