package webapp

import (
	. "github.com/jpsember/golang-base/base"
	"github.com/jpsember/golang-base/webapp/gen/webapp_data"
	. "github.com/jpsember/golang-base/webserv"
)

type LandingPageStruct struct {
	BasicPage
}

type LandingPage = *LandingPageStruct

func NewLandingPage(sess Session, parentWidget Widget) LandingPage {
	t := &LandingPageStruct{
		NewBasicPage(sess, parentWidget),
	}
	t.devLabel = "landing_page"
	return t
}

func (p LandingPage) Generate() {
	s := p.session.State
	s.DeleteEach(id_user_name, id_user_pwd, id_user_pwd_verify, id_user_email)

	m := p.GenerateHeader()

	m.Label("gallery").Align(AlignRight).Size(SizeTiny).Listener(p.galleryListener).AddButton()
	m.Col(6)
	m.Open()
	{
		m.Col(12)
		m.Label("User name").Id(id_user_name).Listener(
			p.validateUserName).AddInput()
		m.Label("Password").Id(id_user_pwd).Listener(p.validateUserPwd).AddPassword()
		m.Open()
		m.Col(6)
		{
			m.Listener(p.signInListener).Label("Sign In").AddButton()
			m.Listener(p.forgotPwdListener).Label("I forgot my password")
			m.Size(SizeTiny)
			m.AddButton()
		}
		m.Close()
	}
	m.Close()
	m.Open()
	{
		m.Listener(p.signUpListener)
		m.Label("Sign Up").AddButton()
	}
	m.Close()
}

func (p LandingPage) validateUserName(s Session, widget Widget) {
	auxValidateUserName(s, widget, s.GetValueString(), VALIDATE_ONLY_NONEMPTY)
}

func (p LandingPage) validateUserPwd(s Session, widget Widget) {
	value := s.GetValueString()
	auxValidateUserPwd(s, widget, value, VALIDATE_ONLY_NONEMPTY)
}

func (p LandingPage) signInListener(sess Session, widget Widget) {

	s := sess.State
	userName := s.OptString(id_user_name, "")
	pwd := s.OptString(id_user_pwd, "")

	var err1 error
	userName, err1 = ValidateUserName(userName, VALIDATE_ONLY_NONEMPTY)
	var err2 error
	pwd, err2 = ValidateEmailAddress(pwd, VALIDATE_ONLY_NONEMPTY)

	sess.SetWidgetIdProblem(id_user_name, err1)
	sess.SetWidgetIdProblem(id_user_pwd, err2)

	errcount := WidgetErrorCount(p.parentPage, sess.State)
	if errcount != 0 {
		return
	}

	user, err := webapp_data.ReadUserWithName(userName)
	userId := user.Id()
	CheckOk(err)

	if userId == 0 {
		sess.SetWidgetIdProblem(id_user_name, "No such user, or incorrect password")
		return
	}

	if IsUserLoggedIn(userId) {
		Todo("Log user out of other sessions?")
		sess.SetWidgetIdProblem(id_user_name, "User is already logged in")
		return
	}

	userData, _ := webapp_data.ReadUser(userId)
	if userData.Id() == 0 {
		sess.SetWidgetIdProblem(id_user_name, "User is unavaliable; sorry")
		return
	}
	if AutoActivateUser {
		if userData.State() == webapp_data.UserStateWaitingActivation {
			Alert("Activating user automatically (without email verification)")
			userData = userData.ToBuilder().SetState(webapp_data.UserStateActive).Build()
			webapp_data.UpdateUser(userData)
		}
	}
	errMsg := ""
	switch userData.State() {
	case webapp_data.UserStateActive:
		// This is ok.
	case webapp_data.UserStateWaitingActivation:
		errMsg = "This user has not been activated yet"
	default:
		errMsg = "This user is in an unsupported state"
		Alert("Unsupported user state:", INDENT, userData)
	}

	if errMsg != "" {
		sess.SetWidgetIdProblem(id_user_name, errMsg)
		return
	}

	if !TryRegisteringUserAsLoggedIn(sess, user, true) {
		sess.SetWidgetIdProblem(id_user_name, "Unable to log in at this time")
		return
	}

	switch user.UserClass() {
	case webapp_data.UserClassDonor:
		sp := NewAnimalFeedPage(sess, p.parentPage)
		sp.Generate()
		break
	case webapp_data.UserClassManager:
		Todo("?Maybe make AnimalFeed, Manager pages implement a common interface")
		sp := NewManagerPage(sess, p.parentPage)
		sp.Generate()
	}

}

func (p LandingPage) signUpListener(s Session, widget Widget) {
	NewSignUpPage(s, widget).Generate()
}

func (p LandingPage) galleryListener(sess Session, widget Widget) {
	NewGalleryPage(sess, p.parentPage).Generate()
}

func (p LandingPage) forgotPwdListener(sess Session, widget Widget) {

	s := sess.State
	userName := s.OptString(id_user_name, "")

	if userName == "" {
		Todo("disable button if no user name entered")
		return
	}

	user, err := webapp_data.ReadUserWithName(userName)
	userId := user.Id()

	if err != nil {
		Alert("Not revealing that 'no such user exists' in forgot password logic")
	}
	if userId != 0 {
		Todo("Send email")
	}
	sess.SetWidgetIdProblem(id_user_name, "An email has been sent with a link to change your password.")
}
