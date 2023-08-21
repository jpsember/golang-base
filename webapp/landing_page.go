package webapp

import (
	. "github.com/jpsember/golang-base/base"
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

	if Alert("!prefilling user name and password") {
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
		m.Listener(p.signInListener).Label("Sign In").AddButton()
	}
	m.Close()
	m.Open()
	{
		m.Listener(p.signUpListener)
		m.Label("Sign Up").AddButton()
	}
	m.Close()
}

func (p LandingPage) validateUserName(s Session, widget Widget) error {
	return auxValidateUserName(s, widget, s.GetValueString(), VALIDATE_ONLY_NONEMPTY)
}

func (p LandingPage) validateUserPwd(s Session, widget Widget) error {
	value := s.GetValueString()
	return auxValidateUserPwd(s, widget, value, VALIDATE_ONLY_NONEMPTY)
}

func (p LandingPage) signInListener(s Session, widget Widget) error {

	userName := s.State.OptString(id_user_name, "")
	pwd := s.State.OptString(id_user_pwd, "")
	Todo("ability to read value using widget id")
	if userName == "" {
		s.SetWidgetProblem(getWidget(s, id_user_name), ErrorEmptyUserName)
	}
	if pwd == "" {
		s.SetWidgetProblem(getWidget(s, id_user_pwd), ErrorEmptyUserPassword)

	}

	errcount := WidgetErrorCount(p.parentWidget, s.State)
	Pr("error count:", errcount)
	if errcount == 0 {
		sp := NewAnimalFeedPage(s, p.parentWidget)
		sp.Generate()
	}
	return nil
}

func (p LandingPage) signUpListener(s Session, widget Widget) error {
	sp := NewSignUpPage(s, p.parentWidget)
	sp.Generate()
	return nil
}
