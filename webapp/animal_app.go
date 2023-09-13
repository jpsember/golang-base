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
			Todo("!This should be moved to the webserv package, maybe if an initialization parameter was specified")
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
	pr := PrIf(true)
	pr("handler, request:", req.RequestURI)

	// We don't know what the session is yet, so we don't have a lock on it...
	sess := DetermineSession(oper.sessionManager, w, req, true)

	Todo("This shouldn't be done until we have a lock on the session; maybe lock the session throughout the handler?  Or do we already have it?")

	// Now that we have the session, lock it
	Todo("But when we kill the session, i.e. logging out, do we still have the lock?")
	sess.Mutex2.Lock()
	defer sess.Mutex2.Unlock()

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

				if true {
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

		Todo("!Move as much of this as possible to the webserv package")
		path := url.Path
		var text string
		var flag bool

		pr("url path:", path)
		if path == "/ajax" {
			Todo("!Use TrimIfPrefix here as well")
			sess.HandleAjaxRequest(w, req)
		} else if text, flag = TrimIfPrefix(path, "/r/"); flag {
			pr("handling blob request with:", text)
			err = oper.handleBlobRequest(w, req, text)
		} else if text, flag = TrimIfPrefix(path, `/upload/`); flag {
			pr("handling upload request with:", text)
			sess.HandleUploadRequest(w, req, text)
		} else {
			result := oper.animalURLRequestHandler(w, req, sess, path)
			if !result {
				// If we fail to parse any requests, assume it's a resource, like that stupid favicon
				pr("handling resource request for:", path)
				err = sess.HandleResourceRequest(w, req, oper.resources)
			}
		}
	}

	if err != nil {
		sess.SetRequestProblem(err)
	}

	Todo("This code should be done while the lock is still held")
	if p := sess.GetRequestProblem(); p != nil {
		Pr("...problem with request, URL:", req.RequestURI, INDENT, p)
	}
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
	sb := NewMarkupBuilder()
	oper.writeHeader(sb)
	CheckState(sess.PageWidget != nil, "no PageWidget!")
	RenderWidget(sess.PageWidget, sess, sb)
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

func (oper AnimalOper) AcquireLockAndCallURLRequestHandler(sess Session, path string) {

}

const (
	SessionKey_User     = "user"
	SessionKey_FeedList = "feed.list"
	SessionKey_MgrList  = "mgr.list"
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

// This is our handler for serving up entire pages, as opposed to AJAX requests.
func (oper AnimalOper) animalURLRequestHandler(w http.ResponseWriter, req *http.Request, s Session, expr string) bool {
	pr := PrIf(true)
	pr("animalURLRequestHandler:", expr)

	if expr == "/" {
		oper.processFullPageRequest(s, w, req)
		return true
	}

	if _, found := TrimIfPrefix(expr, "/manager"); found {
		NewManagerPage(s, s.PageWidget).Generate()
		oper.processFullPageRequest(s, w, req)
		return true
	}
	Todo("Experiment: checking for editing a particular animal")
	if remainder, found := TrimIfPrefix(expr, "/edit/"); found {
		if animalId, err := ParseAsPositiveInt(remainder); err == nil {
			pr("generating page to edit animal #", animalId)
			NewEditAnimalPage(s, s.PageWidget, animalId).Generate()
			oper.processFullPageRequest(s, w, req)

			//	oper.ExperimentSendPageToClient(s, w)
			return true
		}
		return false
	}
	return false
}
