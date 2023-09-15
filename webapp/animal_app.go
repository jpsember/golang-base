package webapp

import (
	. "github.com/jpsember/golang-base/app"
	. "github.com/jpsember/golang-base/base"
	"github.com/jpsember/golang-base/jimg"
	. "github.com/jpsember/golang-base/webapp/gen/webapp_data"
	. "github.com/jpsember/golang-base/webserv"
	"regexp"
	"strings"
)

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
var SharedPageRequester PageRequester

func (oper AnimalOper) Perform(app *App) {
	//ClearAlertHistory()
	ExitOnPanic()

	oper.appRoot = AscendToDirectoryContainingFileM("", "go.mod").JoinM("webserv")
	oper.resources = oper.appRoot.JoinM("resources")
	oper.headerMarkup = oper.resources.JoinM("header.html").ReadStringM()
	oper.prepareDatabase()

	SharedPageRequester = NewPageRequester()
	Todo("!Emphasize that PageRequester must be threadsafe")
	oper.registerPages(SharedPageRequester)

	DevLabelRenderer = AddDevPageLabel
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

func devLabelRenderer(s Session, p Page) {
}

// JServer callback to perform initialization for a new session.  We assign a user,
// and open the landing page. It might get replaced by another page immediately...?
func (oper AnimalOper) PrepareSession(sess Session) {
	user := DefaultUser
	sess.PutSessionData(SessionKey_User, user)
	CheckState(user.Id() == 0)

	if Alert("!Doing auto login") {
		oper.debugAutoLogIn(sess)
	}
}

// JServer callback to handle a request.  Returns true if it was handled.
func (oper AnimalOper) HandleRequest(s Session, path string) bool {
	pr := PrIf(false)

	if s == nil {
		Alert("#50HandleRequest, but session is nil for:", path)
		return false
	}
	var text string
	var flag bool
	if text, flag = TrimIfPrefix(path, BlobURLPrefix); flag {
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

func (oper AnimalOper) sendFullPage(sess Session) {
	CheckState(sess.PageWidget != nil, "no PageWidget!")
	sb := NewMarkupBuilder()
	oper.writeHeader(sb)
	RenderWidget(sess.PageWidget, sess, sb)
	sess.RequestClientInfo(sb)
	oper.writeFooter(sess, sb)
	WriteResponse(sess.ResponseWriter, "text/html", sb.Bytes())
}

// Generate the boilerplate header and scripts markup
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

	Todo("move this to webserv module")

	// Add a bit of javascript that will change the url to what we want
	expr := s.NewBrowserPath()
	if expr != "" {
		code := `
<script type="text/javascript">
var url = location.origin+'` + expr + `'
history.replaceState(null, null, url)
</script>
`
		// ^^^I suspect we don't want to do pushState if we got here due to user pressing the back button.
		bp.WriteString(code)
	}
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

// Get session's User, or default user if there isn't one.
func OptSessionUser(sess Session) User {
	u := sess.GetSessionData(SessionKey_User).(User)
	if u == nil {
		u = DefaultUser
	}
	return u
}

func SessionUserIs(sess Session, class UserClass) bool {
	user := OptSessionUser(sess)
	return user.UserClass() == class
}

func SessionDefaultPage(sess Session) Page {
	user := OptSessionUser(sess)
	return SharedPageRequester.DefaultPagePage(user)
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

	user2, _ := ReadUserWithName("manager1")
	Alert("?Auto logging in", user2.Id(), user2.Name())
	if user2.Id() == 0 {
		return
	}
	if !TryLoggingIn(sess, user2) {
		return
	}
	if true {
		sess.SwitchToPage(NewFeedPage(sess))
		return
	}
}

// ------------------------------------------------------------------------------------
// User state and current page
// ------------------------------------------------------------------------------------

// Strings that start with {zero or more lowercase letters}/
var looksLikePageRexEx = CheckOkWith(regexp.Compile(`^[a-z]*\/`))

// Parse URL requested by client, and serve up an appropriate page.
func (oper AnimalOper) processPageRequest(s Session, path string) bool {

	// If path is NOT one of "", "pagename", or "pagename[non-alpha]..., exit with false immediately
	{
		// Add a trailing / to make this logic simpler
		modifiedPath := path + `/`
		if !looksLikePageRexEx.MatchString(modifiedPath) {
			return false
		}
		// Extract page name
		i := strings.IndexByte(modifiedPath, '/')
		pageName := modifiedPath[0:i]

		// If what remains is a nonempty string that isn't the name of a page, exit
		if pageName != "" && SharedPageRequester.PageWithName(pageName) == nil {
			return false
		}
	}

	page := SharedPageRequester.Process(s, path)
	if page != nil {
		s.SwitchToPage(page)
		oper.sendFullPage(s)
		return true
	}

	return false
}

func (oper AnimalOper) registerPages(r PageRequester) {
	r.RegisterPages(LandingPageTemplate, GalleryPageTemplate, SignUpPageTemplate, FeedPageTemplate, ManagerPageTemplate,
		ViewAnimalPageTemplate, CreateAnimalPageTemplate, EditAnimalPageTemplate)
}
