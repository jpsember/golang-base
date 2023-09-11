package webapp

import (
	. "github.com/jpsember/golang-base/app"
	. "github.com/jpsember/golang-base/base"
	"github.com/jpsember/golang-base/jimg"
	. "github.com/jpsember/golang-base/webapp/gen/webapp_data"
	. "github.com/jpsember/golang-base/webserv"
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

	oper.sessionManager = BuildSessionMap()
	oper.appRoot = AscendToDirectoryContainingFileM("", "go.mod").JoinM("webserv")
	oper.resources = oper.appRoot.JoinM("resources")

	dataSourcePath := ProjectDirM().JoinM("webapp/sqlite/animal_app_TMP_.db")

	if false && DevDatabase && Alert("Deleting database:", dataSourcePath) {
		DeleteDatabase(dataSourcePath)
	}
	CreateDatabase(dataSourcePath.String())
	oper.prepareDatabase()

	if DevDatabase {
		PopulateDatabase()
	}

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

var jumped bool

// A handler such as this must be thread safe!
func (oper AnimalOper) handle(w http.ResponseWriter, req *http.Request) {
	pr := PrIf(false)
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
	optUser := sess.OptSessionData(SessionKey_User)
	if optUser == nil {
		user := AssignUserToSession(sess)
		CheckState(user.Id() == 0)
		oper.constructPageWidget(sess)

		NewLandingPage(sess, sess.PageWidget).Generate()

		if !jumped && true && Alert("Jumping to different page") {
			jumped = true
			for {
				if false {
					NewGalleryPage(sess, sess.PageWidget).Generate()
					break
				}
				user2, _ := ReadUserWithName("manager1")
				if user2.Id() == 0 {
					break
				}
				if !TryLoggingIn(sess, user2) {
					break
				}

				if false {
					NewAnimalFeedPage(sess, sess.PageWidget).Generate()
					break
				}
				if false {
					NewCreateAnimalPage(sess, sess.PageWidget).Generate()
					break
				}

				NewManagerPage(sess, sess.PageWidget).Generate()
				break
			}
		}
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
			pr("handling blob request with:", text)
			err = oper.handleBlobRequest(w, req, text)
		} else if text, flag = extractPrefix(path, `/upload/`); flag {
			pr("handling upload request with:", text)
			sess.HandleUploadRequest(w, req, text)
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

func (oper AnimalOper) handleBlobRequest(w http.ResponseWriter, req *http.Request, blobId string) error {
	blob := SharedWebCache.GetBlobWithName(blobId)
	if blob.Id() == 0 {
		Alert("#50Can't find blob with name:", Quoted(blobId))
	}

	err := WriteResponse(w, InferContentTypeFromBlob(blob), blob.Data())
	Todo("?Detect someone requesting huge numbers of items that don't exist?")
	return err
}

func (oper AnimalOper) processFullPageRequest(sess Session, w http.ResponseWriter, req *http.Request) {
	// Construct a session if none found, and a widget for a full webpage
	//sess := DetermineSession(oper.sessionManager, w, req, true)
	sess.Mutex.Lock()
	defer sess.Mutex.Unlock()

	sb := NewMarkupBuilder()
	oper.writeHeader(sb)
	CheckState(sess.PageWidget != nil, "no PageWidget!")
	sess.PageWidget.RenderTo(sess, sb)
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
func AssignUserToSession(sess Session) User {
	user := DefaultUser
	sess.PutSessionData(SessionKey_User, user)
	return user
}

func (oper AnimalOper) prepareDatabase() {
	if b, _ := ReadBlob(1); b.Id() == 0 {
		// Generate default images as blobs
		oper.resources = oper.appRoot.JoinM("resources")

		animalPicPlaceholderPath := oper.resources.JoinM("placeholder.jpg")
		img := CheckOkWith(jimg.DecodeImage(animalPicPlaceholderPath.ReadBytesM()))
		img = img.ScaleToSize(AnimalPicSizeNormal)
		jpeg := CheckOkWith(img.ToJPEG())
		Todo("?Later, keep the original image around for crop adjustments; but for now, scale and store immediately")
		b := NewBlob()
		b.SetData(jpeg)
		AssignBlobName(b)
		created, err := CreateBlob(b)
		CheckOk(err)
		CheckState(created.Id() == 1, "unexpected id for placeholder:", created.Id())
	}
}

const (
	SessionKey_User    = "user"
	SessionKey_MgrList = "mgr.list"
)

func TryLoggingIn(s Session, user User) bool {
	success := false
	if TryRegisteringUserAsLoggedIn(user.Id(), true) {
		success = true
		s.PutSessionData(SessionKey_User, user)
	}
	return success
}

func SessionUser(sess Session) User {
	user := OptSessionUser(sess)
	if user.Id() == 0 {
		BadState("session user has id zero")
	}
	return user
}

func OptSessionUser(sess Session) User {
	return sess.GetSessionData(SessionKey_User).(User)
}
