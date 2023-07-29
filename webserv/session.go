package webserv

import (
	. "github.com/jpsember/golang-base/base"
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
	repaintSet    *Set[string]

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
	Todo("!Restore user session from filesystem/database")
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
	b := w.GetBaseWidget()
	pr := PrIf(debRepaint)
	id := b.Id
	pr("Repaint:", id)
	if s.repaintSet.Add(id) {
		pr("...adding to set")
	}
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
	s.Mutex.Lock()
	s.responseWriter = w
	s.request = req
	s.repaintSet = NewSet[string]()
	s.requestProblem = ""
	s.parseAjaxRequest(req)
	s.processClientMessage()
	s.sendAjaxResponse()
}

// Serve a request for a resource
func (s Session) HandleResourceRequest(w http.ResponseWriter, req *http.Request, resourcePath Path) {
	defer s.discardRequest()
	s.Mutex.Lock()
	s.responseWriter = w
	s.request = req
	s.requestProblem = ""

	var err error
	for {
		resource := req.URL.Path
		var resPath Path
		resPath, err = resourcePath.Join(resource)
		if err != nil {
			break
		}

		var content []byte
		content, err = resPath.ReadBytes()
		if err != nil {
			break
		}
		_, err = s.responseWriter.Write(content)
		break
	}
	if err != nil {
		s.SetRequestProblem(err)
	}
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
	//Pr("parsed ajax request, values:", s.widgetValues, "ids:", s.widgetIds)
}

func (s Session) processClientMessage() {
	// At present, we will assume that the request consists of a single widget id, and perhaps a single value
	// for that widget
	//
	widget := s.GetWidget()
	b := widget.GetBaseWidget()

	if !s.Ok() {
		return
	}
	listener := b.Listener
	if listener == nil {
		s.SetRequestProblem("no listener for id", b.Id)
		return
	}
	if !widget.GetBaseWidget().Enabled() {
		s.SetRequestProblem("widget is disabled", b.Id)
		return
	}
	listener(s, widget)
}

func (s Session) processRepaintFlags(debugDepth int, w Widget, refmap JSMap, repaint bool) {
	b := w.GetBaseWidget()
	id := b.Id
	pr := PrIf(debRepaint)
	pr(Dots(debugDepth*4)+IntToString(debugDepth), "repaint, flag:", repaint, "id:", id)

	if !repaint {
		if s.repaintSet.Contains(id) {
			repaint = true
			pr(Dots(debugDepth*4), "repaint flag was set; repainting entire subtree")
		}
	}

	if repaint {
		m := NewMarkupBuilder()
		w.RenderTo(m, s.State)
		refmap.Put(id, m.String())
	}

	for _, c := range w.GetChildren() {
		s.processRepaintFlags(1+debugDepth, c, refmap, repaint)
	}
}

const respKeyWidgetsToRefresh = "w"

var debRepaint = false && Alert("debRepaint")

// Send Ajax response back to client.
func (s Session) sendAjaxResponse() {
	if !s.Ok() {
		return
	}
	pr := PrIf(debRepaint)

	jsmap := NewJSMap()

	// refmap will be the map sent to the client with the widgets
	refmap := NewJSMap()

	s.processRepaintFlags(0, s.PageWidget, refmap, false)

	jsmap.Put(respKeyWidgetsToRefresh, refmap)
	pr("sending back to Ajax caller:", INDENT, jsmap)
	content := jsmap.CompactString()

	s.responseWriter.Write([]byte(content))
}

// Discard state added to session to serve a request; release session lock.
func (s Session) discardRequest() {
	problem := s.GetRequestProblem()
	if problem != "" {
		Pr("Problem processing client message:", INDENT, problem)
	}
	s.responseWriter = nil
	s.request = nil
	s.requestProblem = ""
	s.widgetValues = nil
	s.widgetIds = nil
	s.repaintSet = nil
	s.Mutex.Unlock()
}

func (s Session) SetRequestProblem(message ...any) Session {
	if s.requestProblem == "" {
		s.requestProblem = "Problem with ajax request: " + ToString(message...)
		Pr("...set request problem:", s.requestProblem)
	}
	return s
}

func (s Session) GetRequestProblem() string {
	return s.requestProblem
}

func (s Session) Ok() bool {
	return s.requestProblem == ""
}

// Read request's (single) widget id
func (s Session) GetWidgetId() string {
	if s.widgetIds == nil || len(s.widgetIds) != 1 {
		s.SetRequestProblem("Unable to get widget id")
		return ""
	}
	return s.widgetIds[0]
}

// Read request's widget value as a string
func (s Session) GetValueString() string {
	if s.widgetValues == nil || len(s.widgetValues) != 1 {
		s.SetRequestProblem("Unable to get widget value")
		return ""
	}
	return s.widgetValues[0]
}

// Read request's widget value as a boolean
func (s Session) GetValueBoolean() bool {
	if s.widgetValues == nil || len(s.widgetValues) != 1 {
		s.SetRequestProblem("Unable to get widget value")
		return false
	}
	str := s.widgetValues[0]
	switch str {
	case "true":
		return true
	case "false":
		return false
	default:
		s.SetRequestProblem("Unable to parse boolean widget value:", Quoted(str))
		return false
	}
}

func (s Session) GetWidget() Widget {
	widgetId := s.GetWidgetId()
	if s.Ok() {
		widget, ok := s.widgetMap()[widgetId]
		if ok {
			return widget
		}
		s.SetRequestProblem("no widget found with id", widgetId)
	}
	return nil
}

func getProblemId(w Widget) string {
	return w.GetBaseWidget().Id + ".problem"
}

func (s Session) ClearWidgetProblem(widget Widget) {
	key := getProblemId(widget)
	s.State.Delete(key)
}

func (s Session) SetWidgetProblem(widget Widget, s2 string) {
	CheckArg(s2 != "")
	key := getProblemId(widget)
	s.State.Put(key, s2)
}

var cachedCurrentDirectoryString = CurrentDirectory().String()
