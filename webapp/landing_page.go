package webapp

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/webapp/gen/webapp_data"
	. "github.com/jpsember/golang-base/webserv"
)

// ------------------------------------------------------------------------------------
// Page implementation
// ------------------------------------------------------------------------------------

const LandingPageName = "signin"

var LandingPageTemplate = NewLandingPage(nil)

// ------------------------------------------------------------------------------------

type LandingPageStruct struct {
	session  Session
	animalId int
	editing  bool
}

type LandingPage = *LandingPageStruct

func (p LandingPage) Request(s Session, parser PathParse) Page {
	user := OptSessionUser(s)
	if user.Id() == 0 {
		return LandingPageTemplate
	}
	return nil
}

func NewLandingPage(sess Session, args ...any) Page {
	t := &LandingPageStruct{
		session: sess,
		editing: true,
	}
	return t
}
func (p LandingPage) Session() Session { return p.session }

func (p LandingPage) Name() string {
	return LandingPageName
}

func (p LandingPage) Construct(s Session, args ...any) Page {
	return NewLandingPage(s, args...)
}
func (p LandingPage) Args() []any { return EmptyPageArgs }

func (p LandingPage) Generate() {
	sess := p.Session()
	CheckState(sess != nil, "There is no session!")
	s := sess.State
	CheckState(s != nil, "there is no State for the session!")
	s.DeleteEach(id_user_name, id_user_pwd, id_user_pwd_verify, id_user_email)

	m := GenerateHeader(p)

	m.Label("gallery").Align(AlignRight).Size(SizeTiny).AddButton(p.galleryListener)
	m.Col(6)
	m.Open()
	{
		m.Col(12)
		m.Label("User name").Id(id_user_name).AddInput(p.validateUserName)
		m.Label("Password").Id(id_user_pwd).AddPassword(p.validateUserPwd)
		m.Open()
		m.Col(6)
		{
			m.Label("Sign In").AddButton(p.signInListener)
			m.Label("I forgot my password").Size(SizeTiny).AddButton(p.forgotPwdListener)
		}
		m.Close()
	}
	m.Close()
	m.Open()
	{
		m.Label("Sign Up").AddButton(p.signUpListener)
	}
	m.Close()
}

func (p LandingPage) validateUserName(s Session, widget InputWidget, name string) (string, error) {
	return ValidateUserName(name, VALIDATE_EMPTYOK)
}

func (p LandingPage) validateUserPwd(s Session, widget InputWidget, content string) (string, error) {
	return ValidateUserPassword(content, VALIDATE_ONLY_NONEMPTY)
}

var AutoActivateUser = DevDatabase && Alert("?Automatically activating user")

func (p LandingPage) signInListener(sess Session, widget Widget) {
	pr := PrIf(true)
	s := sess.State
	pr("signInListener; state:", INDENT, s)
	userName := s.OptString(id_user_name, "")
	pwd := s.OptString(id_user_pwd, "")

	var err1 error
	userName, err1 = ValidateUserName(userName, VALIDATE_ONLY_NONEMPTY)
	var err2 error
	pwd, err2 = ValidateEmailAddress(pwd, VALIDATE_ONLY_NONEMPTY)

	pr("id_user_name problem:", err1)
	pr("id_user_pwd  problem:", err2)
	sess.SetWidgetProblem(id_user_name, err1)
	sess.SetWidgetProblem(id_user_pwd, err2)

	var user = DefaultUser
	prob := ""
	for {
		errcount := WidgetErrorCount(sess.PageWidget, sess.State)
		pr("errcount:", errcount)
		if errcount != 0 {
			break
		}

		var err error
		user, err = ReadUserWithName(userName)
		ReportIfError(err)
		userId := user.Id()

		prob = "No such user, or incorrect password"
		if userId == 0 {
			break
		}

		prob = "User is already logged in"
		if IsUserLoggedIn(userId) {
			return
		}

		prob = "User is unavaliable; sorry"
		userData := ReadUserIgnoreError(userId)
		if userData.Id() == 0 {
			break
		}

		if AutoActivateUser {
			if userData.State() == UserStateWaitingActivation {
				Alert("Activating user automatically (without email verification)")
				userData = userData.ToBuilder().SetState(UserStateActive).Build()
				UpdateUser(userData)
			}
		}

		prob = ""
		switch userData.State() {
		case UserStateActive:
			// This is ok.
		case UserStateWaitingActivation:
			prob = "This user has not been activated yet"
		default:
			prob = "This user is in an unsupported state"
		}
		if prob != "" {
			break
		}

		prob = "Unable to log in at this time"
		if !TryLoggingIn(sess, user) {
			break
		}

		prob = ""
		break
	}
	pr("problem is:", prob)
	if prob != "" {
		sess.SetWidgetProblem(id_user_name, prob)
	} else {
		switch user.UserClass() {
		case UserClassDonor:
			NewFeedPage(sess).Generate()
			break
		case UserClassManager:
			NewManagerPage(sess).Generate()
		}
	}
}

func (p LandingPage) signUpListener(s Session, widget Widget) {
	NewSignUpPage(s).Generate()
}

func (p LandingPage) galleryListener(sess Session, widget Widget) {
	NewGalleryPage(sess).Generate()
}

func (p LandingPage) forgotPwdListener(sess Session, widget Widget) {

	for {

		userEmail := sess.WidgetStrValue(id_user_email)

		if userEmail == "" {
			sess.SetWidgetProblem(id_user_email, "Please enter your email address.")
			return
		}

		sess.SetWidgetProblem(id_user_name, "An email has been sent with a link to change your password.")

		//user, err := webapp_data.ReadUserWithName(userName)
		//userId := user.Id()
		//
		//if err != nil {
		//	Alert("Not revealing that 'no such user exists' in forgot password logic")
		//}
		//if userId != 0 {
		//	Todo("Send email")
		//}
		//break
	}

}
