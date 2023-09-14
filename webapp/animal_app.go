package webapp

import (
	. "github.com/jpsember/golang-base/app"
	. "github.com/jpsember/golang-base/base"
	"github.com/jpsember/golang-base/jimg"
	. "github.com/jpsember/golang-base/webapp/gen/webapp_data"
	. "github.com/jpsember/golang-base/webserv"
)

type AnimalOperStruct struct {
	appRoot       Path
	headerMarkup  string
	FullWidth     bool // If true, page occupies full width of screen
	TopPadding    int  // If nonzero, adds padding to top of page
	autoLoggedIn  bool
	resources     Path
	pageRequester PageRequester
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

func (oper AnimalOper) Perform(app *App) {
	//ClearAlertHistory()
	ExitOnPanic()

	oper.appRoot = AscendToDirectoryContainingFileM("", "go.mod").JoinM("webserv")
	oper.resources = oper.appRoot.JoinM("resources")
	oper.headerMarkup = oper.resources.JoinM("header.html").ReadStringM()
	oper.prepareDatabase()

	oper.pageRequester = NewPageRequester()
	oper.registerPages(oper.pageRequester)

	// Initialize and start the JServer
	//
	{
		s := NewJServer()
		s.App = oper
		s.Resources = oper.resources
		s.SessionManager = BuildSessionMap()
		s.BaseURL = "jeff.org"
		s.KeyDir = oper.appRoot.JoinM("https_keys")
		s.StartServing()
	}
}

// JServer callback to perform initialization for a new session.  We assign a user,
// and open the landing page. It might get replaced by another page immediately...?
func (oper AnimalOper) PrepareSession(sess Session) {
	user := DefaultUser
	sess.PutSessionData(SessionKey_User, user)
	CheckState(user.Id() == 0)
	NewLandingPage(sess)
}

// JServer callback to handle a request.  Returns true if it was handled.
func (oper AnimalOper) HandleRequest(s Session, path string) bool {
	pr := PrIf(true)

	var text string
	var flag bool
	if text, flag = TrimIfPrefix(path, "/r/"); flag {
		//pr("handling blob request with:", text)
		oper.handleBlobRequest(s, text)
		return true
	}
	pr("AnimalOper, HandleRequest:", path)

	return oper.processPageRequest(s, path)

}

func (oper AnimalOper) handleBlobRequest(s Session, blobId string) {
	blob := SharedWebCache.GetBlobWithName(blobId)
	if blob.Id() == 0 {
		Alert("#50Can't find blob with name:", Quoted(blobId))
	}
	err := WriteResponse(s.ResponseWriter, InferContentTypeFromBlob(blob), blob.Data())
	Todo("?Detect someone requesting huge numbers of items that don't exist?")
	ReportIfError(err, "Trouble writing blob response")
}

func (oper AnimalOper) renderPage(sess Session) {
	CheckState(sess.PageWidget != nil, "no PageWidget!")
	sb := NewMarkupBuilder()
	oper.writeHeader(sb)
	RenderWidget(sess.PageWidget, sess, sb)
	sess.RequestClientInfo(sb)
	oper.writeFooter(sess, sb)
	WriteResponse(sess.ResponseWriter, "text/html", sb.Bytes())
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
func (oper AnimalOper) writeFooter(s Session, bp MarkupBuilder) {
	bp.CloseTag() // page container

	// Add a bit of javascript that will change the url to what we want
	if s.PendingURLExpr != "" {
		Pr("appending URL expr:", s.PendingURLExpr)
		code := `
<script type="text/javascript">
history.pushState(null, null, location.origin+'` + s.PendingURLExpr + `')
</script>
`
		Pr("Appending code to end of <body>:", VERT_SP, code, VERT_SP)

		bp.WriteString(code)
	}
	//history.pushState("object or string representing the state of the page", "new title", "newURL")
	bp.CloseTag() // body

	bp.A(`</html>`).Cr()

}

func (oper AnimalOper) prepareDatabase() {
	dataSourcePath := ProjectDirM().JoinM("webapp/sqlite/animal_app_TMP_.db")

	if false && DevDatabase && Alert("Deleting database:", dataSourcePath) {
		DeleteDatabase(dataSourcePath)
	}
	CreateDatabase(dataSourcePath.String())

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

	if DevDatabase {
		PopulateDatabase()
	}
}

// ------------------------------------------------------------------------------------
// Data stored with session
// ------------------------------------------------------------------------------------

const (
	SessionKey_User     = "user"
	SessionKey_FeedList = "feed.list"
	SessionKey_MgrList  = "mgr.list"
)

// Get session's User, or nil if there isn't one.
func OptSessionUser(sess Session) User {
	return sess.GetSessionData(SessionKey_User).(User)
}

// Get session's User.
func SessionUser(sess Session) User {
	user := OptSessionUser(sess)
	if user.Id() == 0 {
		BadState("session user has id zero")
	}
	return user
}

// Attempt to make the user logged in.  Return true if successful.
func TryLoggingIn(s Session, user User) bool {
	success := TryRegisteringUserAsLoggedIn(user.Id(), true)
	if success {
		s.PutSessionData(SessionKey_User, user)
	}
	return success
}

// Perform a once-only attempt to log in the user automatically and set a particular page.
// For development only.
func (oper AnimalOper) debugAutoLogIn(sess Session) {
	if oper.autoLoggedIn {
		return
	}
	oper.autoLoggedIn = true
	Alert("Auto logging in")

	//if false {
	//	NewGalleryPage(sess, sess.PageWidget).Generate()
	//	return
	//}
	user2, _ := ReadUserWithName("manager1")
	Pr("read user:", user2)
	if user2.Id() == 0 {
		return
	}
	if !TryLoggingIn(sess, user2) {
		return
	}
	//
	if true {
		NewFeedPage(sess).Generate()
		return
	}
}

// ------------------------------------------------------------------------------------
// User state and current page
// ------------------------------------------------------------------------------------

// Parse URL requested by client, and serve up an appropriate page.
func (oper AnimalOper) processPageRequest(s Session, path string) bool {
	pr := PrIf(true)

	if true {
		page := oper.pageRequester.Process(s, path)
		bp := page.GetBasicPage()
		s.SetURLExpression(bp.PageName)
		Todo("Maybe the Generate function should be in the abstract Page type?")
		bp.Generate()
		oper.renderPage(s)
		return true
	}

	if path == "/" {
		oper.debugAutoLogIn(s)
		oper.renderPage(s)
		s.SetURLExpression("what", "the", "heck")
		return true
	}

	if _, found := TrimIfPrefix(path, "/manager"); found {
		NewManagerPage(s).Generate()
		oper.renderPage(s)
		pr("rendered manager page")
		return true
	}
	Todo("Experiment: checking for editing a particular animal")
	if remainder, found := TrimIfPrefix(path, "/edit/"); found {
		if animalId, err := ParseAsPositiveInt(remainder); err == nil {
			pr("generating page to edit animal #", animalId)
			page := NewEditAnimalPage(s, animalId)
			page.Generate()
			oper.renderPage(s)
			return true
		}
		return false
	}
	return false
}

func (oper AnimalOper) registerPages(r PageRequester) {

	r.RegisterPage(LandingPageTemplate)
	r.RegisterPage(FeedPageTemplate)
	r.RegisterPage(GalleryPageTemplate)

	//
	//
	//
	//Pr("landing page:",x.PageName,x.animalId)
	//y = &x.BasicPageStruct
	//Pr("pointer to basic page struct:",y)
	//
	//
	//q := &x.BasicPageStruct
	//r := q.(LandingPage)
	//
	////z = (y.(*LandingPageStruct))
	//z = &(y.(LandingPage))
	//Pr("
	//
	//r.Register(NewLandingPage(nil))
}
