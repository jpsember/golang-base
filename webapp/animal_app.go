package webapp

import (
	"bytes"
	. "github.com/jpsember/golang-base/app"
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/webapp/gen/webapp_data"
	. "github.com/jpsember/golang-base/webserv"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

var AutoActivateUser = Alert("?Automatically activating user")

type AnimalOperStruct struct {
	sessionManager SessionManager
	appRoot        Path
	resources      Path
	headerMarkup   string
	FullWidth      bool // If true, page occupies full width of screen
	TopPadding     int  // If nonzero, adds padding to top of page
}

type AnimalOper = *AnimalOperStruct

func (oper AnimalOper) UserCommand() string {
	return "widgets"
}

func (oper AnimalOper) GetHelp(bp *BasePrinter) {
	bp.Pr("Demonstrates a web server with AJAX manipulating Widget UI elements")
}

func (oper AnimalOper) ProcessArgs(c *CmdLineArgs) {
}

var DevDatabase = Alert("!Using development database")

// If DevDatabase is active, and user with this name exists, their credentials are plugged in automatically
// at the sign in page by default.
const AutoSignInName = "manager1"

func (oper AnimalOper) Perform(app *App) {
	//ClearAlertHistory()
	ExitOnPanic()

	dataSourcePath := ProjectDirM().JoinM("webapp/sqlite/animal_app_TMP_.db")

	if false && DevDatabase && Alert("Deleting database:", dataSourcePath) {
		DeleteDatabase(dataSourcePath)
	}
	CreateDatabase(dataSourcePath.String())
	if DevDatabase {
		PopulateDatabase()
	}

	oper.sessionManager = BuildSessionMap()
	oper.appRoot = AscendToDirectoryContainingFileM("", "go.mod").JoinM("webserv")
	oper.resources = oper.appRoot.JoinM("resources")

	{
		s := strings.Builder{}
		s.WriteString(oper.resources.JoinM("header.html").ReadStringM())
		oper.headerMarkup = s.String()
	}

	var ourUrl = "jeff.org"

	var keyDir = oper.appRoot.JoinM("https_keys")
	var certPath = keyDir.JoinM(ourUrl + ".crt")
	var keyPath = keyDir.JoinM(ourUrl + ".key")
	Pr("URL:", INDENT, `https://`+ourUrl)

	http.HandleFunc("/",
		func(w http.ResponseWriter, req *http.Request) {
			defer func() {
				if r := recover(); r != nil {
					BadState("<1Panic during http.HandleFunc:", r)
				}
			}()
			oper.handle(w, req)
		})

	err := http.ListenAndServeTLS(":443", certPath.String(), keyPath.String(), nil)

	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

// A handler such as this must be thread safe!
func (oper AnimalOper) handle(w http.ResponseWriter, req *http.Request) {
	pr := PrIf(true)
	pr("handler, request:", req.RequestURI)

	if false && Alert("!If full page requested, discarding sessions") {
		url, err := url.Parse(req.RequestURI)
		if err == nil && url.Path == "/" {
			sess := DetermineSession(oper.sessionManager, w, req, false)
			if sess != nil {
				DiscardAllSessions(oper.sessionManager)
			}
		}
	}

	sess := DetermineSession(oper.sessionManager, w, req, true)
	if sess.AppData == nil {
		AssignUserToSession(sess)
		oper.constructPageWidget(sess)

		user, ok := sess.AppData.(User)
		CheckState(ok, "no User found in sess AppData:", INDENT, sess.AppData)
		CheckState(
			user.Id() == 0)
		NewLandingPage(sess, sess.PageWidget).Generate()
	}

	url, err := url.Parse(req.RequestURI)
	if err == nil {

		path := url.Path
		var text string
		var flag bool

		pr("url path:", path)
		if path == "/ajax" {
			sess.HandleAjaxRequest(w, req)
		} else if path == "/" {
			oper.processFullPageRequest(sess, w, req)
		} else if text, flag = extractPrefix(path, "/r/"); flag {
			err = HandleBlobRequest(w, req, text)
		} else if text, flag = extractPrefix(path, `/upload/`); flag {
			Pr("handling upload request with:", text)
			err = HandleUploadRequest(sess, w, req, text)
		} else {
			pr("handling resource request for:", path)
			err = sess.HandleResourceRequest(w, req, oper.resources)
		}
	}

	if err != nil {
		sess.SetRequestProblem(err)
	}

	if p := sess.GetRequestProblem(); p != "" {
		Pr("...problem with request, URL:", req.RequestURI, INDENT, p)
	}
}

func extractPrefix(text string, prefix string) (string, bool) {
	if strings.HasPrefix(text, prefix) {
		return text[len(prefix):], true
	}
	return text, false
}

func HandleBlobRequest(w http.ResponseWriter, req *http.Request, blobId string) error {
	blob := SharedWebCache.GetBlobWithName(blobId)
	if blob.Id() == 0 {
		Alert("#50Can't find blob with name:", Quoted(blobId))
	}

	err := WriteResponse(w, InferContentTypeFromBlob(blob), blob.Data())
	Todo("?Detect someone requesting huge numbers of items that don't exist?")
	return err
}

func HandleUploadRequest(sess Session, w http.ResponseWriter, req *http.Request, widgetId string) error {
	Todo("!Must ensure thread safety while working with the user session")

	if req.Method != "POST" {
		return Error("upload request was not POST")
	}
	widget := sess.WidgetManager().Opt(widgetId)
	if widget == nil {
		return Error("handling upload request, can't find widget:", widgetId)
	}
	fileUploadWidget, ok := widget.(FileUpload)
	if !ok {
		return Error("handling upload request, widget isn't expected type:", widgetId)
	}

	// From https://freshman.tech/file-upload-golang/
	const MAX_UPLOAD_SIZE = 10_000_000
	req.Body = http.MaxBytesReader(w, req.Body, MAX_UPLOAD_SIZE)
	if err := req.ParseMultipartForm(MAX_UPLOAD_SIZE); err != nil {
		return Error("The uploaded file is too big. Please choose an file that's less than 10MB in size")
	}

	// The argument to FormFile must match the name attribute
	// of the file input on the frontend

	Pr("request FormFile:", req.MultipartForm.File)
	Todo("this multipart map is empty.  Am I using a multi upload when I should be using single?")
	
	file, fileHeader, err := req.FormFile("file")
	if err != nil {
		return Error("trouble getting request FormFile:", err)
	}
	Todo("do something with fileHeader?", fileHeader)

	defer file.Close()

	//// Create the uploads folder if it doesn't
	//// already exist
	//err = os.MkdirAll("./uploads", os.ModePerm)
	//if err != nil {
	//	http.Error(w, err.Error(), http.StatusInternalServerError)
	//	return
	//}
	//
	//// Create a new file in the uploads directory
	//dst, err := os.Create(fmt.Sprintf("./uploads/%d%s", time.Now().UnixNano(), filepath.Ext(fileHeader.Filename)))
	//if err != nil {
	//	http.Error(w, err.Error(), http.StatusInternalServerError)
	//	return
	//}
	//
	//defer dst.Close()

	var buf bytes.Buffer
	foo := io.Writer(&buf)
	length, err1 := io.Copy(foo, file)
	if err1 != nil {
		return Error("failed to read uploaded file into byte array:", err1)
	}
	Pr("bytes buffer length:", len(buf.Bytes()), "read:", length)

	result := buf.Bytes()[0:length]
	Todo("do something with result", result, "and file upload widget", fileUploadWidget)
	return nil
}

func (oper AnimalOper) processFullPageRequest(sess Session, w http.ResponseWriter, req *http.Request) {
	// Construct a session if none found, and a widget for a full webpage
	//sess := DetermineSession(oper.sessionManager, w, req, true)
	sess.Mutex.Lock()
	defer sess.Mutex.Unlock()

	sb := NewMarkupBuilder()
	oper.writeHeader(sb)
	CheckState(sess.PageWidget != nil, "no PageWidget!")
	sess.PageWidget.RenderTo(sb, sess.State)
	sess.RequestClientInfo(sb)
	oper.writeFooter(w, sb)
}

// Generate the biolerplate header and scripts markup
func (oper AnimalOper) writeHeader(bp MarkupBuilder) {
	bp.A(oper.headerMarkup)
	bp.OpenTag("body")
	containerClass := "container"
	if oper.FullWidth {
		containerClass = "container-fluid"
	}
	if oper.TopPadding != 0 {
		containerClass += "  pt-" + IntToString(oper.TopPadding)
	}
	bp.Comments("page container").OpenTag(`div class='` + containerClass + `'`)
}

// Generate the boilerplate footer markup, then write the page to the response
func (oper AnimalOper) writeFooter(w http.ResponseWriter, bp MarkupBuilder) {
	bp.CloseTag() // page container
	bp.CloseTag() // body
	bp.A(`</html>`).Cr()
	WriteResponse(w, "text/html", bp.Bytes())
}

const WidgetIdPage = "main_page"

var alertWidget AlertWidget
var myRand = NewJSRand().SetSeed(1234)

// Assign a widget heirarchy to a session
func (oper AnimalOper) constructPageWidget(sess Session) {
	m := sess.WidgetManager()
	//m.AlertVerbose()

	Todo("?Clarify when we need to *remove* old widgets")
	m.Id(WidgetIdPage)
	widget := m.Open()
	sess.PageWidget = widget
	m.Close()
}

// A new session was created; assign an 'unknown' user to it
func AssignUserToSession(sess Session) {
	sess.AppData = NewUser().Build()
}

func SessionUser(sess Session) User {
	user := OptSessionUser(sess)
	if user.Id() == 0 {
		BadState("session user has id zero")
	}
	return user
}

func OptSessionUser(sess Session) User {
	user, ok := sess.AppData.(User)
	if !ok {
		BadState("no User found in sess AppData:", INDENT, sess.AppData)
	}
	return user
}
