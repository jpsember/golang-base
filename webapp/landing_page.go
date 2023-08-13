package webapp

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/webserv"
)

func CreateLandingPage(sess Session) {

	m := sess.WidgetManager()

	m.Col(12)
	m.Label("Landing Page").Size(SizeLarge).AddHeading()
	m.Col(6).Open()
	{
		m.Col(12)
		m.Label("User name").Id("user_name").Listener(userNameListener).AddInput()
		Todo("!Option for password version of input field")
		m.Label("Password").Id("user_pwd").Listener(userPwdListener).AddInput()
		m.Col(6)
		m.AddSpace()
		m.Listener(signInListener)
		m.Id("sign_in").Label("Sign In").AddButton()
	}
	m.Close()

}

func getWidget(sess Session, id string) Widget {
	return sess.WidgetManager().Get(id)
}

func userNameListener(sess any, widget Widget) {
	Pr("userNameListener", widget.Base().Id)
	s := sess.(Session)
	Todo("some redundancy here, as the id and value are found in the ajax args...")
	wid := s.GetWidgetId()
	s.State.Put(wid, s.GetValueString())
	s.ClearWidgetProblem(widget)
	s.Repaint(widget)
}

func userPwdListener(sess any, widget Widget) {
	Pr("userPwdListener", widget.Base().Id)
	s := sess.(Session)
	wid := s.GetWidgetId()
	s.State.Put(wid, s.GetValueString())
	Todo("if clearing the problem, it should repaint")
	s.ClearWidgetProblem(widget)
	s.Repaint(widget)
}

func signInListener(sess any, widget Widget) {

	s := sess.(Session)

	pr := PrIf(true)

	pr("state:", INDENT, s.State)

	// When user modifies a widget in the browser, the Ajax call stores that user value directly
	// into the state, without any validation.  So we must be sure to perform validation on it, and
	// restore (in some way) whatever value was there before

	browserUserName := getWidget(s, "user_name")
	browserPassword := getWidget(s, "user_pwd")

	Todo("have utility method to read widget value from state")

	userName := s.State.OptString("user_name", "")
	pwd := s.State.OptString("user_pwd", "")

	s.ClearWidgetProblem(browserUserName)
	s.ClearWidgetProblem(browserPassword)
	if userName == "" {
		s.SetWidgetProblem(browserUserName, "Please enter your name")
		s.Repaint(browserUserName)
	}
	if pwd == "" {
		s.SetWidgetProblem(browserPassword, "Please enter your password")
		s.Repaint(browserPassword)
	}

}
