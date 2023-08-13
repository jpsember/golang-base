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
	Pr("userNameListener", widget.GetBaseWidget().Id)
	s := sess.(Session)
	Todo("some redundancy here, as the id and value are found in the ajax args...")
	wid := s.GetWidgetId()
	s.State.Put(wid, s.GetValueString())
	s.ClearWidgetProblem(widget)
	s.Repaint(widget)
}

func userPwdListener(sess any, widget Widget) {
	Pr("userPwdListener", widget.GetBaseWidget().Id)
	s := sess.(Session)
	wid := s.GetWidgetId()
	s.State.Put(wid, s.GetValueString())
	Todo("if clearing the problem, it should repaint")
	s.ClearWidgetProblem(widget)
	s.Repaint(widget)
}

func signInListener(sess any, widget Widget) {
	s := sess.(Session)

	//wid := s.GetWidgetId()

	Pr("state:", INDENT, s.State)

	wUserName := getWidget(s, "user_name")
	wPwd := getWidget(s, "user_pwd")

	Pr("wUserName:", Info(wUserName))

	userName := s.State.OptString("user_name", "")
	pwd := s.State.OptString("user_pwd", "")

	s.ClearWidgetProblem(wUserName)
	s.ClearWidgetProblem(wPwd)
	if userName == "" {
		s.SetWidgetProblem(wUserName, "Please enter your name")
		s.Repaint(wUserName)
	}
	if pwd == "" {
		s.SetWidgetProblem(wPwd, "Please enter your password")
		s.Repaint(wPwd)
	}
	//Pr("user_name readValue:",
	//	wUserName.ReadValue())
	////wUser := s.WidgetManager().Get("user_name")
	////wPwd := s.WidgetManager().Get("user_pwd")
	//Pr("user_name value:", wUser.ReadValue().AsString())
	//Pr("user_pwd value:", wPwd.ReadValue().AsString())

}
