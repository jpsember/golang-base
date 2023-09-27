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

type PostRequestEvent func()

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

	widgetManager   WidgetManager
	stateProvider   *WidgetStateProviderStruct
	listenerContext any

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
	postRequestEvents      []PostRequestEvent
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

func (s Session) processUpload(widgetId string) {
	pr := PrIf("Session.processUpload", true)
	pr("widget id:", widgetId)

	untypedWidget := s.WidgetManager().Opt(widgetId)
	if untypedWidget == nil {
		Alert("Can't find upload widget:", widgetId)
		return
	}

	var ok bool
	var widget FileUpload
	if widget, ok = untypedWidget.(FileUpload); !ok {
		Alert("Not an UploadWidget:", untypedWidget.Id())
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

		Todo("?How do we get the name of the file that was uploaded?")

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
		file, _, err1 := req.FormFile(widget.Id() + ".input")

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
	if problem == "" {
		err := widget.listener(s, widget, result)
		problem = StringFromOptError(err)
	}
	s.SetProblem(widget, problem)
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

func (s Session) auxHandleAjax() {
	pr := PrIf("auxHandleAjax", false)
	pr(VERT_SP, "Session.auxHandleAjax")

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

	// See if the id expression has the form <widget id> '.' <remainder>.
	// If so, treat <remainder>. as prefix for widget value

	id, remainder := ExtractFirstDotArg(widgetIdExpr)
	pr("id:", id, "remainder:", remainder)

	widgetValueExpr := s.ajaxWidgetValue
	s.ajaxWidgetValue = "" // To emphasize that we are done with this field

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
	Todo("!maybe check the lowlistener inside the ProcessWidgetValue func instead?")
	if widget.LowListener() == nil {
		Alert("#50Widget has no low-level listener:", Info(widget))
		return
	}

	// We are juggling two values:  the remainder from the id, and the ajaxValue.
	// We will join them together (where they exist) with '.'
	value := DotJoin(remainder, widgetValueExpr)
	s.ProcessWidgetValue(widget, value, nil)
}

func (s Session) ProcessWidgetValue(widget Widget, value string, context any) {
	pr := PrIf("Session.ProcessWidgetValue", false)
	pr("widget", widget.Id(), "value", QUO, value, "context", context)
	s.listenerContext = context
	updatedValue, err := widget.LowListener()(s, widget, value)
	s.listenerContext = nil
	pr("LowListener returned updatedValue:", updatedValue, "err:", err)
	s.UpdateValueAndProblem(widget, updatedValue, err)
	//
	//}
	//{
	//	if updatedValue != nil {
	//		pr("setting widget value", widget.Id(), "to:", updatedValue)
	//		s.SetWidgetValue(widget, updatedValue)
	//		Todo("!Do we always want to repaint widget if setting its value?")
	//	}
	//	if err != nil {
	//		Pr("got error from widget listener:", widget.Id(), INDENT, err)
	//	}
	//}
	//
	//// If the widget no longer exists, we may have changed pages...
	//if s.widgetManager.Opt(widget.Id()) == nil {
	//	return
	//}
	//// Always update the problem, in case we are clearing a previous error
	//s.SetWidgetProblem(widget.Id(), err)
}

func (s Session) Widget(id string) Widget {
	return s.WidgetManager().Get(id)
}

func (s Session) UpdateValueAndProblemId(widgetId string, optionalValue any, err error) {
	Alert("Would be better to refactor an make this function unnecessary")
	widget := s.Widget(widgetId)
	s.UpdateValueAndProblem(widget, optionalValue, err)
}

func (s Session) UpdateValueAndProblem(widget Widget, optionalValue any, err error) {

	if optionalValue != nil {
		s.SetWidgetValue(widget, optionalValue)
	}

	// If the widget no longer exists, we may have changed pages...
	if s.widgetManager.Opt(widget.Id()) == nil {
		return
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
	return s.RepaintId(w.Id())
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

	for _, f := range s.postRequestEvents {
		f()
	}

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
	s.postRequestEvents = nil
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

func (s Session) SetProblem(widget Widget, problem any) {
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
func (s Session) SetWidgetProblem(widgetId string, problem any) {
	s.SetProblem(s.widgetManager.Get(widgetId), problem)
}

func (s Session) auxSetWidgetProblem(widget Widget, problemText string) {
	key := WidgetIdWithProblem(widget.Id())
	state := s.State
	existingProblem := state.OptString(key, "")
	if existingProblem != problemText {
		// Pr("SetWidgetProblem:", widget.Id(), "from:", existingProblem, "to:", problemText)
		if problemText == "" {
			state.Delete(key)
		} else {
			state.Put(key, problemText)
		}
		s.RepaintId(widget.Id())
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

func (s Session) SwitchToPage(template Page, args PageArgs) {
	pr := PrIf("SwitchToPage", false)
	pr("page:", template.Name(), "from:", Caller())

	if args == nil {
		args = NewPageArgs(nil)
	}
	s.rebuildAndDisplayNewPage(func(s Session) Page {
		return template.ConstructPage(s, args)
	})
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

// If the id has the prefix, remove it.
func compileId(prefix string, id string) string {
	var out string
	if result, removed := TrimIfPrefix(id, prefix); removed {
		out = result
	} else {
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
	key := compileId(p.Prefix, id)
	if false && Alert("checking for non-existent key") {
		if !p.State.HasKey(key) {
			Pr("State has no key", QUO, key, " (id ", QUO, id, "), state:", INDENT, p.State)
			Pr("prefix:", p.Prefix)
		}
	}
	return p.State.OptString(key, "")
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

// Read widget problem, if any
func (s Session) WidgetProblem(w Widget) string {
	p := orBaseProvider(s, w.StateProvider())
	if p.State == nil {
		BadState("no state in state provider!")
	}
	return readStateStringValue(p, WidgetIdWithProblem(w.Id()))
}

// Read widget value; assumed to be a string.
func (s Session) WidgetStringValue(w Widget) string {
	p := orBaseProvider(s, w.StateProvider())
	if p.State == nil {
		BadState("no state in state provider!")
	}
	return readStateStringValue(p, w.Id())
}

func (s Session) SetWidgetValue(w Widget, value any) {
	if s.SetValue(w.Id(), w.StateProvider(), value) {
		s.Repaint(w)
	}
}

// I separated this out from SetWidgetValue, since we may want to update values given just ids and state providers
func (s Session) SetValue(widgetId string, provider WidgetStateProvider, value any) bool {
	pr := PrIf("SetValue", false)
	p := orBaseProvider(s, provider)
	id := compileId(p.Prefix, widgetId)
	oldVal := p.State.OptUnsafe(id)
	changed := value != oldVal
	pr("old:", oldVal, "new:", value, "changed:", changed)
	if changed {
		p.State.Put(id, value)
	}
	return changed
}

// Get the context for the current listener.  For list items, this will be the list element id.
func (s Session) Context() any {
	return s.listenerContext
}

// This merges a couple of separate functions, to reduce the complexity.
func (s Session) rebuildAndDisplayNewPage(pageProvider func(s Session) Page) {
	// Dispose of any existing widgets
	m := s.WidgetManager()
	m.widgetMap = make(map[string]Widget)

	// Build a new page widget
	s.PageWidget = m.Id(WidgetIdPage).Open()
	m.Close()

	// Get the new page (it is now safe to construct, as the old widgets are gone)
	page := pageProvider(s)
	CheckState(page != nil, "no page was provided")
	s.debugPage = page

	// Display the new page
	Todo("!Verify that this works for normal refreshes as well as ajax operations")
	s.Repaint(s.PageWidget)
	s.browserURLExpr = s.ConstructPathFromPage(page)
}

func (s Session) Validate(widget Widget) {
	pr := PrIf("Session.Validate", false)
	pr("id:", widget.Id())
	if widget.LowListener() != nil {
		p := orBaseProvider(s, widget.StateProvider())
		id := compileId(p.Prefix, widget.Id())
		value := p.State.OptUnsafe(id)
		pr(" value from state:", Info(value))
		if jstr, ok := value.(JString); ok {
			str := jstr.AsString()
			pr("...processing widget value", QUO, str)
			s.ProcessWidgetValue(widget, str, nil)
		} else {
			Alert("#50<1Don't know how to validate widget", widget.Id(), "with value", Info(value))
			Todo("Perhaps we need a widget method that returns the current widget's value as 'any'")
		}
	}
	for _, child := range widget.Children() {
		s.Validate(child)
	}
}

// Schedule an event to be executed after the current AJAX request handling has completed.  For example,
// additional widget validations that are triggered by a current validation.
func (s Session) AddPostRequestEvent(event PostRequestEvent) {
	s.postRequestEvents = append(s.postRequestEvents, event)
}
