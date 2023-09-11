package webserv

import (
	"bytes"
	. "github.com/jpsember/golang-base/base"
	"github.com/jpsember/golang-base/webserv/gen/webserv_data"
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

func TryRegisteringUserAsLoggedIn(userId int, loggedInState bool) bool {
	// Don't have webserv refer to webapp_data
	loggedInUsersSetLock.Lock()
	defer loggedInUsersSetLock.Unlock()
	currentState := loggedInUsersSet.Contains(userId)
	changed := currentState != loggedInState
	if changed {
		if loggedInState {
			loggedInUsersSet.Add(userId)
		} else {
			loggedInUsersSet.Remove(userId)
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
	AppData map[string]any

	// widget representing the entire page; nil if not constructed yet
	PageWidget Widget
	// Lock for making request handling thread safe; we synchronize a particular session's requests
	Mutex sync.RWMutex
	// JSMap containing widget values, other user session state
	State JSMap

	BrowserInfo   webserv_data.ClientInfo
	widgetManager WidgetManager

	// Current request variables
	responseWriter   http.ResponseWriter
	request          *http.Request
	requestProblem   string // If nonempty, problem detected with current request
	clientInfoString string // If nonempty information sent from client about screen size, etc
	ajaxWidget       Widget // If not nil, the widget sending the ajax
	ajaxWidgetValue  string // The string representation of the ajax widget's requested value (if there was one)
}

var ourDefaultBrowserInfo = webserv_data.NewClientInfo().SetDevicePixelRatio(1.25).SetScreenSizeX(2560).SetScreenSizeY(1440).Build()

func NewSession() Session {
	s := SessionStruct{
		State:       NewJSMap(),
		BrowserInfo: ourDefaultBrowserInfo,
		AppData:     make(map[string]any),
	}
	Todo("!Restore user session from filesystem/database")
	Todo("?ClientInfo (browser info) not sent soon enough")
	Pr("Created a new session, UserData:", s.AppData)
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
	s.parseAjaxRequest(req)
	if false && Alert("dumping") {
		Pr("Query:", INDENT, req.URL.Query())
	}
	s.processClientMessage()
	s.sendAjaxResponse()
}

func (s Session) HandleUploadRequest(w http.ResponseWriter, req *http.Request, widgetId string) {
	s.Mutex.Lock()
	defer s.discardRequest()
	s.responseWriter = w
	s.request = req
	s.processUpload(w, req, widgetId)
	// Send the usual ajax response
	s.sendAjaxResponse()
}

func (s Session) processUpload(w http.ResponseWriter, req *http.Request, uploadWidgetId string) {

	var fileUploadWidget FileUpload

	widget := s.WidgetManager().Opt(uploadWidgetId)
	if widget == nil {
		Alert("Can't find upload widget:", uploadWidgetId)
		return
	}
	var ok bool
	if fileUploadWidget, ok = widget.(FileUpload); !ok {
		Alert("Not an UploadWidget:", uploadWidgetId)
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
		file, _, err1 := req.FormFile(uploadWidgetId + ".input")
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

	// Always update the problem, in case we are clearing a previous error
	s.SetWidgetProblem(uploadWidgetId, problem)
	if problem == "" {
		err := fileUploadWidget.listener(s, fileUploadWidget, result)
		problem = StringFromOptError(err)
	}
	s.SetWidgetProblem(uploadWidgetId, problem)
}

// Serve a request for a resource
func (s Session) HandleResourceRequest(w http.ResponseWriter, req *http.Request, resourcePath Path) error {
	s.Mutex.Lock()
	defer s.discardRequest()
	s.responseWriter = w

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
			s.ajaxWidgetValue = t2[0]
		}
	}
	s.ajaxWidget = s.WidgetManager().Opt(wid)
	clientInfoArray := v[clientKeyInfo]
	if clientInfoArray != nil && len(clientInfoArray) == 1 {
		s.clientInfoString = clientInfoArray[0]
	}
}

func (s Session) processClientMessage() {
	// Process client info, if it was sent
	if s.clientInfoString != "" {
		s.processClientInfo(s.clientInfoString)
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
	updatedValue, err := widget.LowListener()(s, widget, s.ajaxWidgetValue)

	s.State.Put(widget.Id(), updatedValue)
	if err != nil {
		Pr("got error from widget listener:", widget.Id(), INDENT, err)
	}
	// Always update the problem, in case we are clearing a previous error
	s.SetWidgetProblem(widget.Id(), err)
}

func (s Session) processClientInfo(infoString string) {
	json, err := JSMapFromString(infoString)
	if err != nil {
		Pr("failed to parse json:", err, INDENT, infoString)
		return
	}
	s.BrowserInfo = webserv_data.NewClientInfo(). //
							SetDevicePixelRatio(json.OptFloat32("dp", 1.0)). //
							SetScreenSizeX(json.OptInt("sw", 2000)).         //
							SetScreenSizeY(json.OptInt("sh", 0)).Build()
	Todo("?Datagen generated parse() methods don't report errors cleanly; we will need a wrapper?")
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
		RenderWidget(w, s, m)
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
	defer s.Mutex.Unlock()
	problem := s.requestProblem
	if problem != "" {
		Pr("Problem processing client message:", INDENT, problem)
	}
	s.responseWriter = nil
	s.request = nil
	s.requestProblem = ""
	s.ajaxWidget = nil
	s.ajaxWidgetValue = ""
	s.clientInfoString = ""

	s.WidgetManager().clearRepaintSet()
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

func (s Session) SetWidgetProblem(widgetId string, problem any) {
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
	s.auxSetWidgetProblem(widgetId, text)
}

func (s Session) auxSetWidgetProblem(widgetId string, problemText string) {
	key := WidgetIdWithProblem(widgetId)
	state := s.State
	existingProblem := state.OptString(key, "")
	if existingProblem != problemText {
		// Pr("SetWidgetProblem:", widget.Id(), "from:", existingProblem, "to:", problemText)
		if problemText == "" {
			state.Delete(key)
		} else {
			state.Put(key, problemText)
		}
		s.WidgetManager().RepaintIds(widgetId)
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

// ------------------------------------------------------------------------------------
// Accessing values of widgets other than the widget currently being listened to
// ------------------------------------------------------------------------------------

// Read widget value; assumed to be an int.
func (s Session) WidgetIntValue(id string) int {
	return s.State.OptInt(id, 0)
}

// Read widget value; assumed to be a boolean.
func (s Session) WidgetBooleanValue(id string) bool {
	return s.State.OptBool(id, false)
}

// Read widget value; assumed to be a string.
func (s Session) WidgetStrValue(id string) string {
	return s.State.OptString(id, "")
}

func (s Session) PutSessionData(key string, value any) {
	Pr("Storing user data", key, "=>", TypeOf(value))
	s.AppData[key] = value
}

func (s Session) OptSessionData(key string) any {
	value := s.AppData[key]
	//Pr("Getting user data", key, "=>", TypeOf(value))
	return value
}

func (s Session) GetSessionData(key string) any {
	value := s.OptSessionData(key)
	if value == nil {
		BadState("UserData is null for:", key)
	}
	return value
}

func (s Session) DeleteSessionData(key string) {
	delete(s.AppData, key)
}

func (s Session) GetStaticOrDynamicLabel(widget Widget) (string, bool) {
	Todo("?we are assuming static content is a string here")
	sc := widget.StaticContent()
	if sc != nil {
		return sc.(string), true
	} else {
		return s.WidgetStrValue(widget.Id()), false
	}
}
