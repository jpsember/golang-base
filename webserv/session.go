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

var dbPr = PrIf("", false)

var loggedInUsersSet = NewSet[int]()
var loggedInUsersSetLock sync.RWMutex

func IsUserLoggedIn(userId int) bool {
	loggedInUsersSetLock.Lock()
	defer loggedInUsersSetLock.Unlock()
	result := loggedInUsersSet.Contains(userId)
	dbPr("IsUserLoggedIn:", userId, result)
	return result
}

func TryRegisteringUserAsLoggedIn(userId int, loggedInState bool) bool {
	loggedInUsersSetLock.Lock()
	defer loggedInUsersSetLock.Unlock()
	currentState := loggedInUsersSet.Contains(userId)
	changed := currentState != loggedInState
	if changed {
		if loggedInState {
			loggedInUsersSet.Add(userId)
			dbPr("Registering user as logged in:", userId)

		} else {
			loggedInUsersSet.Remove(userId)
			dbPr("Unregistring user as logged in:", userId)
		}
	}
	return changed
}

func LogUserOut(userId int) bool {
	loggedInUsersSetLock.Lock()
	defer loggedInUsersSetLock.Unlock()
	wasLoggedIn := loggedInUsersSet.Contains(userId)
	if wasLoggedIn {
		loggedInUsersSet.Remove(userId)
		dbPr("LogUserOut, Unregistring user as logged in:", userId)
	}
	return wasLoggedIn
}

func DiscardAllSessions(sessionManager SessionManager) {
	loggedInUsersSetLock.Lock()
	defer loggedInUsersSetLock.Unlock()

	Alert("Discarding all sessions")
	sessionManager.DiscardAllSessions()
	dbPr("DiscardAllSessions, cleared")
	loggedInUsersSet.Clear()
}

type Session = *SessionStruct

type SessionStruct struct {
	Id string

	// For storing an application Oper, for example
	appData map[string]any

	// widget representing the entire page; nil if not constructed yet
	PageWidget Widget
	// lock for making request handling thread safe; we synchronize a particular session's requests
	lock sync.RWMutex
	// JSMap containing widget values, other user session state
	State JSMap

	BrowserInfo webserv_data.ClientInfo
	debugPage   Page // Used only to get the current page's name for rendering in the user header

	app any // ServerApp is stored here, will clean up later

	widgetManager WidgetManager
	stateProvider *WidgetStateProviderStruct
	//baseIdPrefix  string // This is modified for special rendering operations, such as list items

	// Current request variables
	ResponseWriter         http.ResponseWriter
	request                *http.Request
	requestProblem         error  // If not nil, problem detected with current request
	clientInfoString       string // If nonempty information sent from client about screen size, etc
	ajaxWidgetId           string // Id of widget that ajax call is being sent to
	ajaxWidgetValue        string // The string representation of the ajax widget's requested value (if there was one)
	browserURLExpr         string // If not nil, client browser should push this onto the history
	repaintSet             StringSet
	repaintWidgetMarkupMap JSMap // Used only during repainting; the map of widget ids -> markup to be repainted by client
}

func NewSession() Session {
	s := SessionStruct{
		State:       NewJSMap(),
		BrowserInfo: webserv_data.DefaultClientInfo,
		appData:     make(map[string]any),
	}
	s.SetBaseStateProvider(NewStateProvider("", s.State))
	Todo("!Restore user session from filesystem/database")
	Todo("?ClientInfo (browser info) not sent soon enough")
	Todo("?The Session should have WidgetManager embedded within it, so we can call through to its methods")
	return &s
}

func (s Session) PrependId(id string) string {
	return s.BaseStateProvider().Prefix + id
}

func (s Session) PrepareForHandlingRequest(w http.ResponseWriter, req *http.Request) {
	s.ResponseWriter = w
	s.request = req
	s.repaintSet = NewStringSet()
}

// Get WidgetManager for this session, creating one if necessary
func (s Session) WidgetManager() WidgetManager {
	if s.widgetManager == nil {
		s.widgetManager = NewWidgetManager()
	}
	return s.widgetManager
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
func (s Session) HandleAjaxRequest() {
	s.parseAjaxRequest()
	if false && Alert("dumping") {
		Pr("Query:", INDENT, s.request.URL.Query())
	}
	s.auxHandleAjax()
	s.sendAjaxResponse()
}

func (s Session) HandleUploadRequest(widgetId string) {
	s.processUpload(widgetId)
	// Send the usual ajax response
	s.sendAjaxResponse()
}

func (s Session) processUpload(uploadWidgetId string) {

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
		req := s.request

		problem = "upload request was not POST"
		if req.Method != "POST" {
			break
		}

		// From https://freshman.tech/file-upload-golang/

		problem = "The uploaded file is too big. Please choose an file that's less than 10MB in size"
		{
			const MAX_UPLOAD_SIZE = 10_000_000
			req.Body = http.MaxBytesReader(s.ResponseWriter, req.Body, MAX_UPLOAD_SIZE)
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
func (s Session) HandleResourceRequest(resourcePath Path) error {

	var err error
	resource := s.request.URL.Path
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

	WriteResponse(s.ResponseWriter, InferContentTypeM(resource), content)
	return err
}

func (s Session) parseAjaxRequest() {
	// At present, the ajax request parameters are of the form
	//  /ajax? [expr [& expr]*]
	// where expr is:
	//  w=<widget id>
	//  v=<widget value>
	//  i=<client information as json map, encoded as string>
	v := s.request.URL.Query()

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

func extractId(expr string) (string, string) {
	dotPos := strings.IndexByte(expr, '.')
	if dotPos >= 0 {
		return expr[0:dotPos], expr[dotPos+1:]
	}
	return expr, ""
}

func (s Session) auxHandleAjax() {
	pr := PrIf("auxHandleAjax", true)
	pr("Session.auxHandleAjax")

	didSomething := false

	// Process client info, if it was sent
	if s.clientInfoString != "" {
		pr("...processing client info:", s.clientInfoString)
		s.processClientInfo(s.clientInfoString)
		didSomething = true
	}

	// We can now assume that the request consists of a single widget id, and perhaps a single value
	// for that widget

	widgetIdExpr := s.ajaxWidgetId
	pr("widgetIdExpr:", widgetIdExpr)
	if widgetIdExpr == "" {
		if !didSomething {
			s.SetRequestProblem("widget id was empty")
		}
		return
	}

	// See if the id expression has the form <widget id> '.' <remainder>
	//

	id, remainder := extractId(widgetIdExpr)
	pr("id:", id, "remainder:", remainder)

	widgetValueExpr := s.ajaxWidgetValue
	s.ajaxWidgetValue = "" // To emphasize that we are done with this field

	// We are juggling two values:  the remainder from the id, and the ajaxValue

	Todo("!Clarify difference between a widget 'low level listener' and its possible clickListener")

	widget := s.widgetManager.Opt(id)
	if widget == nil {
		pr("no widget with id", Quoted(id), "found to handle value", Quoted(widgetValueExpr))
		return
	}
	pr("found widget with id:", id, "and type:", TypeOf(widget))

	if !widget.Enabled() {
		s.SetRequestProblem("widget is disabled", widget)
		return
	}

	if widget.LowListener() == nil {
		Alert("#50Widget has no low-level listener:", Info(widget))
		return
	}

	value := remainder
	if value == "" {
		value = widgetValueExpr
	} else if widgetValueExpr != "" {
		value = value + "." + widgetValueExpr
	}

	pr("calling LowListener for id:", widget.Id(), "with value:", value)
	updatedValue, err := widget.LowListener()(s, widget, value)
	{
		if err != nil {
			Pr("got error from widget listener:", widget.Id(), INDENT, err)
		} else if updatedValue != nil {
			s.SetWidgetValue(widget, updatedValue)
		}
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
	n, err := ParseOrDefault(json, s.BrowserInfo)
	if !ReportIfError(err, "Trouble parsing BrowserInfo") {
		s.BrowserInfo = n.(webserv_data.ClientInfo)
	}
}

// Mark a widget for repainting
func (s Session) Repaint(w Widget) Session {
	pr := PrIf("", debRepaint)
	pr("Repaint:", w)
	if s.repaintSet.Add(w.Id()) {
		pr("...adding to set")
	}
	return s
}

// Traverse a widget tree, rendering widgets that have been marked for repainting.
func (s Session) processRepaintFlags(w Widget) {
	// For each widget that has been marked for repainting, we send it and its markup
	// to the client.  The children need not be descended to, as they will be repainted
	// by their containers.
	id := w.Id()
	if s.repaintSet.Contains(id) {
		m := NewMarkupBuilder()
		RenderWidget(w, s, m)
		s.repaintWidgetMarkupMap.Put(id, m.String())
	} else {
		for _, c := range w.Children() {
			s.processRepaintFlags(c)
		}
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
	pr := PrIf("", debRepaint)

	jsmap := NewJSMap()
	s.repaintWidgetMarkupMap = NewJSMap()
	s.processRepaintFlags(s.PageWidget)
	jsmap.Put(respKeyWidgetsToRefresh, s.repaintWidgetMarkupMap)
	s.repaintWidgetMarkupMap = nil

	expr := s.browserURLExpr
	if expr != "" {
		jsmap.Put(respKeyURLExpr, expr)
	}
	pr("sending back to Ajax caller:", INDENT, jsmap)
	content := jsmap.CompactString()
	WriteResponse(s.ResponseWriter, "application/json", []byte(content))
}

// Discard state added to session to serve a request.
func (s Session) ReleaseLockAndDiscardRequest() {
	problem := s.requestProblem
	if problem != nil {
		Pr("Problem processing client message:", INDENT, problem)
	}
	s.ResponseWriter = nil
	s.request = nil
	s.requestProblem = nil
	s.ajaxWidgetId = ""
	s.ajaxWidgetValue = ""
	s.clientInfoString = ""
	s.browserURLExpr = ""
	s.repaintSet = nil
	s.lock.Unlock()
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
		s.RepaintIds(widgetId)
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

var logSessionData = false && Alert("logging session data")

func (s Session) PutSessionData(key string, value any) {
	if logSessionData {
		Pr("Storing session data", key, "=>", TypeOf(value))
	}
	s.appData[key] = value
}

func (s Session) OptSessionData(key string) any {
	value := s.appData[key]
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
	delete(s.appData, key)
}

func (s Session) GetStaticOrDynamicLabel(widget Widget) (string, bool) {
	return s.WidgetStringValue(widget), false
}

// ------------------------------------------------------------------------------------
// Page url and arguments
// ------------------------------------------------------------------------------------

func (s Session) SwitchToPage(page Page) {
	//Pr("SwitchToPage:", page.Name(), "from:", Caller())
	s.Repaint(s.PageWidget)
	s.browserURLExpr = s.ConstructPathFromPage(page)
	s.debugPage = page
}

func (s Session) NewBrowserPath() string {
	return s.browserURLExpr
}

func (s Session) ConstructPathFromPage(page Page) string {
	result := "/" + page.Name() + "/" + strings.Join(page.Args(), "/")
	return result
}

// ------------------------------------------------------------------------------------

func SessionApp(s Session) ServerApp {
	return s.app.(ServerApp)
}

// ------------------------------------------------------------------------------------
// Accessing widget values
// ------------------------------------------------------------------------------------

type WidgetStateProviderStruct struct {
	Prefix string // A prefix to remove from the id before constructing its map key
	State  JSMap  // The map containing the state
}

type WidgetStateProvider = *WidgetStateProviderStruct

func (pv WidgetStateProvider) String() string {
	return "{SP " + Quoted(pv.Prefix) + " state:" + Truncated(pv.State.CompactString()) + " }"
}

func NewStateProvider(prefix string, state JSEntity) WidgetStateProvider {
	return &WidgetStateProviderStruct{Prefix: prefix, State: state.AsJSMap()}
}

// If state provider is nil, use default one
func orBaseProvider(s Session, p WidgetStateProvider) WidgetStateProvider {
	if p == nil {
		// This is the state provider if no other one has been specified
		p = s.BaseStateProvider()
	}
	return p
}

func compileId(prefix string, id string) string {
	// If the id has the prefix, remove it
	var out string
	if result, removed := TrimIfPrefix(id, prefix); removed {
		out = result
	} else {
		Alert("Only SUBTRACTING")
		out = id
	}
	//Pr("compileId, prefix:", Quoted(prefix), "id:", id, "returning:", out)
	return out
}

func (s Session) BaseStateProvider() WidgetStateProvider {
	return s.stateProvider
}

func (s Session) SetBaseStateProvider(p WidgetStateProvider) {
	s.stateProvider = p
}

// ------------------------------------------------------------------------------------
// Reading widget state values
// ------------------------------------------------------------------------------------

// Read widget value; assumed to be an int.
func readStateIntValue(p WidgetStateProvider, id string) int {
	return p.State.OptInt(compileId(p.Prefix, id), 0)
}

// Read widget value; assumed to be a bool.
func readStateBoolValue(p WidgetStateProvider, id string) bool {
	return p.State.OptBool(compileId(p.Prefix, id), false)
}

// Read widget value; assumed to be a string.
func readStateStringValue(p WidgetStateProvider, id string) string {
	return p.State.OptString(compileId(p.Prefix, id), "")
}

// Read widget value; assumed to be an int.
func (s Session) IntValue(id string) int {
	return readStateIntValue(s.BaseStateProvider(), id)
}

// Read widget value; assumed to be a boolean.
func (s Session) BoolValue(id string) bool {
	return readStateBoolValue(s.BaseStateProvider(), id)
}

// Read widget value; assumed to be a string.
func (s Session) StringValue(id string) string {
	return readStateStringValue(s.BaseStateProvider(), id)
}

// Read widget value; assumed to be an int.
func (s Session) WidgetIntValue(w Widget) int {
	p := orBaseProvider(s, w.StateProvider())
	return readStateIntValue(p, w.Id())
}

// Read widget value; assumed to be a bool.
func (s Session) WidgetBoolValue(w Widget) bool {
	p := orBaseProvider(s, w.StateProvider())
	return readStateBoolValue(p, w.Id())
}

// Read widget value; assumed to be a string.
func (s Session) WidgetStringValue(w Widget) string {
	p := orBaseProvider(s, w.StateProvider())
	return readStateStringValue(p, w.Id())
}

func (s Session) SetWidgetValue(w Widget, value any) {
	p := orBaseProvider(s, w.StateProvider())
	p.State.Put(compileId(p.Prefix, w.Id()), value)
}
