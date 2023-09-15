package webapp

import (
	. "github.com/jpsember/golang-base/app"
	. "github.com/jpsember/golang-base/base"
	"github.com/jpsember/golang-base/jimg"
	. "github.com/jpsember/golang-base/webapp/gen/webapp_data"
	. "github.com/jpsember/golang-base/webserv"
)

type AnimalOperStruct struct {
	appRoot      Path
	FullWidth    bool // If true, page occupies full width of screen
	TopPadding   int  // If nonzero, adds padding to top of page
	autoLoggedIn bool
	resources    Path
	jserver      JServer
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
	oper.prepareDatabase()

	DevLabelRenderer = AddDevPageLabel
	// Initialize and start the JServer
	//
	{
		s := NewJServer(oper)
		oper.jserver = s
		s.SessionManager = BuildSessionMap()
		s.BaseURL = "jeff.org"
		s.KeyDir = oper.appRoot.JoinM("https_keys")
		preq := s.PgRequester
		preq.RegisterPages(LandingPageTemplate, GalleryPageTemplate, NewSignUpPage(nil), FeedPageTemplate, ManagerPageTemplate,
			ViewAnimalPageTemplate, CreateAnimalPageTemplate, EditAnimalPageTemplate)
		s.AddResourceHandler(BlobURLPrefix, oper.handleBlobRequest)
		s.StartServing()
	}

}

func (oper AnimalOper) Resources() Path {
	return oper.resources
}
func (oper AnimalOper) UserForSession(s Session) AbstractUser {
	return OptSessionUser(s)
}

func (oper AnimalOper) DefaultPageForUser(abstractUser AbstractUser) Page {
	user := abstractUser.(User)
	userId := 0
	if user != nil {
		userId = user.Id()
	}
	var result Page
	if userId == 0 || !IsUserLoggedIn(user.Id()) {
		result = LandingPageTemplate
	} else {
		switch user.UserClass() {
		case UserClassDonor:
			result = FeedPageTemplate
		case UserClassManager:
			result = ManagerPageTemplate
		default:
			NotSupported("page for", user.UserClass())
		}
	}
	return result
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

func (oper AnimalOper) handleBlobRequest(s Session, blobId string) {
	blob := SharedWebCache.GetBlobWithName(blobId)
	if blob.Id() == 0 {
		Alert("#50Can't find blob with name:", Quoted(blobId))
	}
	err := WriteResponse(s.ResponseWriter, InferContentTypeFromBlob(blob), blob.Data())
	Todo("?Detect someone requesting huge numbers of items that don't exist?")
	ReportIfError(err, "Trouble writing blob response")
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

func (oper AnimalOper) registerPages(r PageRequester) {
	r.RegisterPages(LandingPageTemplate, GalleryPageTemplate, NewSignUpPage(nil), FeedPageTemplate, ManagerPageTemplate,
		ViewAnimalPageTemplate, CreateAnimalPageTemplate, EditAnimalPageTemplate)
}
