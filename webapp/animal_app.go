package webapp

import (
	. "github.com/jpsember/golang-base/app"
	. "github.com/jpsember/golang-base/base"
	"github.com/jpsember/golang-base/jimg"
	. "github.com/jpsember/golang-base/webapp/gen/webapp_data"
	. "github.com/jpsember/golang-base/webserv"
	"net/http"
)

var AutoActivateUser = Alert("?Automatically activating user")

type AnimalOperStruct struct {
	appRoot      Path
	headerMarkup string
	FullWidth    bool // If true, page occupies full width of screen
	TopPadding   int  // If nonzero, adds padding to top of page
	autoLoggedIn bool
	resources    Path
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

	s := NewJServer()
	//oper.server = s
	s.App = oper
	oper.appRoot = AscendToDirectoryContainingFileM("", "go.mod").JoinM("webserv")
	oper.resources = oper.appRoot.JoinM("resources")
	s.Resources = oper.resources
	s.SessionManager = BuildSessionMap()
	s.BaseURL = "jeff.org"
	s.KeyDir = oper.appRoot.JoinM("https_keys")

	dataSourcePath := ProjectDirM().JoinM("webapp/sqlite/animal_app_TMP_.db")

	if false && DevDatabase && Alert("Deleting database:", dataSourcePath) {
		DeleteDatabase(dataSourcePath)
	}
	CreateDatabase(dataSourcePath.String())
	oper.prepareDatabase()

	if DevDatabase {
		PopulateDatabase()
	}

	oper.headerMarkup = s.Resources.JoinM("header.html").ReadStringM()

	s.StartServing()
}

func (oper AnimalOper) PrepareSession(sess Session) {
	user := AssignUserToSession(sess)
	CheckState(user.Id() == 0)
	oper.constructPageWidget(sess)
	NewLandingPage(sess, sess.PageWidget).Generate()
}

func (oper AnimalOper) HandleRequest(s Session, path string) bool {
	pr := PrIf(true)

	pr("HandleRequest:", path)

	var text string
	var flag bool
	if text, flag = TrimIfPrefix(path, "/r/"); flag {
		pr("handling blob request with:", text)
		err := oper.handleBlobRequest(s, text)
		ReportIfError(err, "handling blob request")
		return true
	}

	if path == "/" {
		oper.debugAutoLogIn(s)
		oper.processFullPageRequest(s)
		return true
	}

	if _, found := TrimIfPrefix(path, "/manager"); found {
		NewManagerPage(s, s.PageWidget).Generate()
		oper.processFullPageRequest(s)
		return true
	}
	Todo("Experiment: checking for editing a particular animal")
	if remainder, found := TrimIfPrefix(path, "/edit/"); found {
		if animalId, err := ParseAsPositiveInt(remainder); err == nil {
			pr("generating page to edit animal #", animalId)
			NewEditAnimalPage(s, s.PageWidget, animalId).Generate()
			oper.processFullPageRequest(s)
			return true
		}
		return false
	}
	return false
}

func (oper AnimalOper) handleBlobRequest(s Session, blobId string) error {
	blob := SharedWebCache.GetBlobWithName(blobId)
	if blob.Id() == 0 {
		Alert("#50Can't find blob with name:", Quoted(blobId))
	}

	err := WriteResponse(s.ResponseWriter, InferContentTypeFromBlob(blob), blob.Data())
	Todo("?Detect someone requesting huge numbers of items that don't exist?")
	return err
}

func (oper AnimalOper) processFullPageRequest(sess Session) {
	sb := NewMarkupBuilder()
	oper.writeHeader(sb)
	CheckState(sess.PageWidget != nil, "no PageWidget!")
	RenderWidget(sess.PageWidget, sess, sb)
	sess.RequestClientInfo(sb)
	oper.writeFooter(sess.ResponseWriter, sb)
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

var alertWidget AlertWidget
var myRand = NewJSRand().SetSeed(1234)

const WidgetIdPage = "main_page"

// Assign a widget heirarchy to a session
func (oper AnimalOper) constructPageWidget(sess Session) {
	m := sess.WidgetManager()
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

// Perform a once-only attempt to log in the user automatically and set a particular page.
// For development only.
func (oper AnimalOper) debugAutoLogIn(sess Session) {
	if oper.autoLoggedIn {
		return
	}
	oper.autoLoggedIn = true
	Alert("Auto logging in")

	if false {
		NewGalleryPage(sess, sess.PageWidget).Generate()
		return
	}
	user2, _ := ReadUserWithName("manager1")
	if user2.Id() == 0 {
		return
	}
	if !TryLoggingIn(sess, user2) {
		return
	}

	if true {
		NewAnimalFeedPage(sess, sess.PageWidget).Generate()
		return
	}
	if false {
		NewCreateAnimalPage(sess, sess.PageWidget).Generate()
		return
	}

	NewManagerPage(sess, sess.PageWidget).Generate()
}
