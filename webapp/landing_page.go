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

func (p LandingPage) signInListener(sess Session, widget Widget) error {

	s := sess.State
	userName := s.OptString(id_user_name, "")
	pwd := s.OptString(id_user_pwd, "")

	var err1 error
	userName, err1 = ValidateUserName(userName, VALIDATE_ONLY_NONEMPTY)
	var err2 error
	pwd, err2 = ValidateEmailAddress(pwd, VALIDATE_ONLY_NONEMPTY)

	sess.SetWidgetIdProblem(id_user_name, err1)
	sess.SetWidgetIdProblem(id_user_pwd, err2)

	var user webapp_data.User
	prob := ""
	for {
		errcount := WidgetErrorCount(p.parentPage, sess.State)
		if errcount != 0 {
			break
		}

		var err error
		user, err = webapp_data.ReadUserWithName(userName)
		userId := user.Id()
		CheckOk(err)

		prob = "No such user, or incorrect password"
		if userId == 0 {
			break
		}

		prob = "User is already logged in"
		if IsUserLoggedIn(userId) {
			return nil
		}

		prob = "User is unavaliable; sorry"
		userData, _ := webapp_data.ReadUser(userId)
		if userData.Id() == 0 {
			break
		}

		if AutoActivateUser {
			if userData.State() == webapp_data.UserStateWaitingActivation {
				Alert("Activating user automatically (without email verification)")
				userData = userData.ToBuilder().SetState(webapp_data.UserStateActive).Build()
				webapp_data.UpdateUser(userData)
			}
		}

		prob = ""
		switch userData.State() {
		case webapp_data.UserStateActive:
			// This is ok.
		case webapp_data.UserStateWaitingActivation:
			prob = "This user has not been activated yet"
		default:
			prob = "This user is in an unsupported state"
		}
		if prob != "" {
			break
		}

		prob = "Unable to log in at this time"
		if !TryRegisteringUserAsLoggedIn(sess, user, true) {
			break
		}

		prob = ""
		break
	}
	if prob != "" {
		sess.SetWidgetIdProblem(id_user_name, prob)
	} else {
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
	return nil
}

func (p LandingPage) signUpListener(s Session, widget Widget) error {
	NewSignUpPage(s, p.parentPage).Generate()
	return nil
}

func (p LandingPage) galleryListener(sess Session, widget Widget) error {
	NewGalleryPage(sess, p.parentPage).Generate()
	return nil
}

func (p LandingPage) forgotPwdListener(sess Session, widget Widget) error {

	problem := ""
	for {

		s := sess.State
		userName := s.OptString(id_user_name, "")

		Todo("Change this to 'enter your email' instead")
		problem = "Please enter your user name"
		if userName == "" {
			break
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
		break
	}

	if problem != "" {
		return Error(problem)
	}
	return nil
}
