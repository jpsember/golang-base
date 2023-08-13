package webapp

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/webserv"
)

func CreateLandingPage(sess Session) {

	m := sess.WidgetManager()

	m.Col(12)
	m.Text("Landing Page").Size(SizeLarge).AddHeading()
	m.Col(6).Open()
	{
		m.Col(12)
		m.Label("User name").Id("user_name").AddInput()
		Todo("!Option for password version of input field")
		m.Label("Password").Id("user_pwd").AddInput()
		m.Col(6)
		m.AddSpace()
		m.Listener(signInListener)
		m.Id("sign_in").Text("Sign In").AddButton()
	}
	m.Close()

}

func signInListener(sess any, widget Widget) {
	s := sess.(Session)
	wid := s.GetWidgetId()
	Todo("do something to sign in the user")
	Pr("signInListener", wid)
}
