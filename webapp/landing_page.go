package webapp

import (
	. "github.com/jpsember/golang-base/base"
	"github.com/jpsember/golang-base/webapp/gen/webapp_data"
	. "github.com/jpsember/golang-base/webserv"
)

type LandingPageStruct struct {
	sess         Session
	parentWidget Widget
}

type LandingPage = *LandingPageStruct

func NewLandingPage(sess Session, parentWidget Widget) LandingPage {
	t := &LandingPageStruct{
		sess:         sess,
		parentWidget: parentWidget,
	}
	return t
}

func (p LandingPage) Generate() {
	s := p.sess.State
	s.DeleteEach(id_user_name, id_user_pwd, id_user_pwd_verify, id_user_email)

	if false && Alert("!prefilling user name and password") {
		s.Put(id_user_name, "Bartholemew").Put(id_user_pwd, "01234password")
	}

	m := p.sess.WidgetManager()
	m.With(p.parentWidget)

	m.Col(12)
	m.Label("Landing Page").Size(SizeLarge).AddHeading()
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

	errcount := WidgetErrorCount(p.parentWidget, sess.State)
	if errcount != 0 {
		return
	}

	userId, err := Db().FindUserWithName(userName)
	if err != nil {
		sess.SetWidgetIdProblem(id_user_name, "No such user, or incorrect password")
		return
	}

	if IsUserLoggedIn(userId) {
		Todo("Log user out of other sessions?")
		sess.SetWidgetIdProblem(id_user_name, "User is already logged in")
		return
	}

	userData, _ := Db().ReadUser(userId)
	if userData == nil {
		sess.SetWidgetIdProblem(id_user_name, "User is unavaliable; sorry")
		return
	}
	if AutoActivateUser {
		if userData.State() == webapp_data.UserstateWaitingActivation {
			Alert("Activating user automatically (without email verification)")
			userData = userData.ToBuilder().SetState(webapp_data.UserstateActive).Build()
			Db().UpdateUser(userData)
		}
	}
	errMsg := ""
	switch userData.State() {
	case webapp_data.UserstateActive:
		// This is ok.
	case webapp_data.UserstateWaitingActivation:
		errMsg = "This user has not been activated yet"
	default:
		errMsg = "This user is in an unsupported state"
		Alert("Unsupported user state:", INDENT, userData)
	}

	if errMsg != "" {
		sess.SetWidgetIdProblem(id_user_name, errMsg)
		return
	}

	if !TryRegisteringUserAsLoggedIn(userId, true) {
		sess.SetWidgetIdProblem(id_user_name, "Unable to log in at this time")
		return
	}

	sp := NewAnimalFeedPage(sess, p.parentWidget)
	sp.Generate()
}

func (p LandingPage) signUpListener(s Session, widget Widget) {
	sp := NewSignUpPage(s, p.parentWidget)
	sp.Generate()
}

func (p LandingPage) forgotPwdListener(sess Session, widget Widget) {

	s := sess.State
	userName := s.OptString(id_user_name, "")

	if userName == "" {
		Todo("disable button if no user name entered")
		return
	}

	userId, err := Db().FindUserWithName(userName)

	if err != nil {
		Alert("Not revealing that 'no such user exists' in forgot password logic")
	}
	if userId != 0 {
		Todo("Send email")
	}
	sess.SetWidgetIdProblem(id_user_name, "An email has been sent with a link to change your password.")
}
