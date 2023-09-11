package webserv

import (
	"bytes"
	. "github.com/jpsember/golang-base/base"
	"github.com/jpsember/golang-base/webapp/gen/webapp_data"
	"io"
	"net/http"
	"strings"
	"sync"
)

var loggedInUsersSet = NewSet[int]()
var loggedInUsersSetLock sync.RWMutex

func IsUserLoggedIn(userId int) bool {
	loggedInUsersSetLock.Lock()
	defer loggedInUsersSetLock.Unlock()
	return loggedInUsersSet.Contains(userId)
}

func TryRegisteringUserAsLoggedIn(sess Session, user webapp_data.User, loggedInState bool) bool {
	loggedInUsersSetLock.Lock()
	defer loggedInUsersSetLock.Unlock()

	userId := user.Id()
	currentState := loggedInUsersSet.Contains(userId)
	changed := currentState != loggedInState
	if changed {
		if loggedInState {
			loggedInUsersSet.Add(userId)
			sess.AppData = user
		} else {
			loggedInUsersSet.Remove(userId)
			sess.AppData = nil
		}
	}
	return changed
}

func DiscardAllSessions(sessionManager SessionManager) {
	loggedInUsersSetLock.Lock()
	defer loggedInUsersSetLock.Unlock()

	Alert("Discarding all sessions")
	sessionManager.DiscardAllSessions()
	loggedInUsersSet.Clear()
}

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

	// Current request variables
	responseWriter http.ResponseWriter
	request        *http.Request
	// If nonempty, problem detected with current request
	requestProblem string
	clientInfo     []string
	ajaxWidget     Widget
	widgetValue    string
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
		s.widgetManager = NewWidgetManager(s)
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
	s.requestProblem = ""
	s.parseAjaxRequest(req)
	s.WidgetManager().clearRepaintSet()
	if false && Alert("dumping") {
		Pr("Query:", INDENT, req.URL.Query())
	}
	s.processClientMessage()
	s.sendAjaxResponse()
}

func (s Session) HandleUploadRequest(w http.ResponseWriter, req *http.Request, widgetId string) {

	defer s.discardRequest()
	Todo("!lots of duplicated code here with HandleAjaxRequest")
	s.Mutex.Lock()
	s.responseWriter = w
	s.request = req
	s.requestProblem = ""

	s.WidgetManager().clearRepaintSet()

	s.processUpload(w, req, widgetId)

	// Send the usual ajax response

	s.sendAjaxResponse()
}

func (s Session) processUpload(w http.ResponseWriter, req *http.Request, widgetId string) {

	var fileUploadWidget FileUpload

	widget := s.WidgetManager().Opt(widgetId)
	if widget == nil {
		Alert("Can't find upload widget:", widgetId)
		return
	}
	var ok bool
	if fileUploadWidget, ok = widget.(FileUpload); !ok {
		Alert("Not an UploadWidget:", widgetId)
		return
	}

	problem := ""
	var result []byte

	for {

		problem = "upload request was not POST"
		if req.Method != "POST" {
			break
		}

		// From https://freshman.tech/file-upload-golang/

		problem = "The uploaded file is too big. Please choose an file that's less than 10MB in size"
		{
			const MAX_UPLOAD_SIZE = 10_000_000
			req.Body = http.MaxBytesReader(w, req.Body, MAX_UPLOAD_SIZE)
			if err := req.ParseMultipartForm(MAX_UPLOAD_SIZE); err != nil {
				Todo("this should be returned to the user as a widget error msg")
				break
			}
		}

		// The argument to FormFile must match the name attribute
		// of the file input on the frontend; not sure what that is about

		problem = "trouble getting request FormFile"
		file, _, err1 := req.FormFile(widgetId + ".input")
		if err1 != nil {
			break
		}

		problem = "failed to read uploaded file into byte array"
		var buf bytes.Buffer
		length, err1 := io.Copy(io.Writer(&buf), file)
		file.Close()
		if err1 != nil {
			break
		}

		CheckArg(len(buf.Bytes()) == int(length))
		result = buf.Bytes()

		Todo("!Must ensure thread safety while working with the user session")
		problem = ""
		break
	}
	if problem != "" {
		s.SetWidgetProblem(fileUploadWidget, problem)
		Alert("Problem with upload:", problem)
		return
	}
	fileUploadWidget.listener(s, fileUploadWidget, result)
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

	t1 := v[clientKeyWidget]
	t2 := v[clientKeyValue]
	var wid string
	// A value is optional, as buttons don't send them.
	if len(t1) == 1 && len(t2) <= 1 {
		wid = t1[0]
		if len(t2) == 1 {
			s.widgetValue = t2[0]
		}
	}
	s.ajaxWidget = s.WidgetManager().Opt(wid)
	s.clientInfo, _ = v[clientKeyInfo]
}

func (s Session) processClientMessage() {
	// Process client info, if it was sent
	if info, err := getSingleValue(s.clientInfo); err == nil {
		s.processClientInfo(info)
		// If there isn't a widget message as well, do nothing else
		if s.ajaxWidget == nil {
			return
		}
	}

	// At present, we will assume that the request consists of a single widget id, and perhaps a single value
	// for that widget
	//
	widget := s.ajaxWidget
	if widget == nil {
		s.SetRequestProblem("no widget found", widget)
		return
	}

	if !widget.Enabled() {
		s.SetRequestProblem("widget is disabled", widget)
		return
	}

	if widget.LowListener() == nil {
		Alert("#50Widget has no low-level listener:", Info(widget))
		return
	}
	updatedValue, err := widget.LowListener()(s, widget, s.widgetValue)

	s.State.Put(widget.Id(), updatedValue)
	if err != nil {
		Pr("got error from widget listener:", widget.Id(), INDENT, err)
	}
	// Always update the problem, in case we are clearing a previous error
	s.SetWidgetProblem(widget, err)
}

func (s Session) processClientInfo(infoString string) {
	json, err := JSMapFromString(infoString)
	if err != nil {
		Pr("failed to parse json:", err, INDENT, infoString)
		return
	}
	Todo("!process client info:", INDENT, json)
}

func (s Session) processRepaintFlags(repaintSet StringSet, debugDepth int, w Widget, refmap JSMap, repaint bool) {
	id := w.Id()
	pr := PrIf(debRepaint)
	pr(Dots(debugDepth*4)+IntToString(debugDepth), "repaint, flag:", repaint, "id:", id)

	if !repaint {
		if repaintSet.Contains(id) {
			repaint = true
			pr(Dots(debugDepth*4), "repaint flag was set; repainting entire subtree")
		}
	}

	if repaint {
		m := NewMarkupBuilder()
		w.RenderTo(m, s.State)
		refmap.Put(id, m.String())
	}

	for _, c := range w.Children().Array() {
		s.processRepaintFlags(repaintSet, 1+debugDepth, c, refmap, repaint)
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

	s.processRepaintFlags(s.WidgetManager().repaintSet, 0, s.PageWidget, refmap, false)

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
	s.widgetValue = ""
	s.ajaxWidget = nil
	s.WidgetManager().clearRepaintSet()
	s.Mutex.Unlock()
}

func (s Session) SetRequestProblem(message ...any) Session {
	if s.requestProblem == "" {
		s.requestProblem = "Problem with ajax request: " + ToString(message...)
		Alert("#50<2 setting request problem:", s.requestProblem)
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

func (s Session) SetWidgetIdProblem(widgetId string, problem any) {
	widget := s.WidgetManager().Get(widgetId)
	s.SetWidgetProblem(widget, problem)
}

func (s Session) SetWidgetProblem(widget Widget, problem any) {
	var text string
	if problem != nil {
		switch t := problem.(type) {
		case string:
			text = t
		case error:
			text = t.Error()
		default:
			BadArg("<1Unsupported type")
		}
	}
	s.auxSetWidgetProblem(widget, text)
}

func (s Session) auxSetWidgetProblem(widget Widget, problemText string) {
	key := WidgetIdWithProblem(widget.Id())
	state := s.State
	existingProblem := state.OptString(key, "")
	if existingProblem != problemText {
		Pr("SetWidgetProblem:", widget.Id(), "from:", existingProblem, "to:", problemText)
		if problemText == "" {
			state.Delete(key)
		} else {
			state.Put(key, problemText)
		}
		s.WidgetManager().Repaint(widget)
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

func (s Session) DeleteStateErrors() {
	m := s.State.MutableWrapped()
	Todo("safe to delete key while iterating through them?")
	for k, _ := range m {
		if strings.HasSuffix(k, ".error") {
			delete(m, k)
		}
	}
}

func (s Session) DeleteStateError(id string) {
	m := s.State.MutableWrapped()
	delete(m, id)
}

func (s Session) DeleteStateFieldsWithPrefix(prefix string) {
	m := s.State.MutableWrapped()
	for k, _ := range m {
		if strings.HasPrefix(k, ".error") {
			delete(m, k)
		}
	}
}
