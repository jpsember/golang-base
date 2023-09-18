package webapp

import (
	. "github.com/jpsember/golang-base/app"
	. "github.com/jpsember/golang-base/base"
	"github.com/jpsember/golang-base/jimg"
	. "github.com/jpsember/golang-base/webapp/gen/webapp_data"
	. "github.com/jpsember/golang-base/webserv"
)

const AutoLogInName = "manager1"

var DevDatabase = Alert("!Using development database")

type AnimalOperStruct struct {
	appRoot      Path
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

func (oper AnimalOper) Perform(app *App) {
	ClearAlertHistory()
	ExitOnPanic()

	oper.appRoot = AscendToDirectoryContainingFileM("", "go.mod").JoinM("webserv")
	oper.resources = oper.appRoot.JoinM("resources")
	oper.prepareDatabase()

	DebugUIFlag = true

	s := NewJServer(oper)
	s.SessionManager = BuildSessionMap()
	s.BaseURL = "jeff.org"
	s.KeyDir = oper.appRoot.JoinM("https_keys")
	SharedWebCache = ConstructSharedWebCache()
	s.BlobCache = SharedWebCache
	s.StartServing()
}

// ------------------------------------------------------------------------------------
// ServerApp interface
// ------------------------------------------------------------------------------------

func (oper AnimalOper) PageTemplates() []Page {
	return []Page{
		LandingPageTemplate, GalleryPageTemplate, NewSignUpPage(nil), FeedPageTemplate, ManagerPageTemplate,
		ViewAnimalPageTemplate, CreateAnimalPageTemplate, EditAnimalPageTemplate,
	}
}

func (oper AnimalOper) Resources() Path {
	return oper.resources
}
func (oper AnimalOper) UserForSession(s Session) AbstractUser {
	return OptSessionUser(s)
}

func (oper AnimalOper) DefaultPageForUser(abstractUser AbstractUser) Page {
	if true && Alert("gallery") {
		return GalleryPageTemplate
	}
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

// JServer callback to perform any optional additional initialization for a new session.
func (oper AnimalOper) PrepareSession(sess Session) {

	// Perform a once-only attempt to do an auto login
	for {
		nm := AutoLogInName
		if nm == "" {
			break
		}
		Todo("!Auto logging in:", nm)
		if oper.autoLoggedIn {
			break
		}
		oper.autoLoggedIn = true

		user2, _ := ReadUserWithName(nm)
		if user2.Id() == 0 {
			Alert("Can't find auto login user:", nm)
			break
		}
		if !TryLoggingIn(sess, user2) {
			break
		}

		break
	}
}

// ------------------------------------------------------------------------------------

func (oper AnimalOper) prepareDatabase() {
	dataSourcePath := ProjectDirM().JoinM("webapp/sqlite/animal_app_TMP_.db")

	if false && DevDatabase && Alert("Deleting database:", dataSourcePath) {
		DeleteDatabase(dataSourcePath)
	}
	CreateDatabase(dataSourcePath.String())

	if b, _ := ReadBlob(1); b.Id() == 0 {

		// Generate default images as blobs
		animalPicPlaceholderPath := oper.Resources().JoinM("placeholder.jpg")
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
	SessionKey_User = "user"
)

// Get session's User, or default user if there isn't one.
func OptSessionUser(sess Session) User {
	u := DefaultUser
	data := sess.OptSessionData(SessionKey_User)
	if data != nil {
		u = data.(User)
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

// Attempt to log the user out. Return true if successful.
func LogOut(s Session) bool {
	user := SessionUser(s)
	if user.Id() == 0 {
		Alert("#50Attempt to log out user that is not logged in:", INDENT, user)
		return false
	}
	wasLoggedIn := LogUserOut(user.Id())
	if !wasLoggedIn {
		Alert("#50LogUserOut returned false:", INDENT, user)
	}
	s.PutSessionData(SessionKey_User, nil)
	return true
}
