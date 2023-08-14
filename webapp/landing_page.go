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

	m := p.sess.WidgetManager()

	m.Col(12)
	m.Label("Landing Page").Size(SizeLarge).AddHeading()
	m.Col(6)
	m.Open()
	{
		m.Col(12)
		m.Label("User name").Id(id_user_name).Listener(validateUserName).AddInput()
		m.Label("Password").Id(id_user_pwd).Listener(validateUserPwd).AddPassword()
		m.Listener(p.signInListener).Label("Sign In").AddButton()
	}
	m.Close()
	m.Open()
	{
		Todo("does it reset columns to 12?")
		m.Listener(signUpListener)
		m.Label("Sign Up").AddButton()
	}
	m.Close()
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
		Pr("no errors, continuing")
		Todo("if everything worked out, change the displayed page / login state?")
	}
	return nil
}
