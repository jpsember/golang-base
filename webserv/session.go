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

type ClickListener func(sess *SessionStruct, id string)

//type URLRequestHandler func(sess *SessionStruct, expr string) bool

type Session = *SessionStruct

type SessionStruct struct {
	Id string

	// For storing an application Oper, for example
	AppData map[string]any

	// widget representing the entire page; nil if not constructed yet
	PageWidget Widget
	// Lock for making request handling thread safe; we synchronize a particular session's requests
	Lock sync.RWMutex
	// JSMap containing widget values, other user session state
	State JSMap

	BrowserInfo   webserv_data.ClientInfo
	widgetManager WidgetManager
	clickListener ClickListener
	//URLRequestHandler URLRequestHandler

	// Current request variables
	responseWriter   http.ResponseWriter
	request          *http.Request
	requestProblem   error  // If not nil, problem detected with current request
	clientInfoString string // If nonempty information sent from client about screen size, etc
	ajaxWidgetId     string // Id of widget that ajax call is being sent to
	ajaxWidgetValue  string // The string representation of the ajax widget's requested value (if there was one)
	pendingURLExpr   string // If not nil, client browser should push this onto the history
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
	Todo("?The Session should have WidgetManager embedded within it, so we can call through to its methods")
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

	// A value is optional, as buttons don't send them.
	if len(t1) == 1 && len(t2) <= 1 {
		s.ajaxWidgetId = t1[0]
		if len(t2) == 1 {
			s.ajaxWidgetValue = t2[0]
		}
	}
	clientInfoArray := v[clientKeyInfo]
	if clientInfoArray != nil && len(clientInfoArray) == 1 {
		s.clientInfoString = clientInfoArray[0]
	}

}

func (s Session) processClientMessage() {

	didSomething := false

	// Process client info, if it was sent
	if s.clientInfoString != "" {
		s.processClientInfo(s.clientInfoString)
		didSomething = true
	}

	// We can now assume that the request consists of a single widget id, and perhaps a single value
	// for that widget
	//

	widgetId := s.ajaxWidgetId
	if widgetId == "" {
		if !didSomething {
			s.SetRequestProblem("widget id was empty")
		}
		return
	}

	widget := s.widgetManager.Opt(widgetId)
	// If there is no widget with this id, inform the default listener (clarify the terminology later)
	if widget == nil {
		s.processClickEvent(widgetId)
		return
	}

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

func (s Session) processClickEvent(sourceId string) {
	listener := s.clickListener
	if listener == nil {
		Alert("#50No ClickListener for id:" + sourceId)
		return
	}
	listener(s, sourceId)
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
const respKeyURLExpr = "u"

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
	if s.pendingURLExpr != "" {
		jsmap.Put(respKeyURLExpr, s.pendingURLExpr)
		Pr("sending url expression:", s.pendingURLExpr)
	}
	pr("sending back to Ajax caller:", INDENT, jsmap)
	content := jsmap.CompactString()
	WriteResponse(s.responseWriter, "application/json", []byte(content))
}

// Discard state added to session to serve a request.
func (s Session) ReleaseLockAndDiscardRequest() {
	Todo("This can be done for *all* requests to simplify things")
	problem := s.requestProblem
	if problem != nil {
		Pr("Problem processing client message:", INDENT, problem)
	}
	s.responseWriter = nil
	s.request = nil
	s.requestProblem = nil
	s.ajaxWidgetId = ""
	s.ajaxWidgetValue = ""
	s.clientInfoString = ""
	s.pendingURLExpr = ""

	Todo("!Consider moving the repaint set from the widget manager to the session, and perhaps other things")
	s.WidgetManager().clearRepaintSet()

	s.Lock.Unlock()
}

func (s Session) SetRequestError(problem error) error {
	if problem != nil && s.requestProblem == nil {
		s.requestProblem = problem
		Alert("#50<2 setting request problem:", s.requestProblem)
	}
	return s.requestProblem
}

func (s Session) SetRequestProblem(message ...any) error {
	return s.SetRequestError(Error("Problem with ajax request: " + ToString(message...)))
}

func (s Session) GetRequestProblem() error {
	return s.requestProblem
}

func (s Session) Ok() bool {
	return s.requestProblem == nil
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

var logSessionData = false && Alert("logging session data")

func (s Session) PutSessionData(key string, value any) {
	if logSessionData {
		Pr("Storing session data", key, "=>", TypeOf(value))
	}
	s.AppData[key] = value
}

func (s Session) OptSessionData(key string) any {
	value := s.AppData[key]
	if logSessionData {
		Pr("Getting session data", key, "=>", TypeOf(value))
	}
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

func (s Session) SetClickListener(listener ClickListener) {
	s.clickListener = listener
}

// Cause a new URL to be pushed onto the browser history.  This gets sent when the AJAX response is sent.
//
// Working from blog:  https://css-tricks.com/using-the-html5-history-api/
func (s Session) SetURLExpression(args ...any) {
	pr := PrIf(true)
	pr("SetURLExpression")

	sb := strings.Builder{}

	for _, arg := range args {
		s := strings.TrimSpace(ToString(arg))
		pr("... ", sb, " + ", Quoted(s))
		if sb.Len() != 0 {
			sb.WriteByte('/')
		}
		sb.WriteString(s)
	}
	s.pendingURLExpr = sb.String()
	pr("pendingURLExpr:", s.pendingURLExpr)

	// Ok, when clicking a button it is not appending the button url to the program 'root' url, rather the current one
	// Solved by using location.origin

	Todo("!What about: 'Make sure to return true from Javascript click handlers when people middle or command click so that we don’t override them accidentally.'")

}
