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

var ValidateWidgetMarkup = false && Alert("ValidateWidgetMarkup is true")

type Session = *SessionStruct

type PostRequestEvent func()

type SessionStruct struct {
	WidgetManagerObj
	SessionId string

	// For storing an application Oper, for example
	appData map[string]any

	// widget representing the entire page; nil if not constructed yet
	PageWidget Widget
	// lock for making request handling thread safe; we synchronize a particular session's requests
	lock sync.RWMutex
	// JSMap containing widget values, other user session state
	//State JSMap

	BrowserInfo webserv_data.ClientInfo
	debugPage   Page // Used only to get the current page's name for rendering in the user header

	app any // ServerApp is stored here, will clean up later

	//stateProvider   *WidgetStateProviderStruct
	listenerContext any

	// Current request variables
	ResponseWriter         http.ResponseWriter
	request                *http.Request
	requestProblem         error  // If not nil, problem detected with current request
	clientInfoString       string // If nonempty information sent from client about screen size, etc
	ajaxWidgetId           string // Id of widget that ajax call is being sent to
	ajaxWidgetValue        string // The string representation of the ajax widget's requested value (if there was one)
	browserURLExpr         string // If not nil, client browser should push this onto the history
	repaintWidgetMarkupMap JSMap  // Used only during repainting; the map of widget ids -> markup to be repainted by client
	postRequestEvents      []PostRequestEvent
}

func NewSession() Session {
	s := SessionStruct{
		//State:       NewJSMap(),
		BrowserInfo: webserv_data.DefaultClientInfo,
		appData:     make(map[string]any),
	}
	s.InitializeWidgetManager()
	//s.setBaseStateProvider(NullStateProvider)
	Todo("!Restore user session from filesystem/database")
	Todo("?ClientInfo (browser info) not sent soon enough")
	Todo("?The Session should have WidgetManager embedded within it, so we can call through to its methods")
	return &s
}

func DiscardAllSessions(sessionManager SessionManager) {
	loggedInUsersSetLock.Lock()
	defer loggedInUsersSetLock.Unlock()

	Alert("Discarding all sessions")
	sessionManager.DiscardAllSessions()
	dbPr("DiscardAllSessions, cleared")
	loggedInUsersSet.Clear()
}

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

var loggedInUsersSet = NewSet[int]()
var loggedInUsersSetLock sync.RWMutex

func (s Session) PrependId(id string) string {
	p := s.StateProvider()
	if p == nil {
		return id
	}
	Pr("prepending prefix", p.Prefix, "to id:", id)
	return p.Prefix + id
}

func (s Session) PrepareForHandlingRequest(w http.ResponseWriter, req *http.Request) {
	s.ResponseWriter = w
	s.request = req
}

func (s Session) ToJson() *JSMapStruct {
	m := NewJSMap()
	m.Put("id", s.Id)
	return m
}

func ParseSession(source JSEntity) Session {
	var s = source.(*JSMapStruct)
	var n = NewSession()
	n.SessionId = s.OptString("id", "")
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

	untypedWidget := s.Opt(widgetId)
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
	pr("start handling")

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

	widget := s.Opt(id)
	if widget == nil {
		Pr("no widget with id", Quoted(id), "found to handle value", Quoted(widgetValueExpr))
		Pr("state provider:", s.StateProvider())
		Pr("widget map:", INDENT, s.widgetMap)
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
}

func (s Session) UpdateValueAndProblem(widget Widget, optionalValue any, err error) {
	if optionalValue != nil {
		s.SetWidgetValue(widget, optionalValue)
	}
	// If the widget no longer exists, we may have changed pages...
	if !s.exists(widget.Id()) {
		return
	}
	Pr("updateValueAndProblem, widget id:", widget.Id(), "err:", err)
	// Always update the problem, in case we are clearing a previous error
	s.SetProblem(widget, err)
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

// Traverse a widget tree, rendering widgets that have been marked for repainting.
func (s Session) processRepaintFlags(w Widget) {
	// For each widget that has been marked for repainting, we send it and its markup
	// to the client.  The children need not be descended to, as they will be repainted
	// by their containers.
	if w.IsRepaint() {
		m := NewMarkupBuilder()
		RenderWidget(w, s, m)
		content := m.String()
		if ValidateWidgetMarkup {
			mp, err := SharedHTMLValidator().Validate(content)
			if err != nil {
				Pr(VERT_SP, "Markup failed validation:", INDENT, content)
				Pr(mp)
				BadState("failed validation")
			}
		}
		s.repaintWidgetMarkupMap.Put(w.Id(), content)
		w.ClearRepaint()
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
	pr := PrIf("sendAjaxResponse", debRepaint)

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

// ------------------------------------------------------------------------------------
// Widget problems
// ------------------------------------------------------------------------------------

var prProb = PrIf("Widget Problems", false)

// Read widget problem.  Returns an empty string if it hasn't got one.
func (s Session) WidgetProblem(w Widget) string {
	pr := prProb
	p := s.provider(w)
	if p == nil {
		return ""
	}
	result := readStateStringValue(p, widgetProblemKey(w))
	pr("problem for", w.Id(), "is:", QUO, result)
	return result
}

func (s Session) SetProblem(widget Widget, problem any) {
	pr := prProb
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
	p := s.provider(widget)
	if p == nil {
		CheckState(text == "", "no state provider")
		return
	}
	key := widgetProblemKey(widget)
	state := p.State
	existingProblem := state.OptString(key, "")
	if existingProblem != text {
		pr("changing from:", QUO, existingProblem, "to:", QUO, text)
		if text == "" {
			state.Delete(key)
		} else {
			state.Put(key, text)
		}
		widget.Repaint()
	}
}

func (s Session) SetWidgetProblem(widgetId string, problem any) {
	s.SetProblem(s.Get(widgetId), problem)
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

// ------------------------------------------------------------------------------------

func SessionApp(s Session) ServerApp {
	return s.app.(ServerApp)
}

// ------------------------------------------------------------------------------------
// Accessing widget values
// ------------------------------------------------------------------------------------

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

//func (s Session) baseStateProvider() WidgetStateProvider {
//	return s.stateProvider
//}
//
//func (s Session) setBaseStateProvider(p WidgetStateProvider) {
//	s.stateProvider = p
//}

// ------------------------------------------------------------------------------------
// Reading widget state values
// ------------------------------------------------------------------------------------
// Read widget value; assumed to be a string.
func (s Session) WidgetStringValue(w Widget) string {
	pr := PrIf("WidgetStringValue", true)

	// If the session has a stacked state provider and its (non-empty) prefix matches this widget's id,
	// take state from that instead.
	id := w.Id()
	p := s.provider(w)
	pr(VERT_SP, "id:", id, "provider:", p)

	state := s.stackedState()
	pr("stacked state:", state.StateProvider)

	if state.IdPrefix != "" {
		effectiveId, hadPrefix := TrimIfPrefix(id, state.StateProvider.Prefix)
		pr("TrimIfPrefix produced:", effectiveId, hadPrefix)
		if hadPrefix {
			id = effectiveId
			p = state.StateProvider
			pr("using override state provider:", p)
		}
	}

	result := readStateStringValue(p, id)
	pr("result:", result)
	return result
}

// Read widget value; assumed to be an int.
func (s Session) WidgetIntValue(w Widget) int {
	p := s.provider(w)
	return readStateIntValue(p, w.Id())
}

// Read widget value; assumed to be a bool.
func (s Session) WidgetBoolValue(w Widget) bool {
	p := s.provider(w)
	return readStateBoolValue(p, w.Id())
}

func (s Session) SetWidgetValue(w Widget, value any) {
	pr := PrIf("SetWidgetValue", false)
	p := s.provider(w)
	id := compileId(p.Prefix, w.Id())
	oldVal := p.State.OptUnsafe(id)
	changed := value != oldVal
	pr("old:", oldVal, "new:", value, "changed:", changed)
	if changed {
		p.State.Put(id, value)
		w.Repaint()
	}
}

func (s Session) provider(w Widget) WidgetStateProvider {
	Todo("The Session state provider methods should maybe be deleted, and use only the WidgetManager ones")
	p := w.StateProvider()
	if p == nil {
		p = s.StateProvider()
		if p == nil {
			BadState("there is no session state provider for id:", w.Id())
		}
	}
	return p
}

// Get the context for the current listener.  For list items, this will be the list element id.
func (s Session) Context() any {
	return s.listenerContext
}

// This merges a couple of separate functions, to reduce the complexity.
func (m Session) rebuildAndDisplayNewPage(pageProvider func(s Session) Page) {
	// Dispose of any existing widgets
	m.widgetMap = make(map[string]Widget)

	// Build a new page widget
	m.PageWidget = m.Id(WidgetIdPage).Open()
	m.Close()

	// Get the new page (it is now safe to construct, as the old widgets are gone)
	page := pageProvider(m)
	CheckState(page != nil, "no page was provided")
	m.debugPage = page
	//Pr(VERT_SP, "changed page to", page.Name(), INDENT, Callers(1, 5), VERT_SP)

	// Display the new page
	Todo("!Verify that this works for normal refreshes as well as ajax operations")
	m.PageWidget.Repaint()

	//func (s Session) constructPathFromPage(page Page) string {
	m.browserURLExpr = "/" + page.Name() + "/" + strings.Join(page.Args(), "/")
	//return result
	//}
	//	m.browserURLExpr = m.constructPathFromPage(page)
}

func (s Session) ValidateAndCountErrors(widget Widget) int {
	s.Validate(widget)
	return s.WidgetErrorCount(widget)
}

func (s Session) WidgetErrorCount(widget Widget) int {
	Todo("?Use 's' instead of 'sess' everywhere")
	count := 0
	problemText := s.WidgetProblem(widget)
	if problemText != "" {
		count++
	}
	for _, child := range widget.Children() {
		count += s.WidgetErrorCount(child)
	}
	return count
}

func (s Session) Validate(widget Widget) {
	pr := PrIf("Session.Validate", false)
	pr("id:", widget.Id())
	if widget.LowListener() != nil {
		valAsString, applicable := widget.ValidationValue(s)
		if applicable {
			pr("...calling low level listener with", QUO, valAsString)
			updatedValue, err := widget.LowListener()(s, widget, valAsString)
			pr("updated value, err:", updatedValue, err)
			if DebugWidgetRepaint {
				Pr("Session.Validate, widget has changed value to:", QUO, valAsString)
			}
			s.UpdateValueAndProblem(widget, updatedValue, err)
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
