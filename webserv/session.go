package webserv

import (
	"bytes"
	. "github.com/jpsember/golang-base/base"
	"github.com/jpsember/golang-base/webserv/gen/webserv_data"
	"io"
	"net/http"
	"regexp"
	"strings"
	"sync"
)

var dbPr = PrIf("", false)

var ValidateWidgetMarkup = false && Alert("ValidateWidgetMarkup is true")

type Session = *SessionStruct

type PostRequestEvent func()

type PendingPage struct {
	template Page
	args     PageArgs
}

type SessionStruct struct {
	WidgetManagerObj
	SessionId string

	// For storing an application Oper, for example
	appData map[string]any

	// widget representing the entire page; nil if not constructed yet
	PageWidget Widget
	// lock for making request handling thread safe; we synchronize a particular session's requests
	lock sync.RWMutex

	BrowserInfo webserv_data.ClientInfo
	debugPage   Page // Used only to get the current page's name for rendering in the user header

	app any // ServerApp is stored here, will clean up later

	listenerContext []string

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
	pendingPage            *PendingPage
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
	var result string
	pref := s.IdPrefix()
	if pref == "" {
		result = id
	} else {
		if strings.HasSuffix(id, pref) {
			Alert("<1#50Id already has prefix:", id, pref, INDENT, Callers(1, 4))
		}
		result = pref + id
	}
	return ValidateHTMLId(result)
}

// ID and NAME tokens must begin with a letter ([A-Za-z]) and may be followed by any number of letters,
//
//	digits ([0-9]), hyphens ("-"), underscores ("_"), colons (":"), and periods (".").
var validHTMLIdExpr = CheckOkWith(regexp.Compile(`^[A-Za-z](?:[A-Za-z0-9\-_:.])*$`))

func ValidateHTMLId(id string) string {
	if !validHTMLIdExpr.MatchString(id) {
		BadArg("Invalid id:", QUO, id)
	}
	return id
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
	Todo("!Do we possibly need to do this during a non-ajax interaction?")
	s.ProcessPendingPage()
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
	pr := PrIf("auxHandleAjax", true)
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

	widgetValueExpr := s.ajaxWidgetValue
	s.ajaxWidgetValue = "" // To emphasize that we are done with this field
	pr("widgetValueExpr:", QUO, widgetValueExpr)

	widgetIdExpr := s.ajaxWidgetId
	pr("widgetIdExpr:", QUO, widgetIdExpr)
	if widgetIdExpr == "" {
		if !didSomething {
			s.SetRequestProblem("widget id was empty")
		}
		return
	}

	// We will assume that any periods in the widgetIdExpr serve to separate the widget id from additional context

	id := widgetIdExpr
	//id, remainder := ExtractFirstDotArg(widgetIdExpr)
	//pr("after parsing id expression, id:", QUO, id, "remainder:", remainder)

	Todo("!Is the old context still needed?")

	colonArgs, indices := extractColonSeparatedArgs(id)

	var widget Widget
	var args []string
	{
		for j := len(colonArgs); j >= 0; j-- {
			var candidate string
			args = colonArgs[j:]
			if j == len(colonArgs) {
				candidate = id
			} else {
				candidate = id[0 : indices[j]-1]
			}
			pr("....looking for widget with id:", QUO, candidate, "and args:", args)
			widget = s.Opt(candidate)
			if widget != nil {
				pr("................FOUND!")
				break
			}
		}
	}

	if widget == nil {
		Pr("no widget with id", Quoted(id), "found to handle value", Quoted(widgetValueExpr))
		Pr("state provider:", s.stackedStateProvider())
		Pr("widget map:", INDENT, s.widgetMap)
		return
	}

	Todo("Do something with args:", args)
	pr(VERT_SP, "found widget with id:", QUO, widget.Id(), "and type:", TypeOf(widget), "args:", args, VERT_SP)

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
	value := widgetValueExpr
	pr(VERT_SP, "processing widget value, widget:", QUO, widget.Id(), "value:", QUO, value, "args:", args, VERT_SP)
	s.ProcessWidgetValue(widget, value, args)
}

func extractColonSeparatedArgs(expr string) ([]string, []int) {
	pr := PrIf("extractColonSeparatedArgs", false)
	pr("expr:", QUO, expr)
	var result []string
	var indices []int
	lastIndex := 0
	cursor := 0
	cmax := len(expr)
	for cursor < cmax {
		newCursor := cursor + 1
		pr("cursor:", cursor, "remaining chars:", expr[cursor:])
		if expr[cursor] == ':' {
			result = append(result, expr[lastIndex:cursor])
			indices = append(indices, lastIndex)
			lastIndex = newCursor
		}
		cursor = newCursor
	}
	result = append(result, expr[lastIndex:cursor])
	indices = append(indices, lastIndex)

	//result = append(result, colonArg{index: lastIndex, value: expr[lastIndex:]})
	pr("returning:", result, indices)
	return result, indices
}

func (s Session) ProcessWidgetValue(widget Widget, value string, args []string) {
	//REFACTOR TO ACCEPT 'colon' ARGUMENTS (as distinct strings)

	pr := PrIf("Session.ProcessWidgetValue", true)
	pr(VERT_SP, "widget", widget.Id(), "value", QUO, value, "args", args)
	s.listenerContext = args
	updatedValue, err := widget.LowListener()(s, widget, value, args)
	s.listenerContext = nil
	pr("LowListener returned updatedValue:", updatedValue, "err:", err)
	s.UpdateValueAndProblem(widget, updatedValue, err)
}

func (s Session) UpdateValueAndProblem(widget Widget, optionalValue any, err error) {
	pr := PrIf("UpdateValueAndProblem", true)
	pr("UpdateValueAndProblem, id:", widget.Id(), "optionalValue:", QUO, optionalValue, "err:", err)
	if optionalValue != nil {
		s.SetWidgetValue(widget, optionalValue)
	}
	// If the widget no longer exists, we may have changed pages...
	if !s.exists(widget.Id()) {
		pr("widget doesn't exist!")
		return
	}
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
	id, p := s.getStateProvider(w)
	key := widgetProblemKey(id)
	result := p.OptString(key, "")
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
	id, p := s.getStateProvider(widget)
	key := widgetProblemKey(id)
	state := p
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
	s.pendingPage = &PendingPage{template: template, args: args}
}

func (s Session) ProcessPendingPage() {
	x := s.pendingPage
	s.pendingPage = nil
	if x == nil {
		return
	}
	args := x.args
	template := x.template

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

func extractKeyFromWidgetId(id string) string {
	var result string
	CheckArg(id != "")
	i := strings.LastIndexByte(id, ':')
	if i >= 0 {
		result = id[i+1:]

	} else {
		result = id
	}
	return result
}

// ------------------------------------------------------------------------------------
// Reading widget state values
// ------------------------------------------------------------------------------------

// Get the WidgetStateProvider for a widget, and the widget value's key (derived by trimming the prefix if appropriate)
func (s Session) getStateProvider(w Widget) (string, JSMap) {
	pr := PrIf("getStateProvider", false)

	// If widget's own state provider has no state, use the one on the stack

	id := w.Id()
	key := extractKeyFromWidgetId(id)

	pr("id:", QUO, id, "key:", QUO, key)

	// Use the widget's state, if one is defined; otherwise, stacked state
	p := w.StateProvider()
	if p == nil {
		pr("using stacked state provider")
		p = s.stackedStateProvider()
	}

	pr("returning key:", QUO, key, "and provider:", p)
	return key, p
}

// Read widget value; assumed to be a string.
func (s Session) WidgetStringValue(w Widget) string {
	id, p := s.getStateProvider(w)
	key := id
	return p.OptString(key, "")
}

// Read widget value; assumed to be an int.
func (s Session) WidgetIntValue(w Widget) int {
	id, p := s.getStateProvider(w)
	return p.OptInt(id, 0)
}

// Read widget value; assumed to be a bool.
func (s Session) WidgetBoolValue(w Widget) bool {
	id, p := s.getStateProvider(w)
	return p.OptBool(id, false)
}

func (s Session) SetWidgetValue(w Widget, value any) {
	pr := PrIf("SetWidgetValue", true)
	id, p := s.getStateProvider(w)
	pr("state provider, state:", p)
	oldVal := p.OptUnsafe(id)
	changed := value != oldVal
	pr("old:", oldVal, "new:", value, "changed:", changed)
	if changed {
		p.Put(id, value)
		pr("repainting", p)
		w.Repaint()
	}
}

// Get the context for the current listener.  For list items, this will be the list element id.
func (s Session) Context() any {
	return s.listenerContext
}

// This merges a couple of separate functions, to reduce the complexity.
func (s Session) rebuildAndDisplayNewPage(pageProvider func(Session) Page) {
	// Dispose of any existing widgets
	s.widgetMap = make(map[string]Widget)

	// Build a new page widget
	s.PageWidget = s.Id(WidgetIdPage).Open()
	s.Close()

	// Get the new page (it is now safe to construct, as the old widgets are gone)
	page := pageProvider(s)
	CheckState(page != nil, "no page was provided")
	s.debugPage = page

	// Display the new page
	Todo("!Verify that this works for normal refreshes as well as ajax operations")
	s.PageWidget.Repaint()

	s.browserURLExpr = "/" + page.Name() + "/" + strings.Join(page.Args(), "/")
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
			updatedValue, err := widget.LowListener()(s, widget, valAsString, nil)
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
