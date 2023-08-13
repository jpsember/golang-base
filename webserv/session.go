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
	clientInfo     []string
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
	b := w.Base()
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
func (s Session) HandleResourceRequest(w http.ResponseWriter, req *http.Request, resourcePath Path) error {
	defer s.discardRequest()
	s.Mutex.Lock()
	s.responseWriter = w
	s.request = req
	s.requestProblem = ""

	var err error
	resource := req.URL.Path
	var resPath Path
	resPath, err = resourcePath.Join(resource)
	if err != nil {
		return err
	}

	var content []byte
	content, err = resPath.ReadBytes()
	if err != nil {
		return err
	}

	WriteResponse(s.responseWriter, InferContentTypeM(resource), content)
	return err
}

func (s Session) parseAjaxRequest(req *http.Request) {
	// At present, the ajax request parameters are of the form
	//  /ajax? [expr [& expr]*]
	// where expr is:
	//  w=<widget id>
	//  v=<widget value>
	//  i=<client information as json map, encoded as string>
	v := req.URL.Query()

	// A url can contain multiple values for a parameter, though we
	// will expected just one.
	s.widgetValues, _ = v[clientKeyValue]
	s.widgetIds, _ = v[clientKeyWidget]
	s.clientInfo, _ = v[clientKeyInfo]
}

func (s Session) processClientMessage() {
	// Process client info, if it was sent
	if info, err := getSingleValue(s.clientInfo); err == nil {
		s.processClientInfo(info)
		// If there isn't a widget message as well, do nothing else
		if len(s.widgetIds) == 0 {
			return
		}
	}

	// At present, we will assume that the request consists of a single widget id, and perhaps a single value
	// for that widget
	//
	widget := s.GetWidget()
	b := widget.Base()

	if !s.Ok() {
		return
	}
	listener := b.Listener
	if listener == nil {
		Todo("?Is it ok to have no listener?")
		//s.SetRequestProblem("no listener for id", b.Id)
		return
	}
	if !b.Enabled() {
		s.SetRequestProblem("widget is disabled", b.Id)
		return
	}

	listener(s, widget)
}

func (s Session) processClientInfo(infoString string) {
	json, err := JSMapFromString(infoString)
	if err != nil {
		Pr("failed to parse json:", err, INDENT, infoString)
		return
	}
	Todo("!process client info:", INDENT, json)
}

func (s Session) processRepaintFlags(debugDepth int, w Widget, refmap JSMap, repaint bool) {
	b := w.Base()
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
	WriteResponse(s.responseWriter, "application/json", []byte(content))
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
		Alert("<2 setting request problem:", s.requestProblem)
	}
	return s
}

func (s Session) GetRequestProblem() string {
	return s.requestProblem
}

func (s Session) Ok() bool {
	return s.requestProblem == ""
}

func getSingleValue(array []string) (string, error) {
	if array != nil && len(array) == 1 {
		return array[0], nil
	}
	return "", Error("expected single string, got:", array)
}

// Read request's (single) widget id
func (s Session) GetWidgetId() string {
	id, err := getSingleValue(s.widgetIds)
	if err != nil {
		s.SetRequestProblem("Unable to get widget id")
		return ""
	}
	return id
}

// Read request's widget value as a string
func (s Session) GetValueString() string {
	value, err := getSingleValue(s.widgetValues)
	if err != nil {
		s.SetRequestProblem("Unable to get widget value")
		return ""
	}
	return value
}

// Read request's widget value as a boolean
func (s Session) GetValueBoolean() bool {
	str := s.GetValueString()
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
	return WidgetId(w) + ".problem"
}

func (s Session) ClearWidgetProblem(widget Widget) {
	s.auxSetWidgetProblem(widget, "")
}

func (s Session) SetWidgetProblem(widget Widget, problemText string) {
	CheckArg(problemText != "")
	s.auxSetWidgetProblem(widget, problemText)
}

func (s Session) auxSetWidgetProblem(widget Widget, problemText string) {
	key := getProblemId(widget)
	state := s.State
	existingProblem := state.OptString(key, "")
	if existingProblem != problemText {
		if problemText == "" {
			state.Delete(key)
		} else {
			state.Put(key, problemText)
		}
		s.Repaint(widget)
	}
}

// Include javascript call within page to get client's display properties.
func (s Session) RequestClientInfo(sb MarkupBuilder) {
	// If necessary, determine client's screen resolution by including some javascript that will make an ajax
	// call back to us with that information.
	if true {
		Alert("!Always making resolution call; might want to avoid infinite calls by only requesting if at least n seconds elapsed")
		sb.A(`<script>jsGetDisplayProperties();</script>`).Cr()
	}
}

var cachedCurrentDirectoryString = CurrentDirectory().String()
