package webapp

import (
	. "github.com/jpsember/golang-base/app"
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/webapp/gen/webapp_data"
	. "github.com/jpsember/golang-base/webserv"
	"strings"
)

const AutoLogInName = "manager1"

var DevAutoLogIn = true && Alert("!Performing auto log in with:", AutoLogInName)
var DevDatabase = Alert("!Using development database")
var DevGallery = false && Alert("Showing gallery")
var AllowTestInputs = DevDatabase && false && Alert("!Allowing test inputs (user name, password, etc)")
var AutoActivateUser = DevDatabase && Alert("?Automatically activating user")

var ProjStructure ProjectStructure

type AnimalOperStruct struct {
	autoLoggedIn bool
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

	ClearAlertHistory(false)
	{
		f := NewPathM("project_structure.json")
		if !f.Exists() {
			BadState("Cannot find:", f)
		}
		ProjStructure = DefaultProjectStructure.Parse(JSMapFromFileM(f)).(ProjectStructure)
	}
	if ProjStructure.DevMachine() {
		DebugUIFlag = true
		ExitOnPanic()
	}

	if true && Alert("doing zoho experiment") {
		oper.zohoExperiment()
		return
	}

	oper.prepareDatabase()

	s := NewJServer(oper)
	s.SessionManager = BuildSessionMap()
	g := ProjStructure
	s.BaseURL = g.BaseUrl()
	s.KeyDir = g.KeyDir()
	s.CertName = g.CertName()
	s.KeyName = g.KeyName()
	SharedWebCache = ConstructSharedWebCache()
	s.BlobCache = SharedWebCache

	// Every several runs, remind to discard tabs
	if oper.DevMode() {
		k := ProjectDirM().JoinM("._SKIP_counter")
		m := JSMapFromFileIfExistsM(k)
		count := m.OptInt("", 0) + 1
		m.Put("", count)
		k.WriteStringM(m.CompactString())
		if count >= 10 {
			k.DeleteFileM()
			Pr(VERT_SP, DASHES, CR, "Take a moment and discard all the tabs")
			SleepMs(4000)
		}
	}

	s.StartServing()
}

// ------------------------------------------------------------------------------------
// ServerApp interface
// ------------------------------------------------------------------------------------

func (oper AnimalOper) PageTemplates() []Page {
	return []Page{
		LandingPageTemplate, ForgotPasswordPageTemplate, CheckMailPageTemplate, GalleryPageTemplate, SignUpPageTemplate,
		FeedPageTemplate, ManagerPageTemplate, ViewAnimalPageTemplate, CreateAnimalPageTemplate,
		EditAnimalPageTemplate,
	}
}

func (oper AnimalOper) Resources() Path {
	return ProjStructure.ResourceDir()
}

func (oper AnimalOper) DevMode() bool {
	return ProjStructure.DevMachine()
}

func (oper AnimalOper) UserForSession(s Session) AbstractUser {
	return OptSessionUser(s)
}

func (oper AnimalOper) DefaultPageForUser(abstractUser AbstractUser) Page {
	if DevGallery {
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
		if !DevAutoLogIn {
			break
		}
		nm := AutoLogInName
		if nm == "" {
			break
		}
		Todo("Auto logging in:", nm)
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
	dataSourcePath := ProjStructure.DbDatasourcePath() //ProjectDirM().JoinM("webapp/sqlite/animal_app_TMP_.db")

	if false && !ProjStructure.DevMachine() && DevDatabase && Alert("Deleting database:", dataSourcePath) {
		DeleteDatabase(dataSourcePath)
	}
	CreateDatabase(dataSourcePath.String())

	if b, _ := ReadBlob(1); b.Id() == 0 {
		// Generate default images as blobs
		animalPicPlaceholderPath := oper.Resources().JoinM("placeholder.jpg")
		b := CreateBlobFromImageFile(animalPicPlaceholderPath)
		//img := CheckOkWith(jimg.DecodeImage(animalPicPlaceholderPath.ReadBytesM()))
		//img = img.ScaleToSize(AnimalPicSizeNormal)
		//jpeg := CheckOkWith(img.ToJPEG())
		//Todo("?Later, keep the original image around for crop adjustments; but for now, scale and store immediately")
		//b := NewBlob()
		//b.SetData(jpeg)
		//AssignBlobName(b)
		//created, err := CreateBlob(b)
		//CheckOk(err)
		CheckState(b.Id() == 1, "unexpected id for placeholder:", b.Id())
	}

	if DevDatabase {
		PopulateDatabase(ProjStructure)
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

func zx(s *strings.Builder, scope string, needComma bool) bool {
	if needComma {
		s.WriteByte(',')
	}
	needComma = true
	s.WriteString("ZohoMail." + scope + ".CREATE")
	s.WriteString(",ZohoMail." + scope + ".READ")
	s.WriteString(",ZohoMail." + scope + ".UPDATE")
	s.WriteString(",ZohoMail." + scope + ".DELETE")
	return needComma
}

func (oper AnimalOper) zohoExperiment() {
	pr := PrIf("zohoExperiment", true)
	pr("folders:")
	pr(SharedZoho().Folders())
}
