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

func validateUserName(s Session, widget Widget, value string, emptyOk bool) error {
	pr := PrIf(true)
	pr("validateUserName")

	// It is here in the listener that we read the 'client requested' value for the widget
	// from the ajax parameters, and write it to the state.  We will validate it here.

	pr("value:", value)
	value, err := ValidateUserName(value, emptyOk)
	pr("validated:", value, "error:", err)

	Todo("Utility function for the following boilerplate?")
	// We want to update the state even if the name is illegal, so user can see what he typed in
	s.State.Put(WidgetId(widget), value)

	if err != nil {
		s.SetWidgetProblem(widget, err.Error())
	} else {
		s.ClearWidgetProblem(widget)
	}
	return err
}

func userNameListener(sess any, widget Widget) error {
	pr := PrIf(false)
	pr("userNameListener", WidgetId(widget))
	s := sess.(Session)

	// It is here in the listener that we read the 'client requested' value for the widget
	// from the ajax parameters, and write it to the state.  We will validate it here.

	value := s.GetValueString()
	return validateUserName(s, widget, value, true)
}

func validateUserPwd(s Session, widget Widget, value string, emptyOk bool) error {
	pr := PrIf(false)
	pr("validateUserPwd:", value)

	value, err := ValidateUserPassword(value, emptyOk)
	pr("afterward:", value, "err:", err)

	s.State.Put(WidgetId(widget), value)
	if err != nil {
		s.SetWidgetProblem(widget, err.Error())
	} else {
		s.ClearWidgetProblem(widget)
	}
	return err
}

func userPwdListener(sess any, widget Widget) error {
	s := sess.(Session)
	value := s.GetValueString()
	return validateUserPwd(s, widget, value, true)
}

func signInListener(sess any, widget Widget) error {
	s := sess.(Session)

	pr := PrIf(true)
	pr("state:", INDENT, s.State)

	browserUserName := getWidget(s, "user_name")
	browserPassword := getWidget(s, "user_pwd")

	err1 := validateUserName(s, browserUserName, s.State.OptString("user_name", ""), false)
	err2 := validateUserPwd(s, browserPassword, s.State.OptString("user_pwd", ""), false)

	pr("user_name err:", err1, "user_pwd err:", err2)
	Todo("if everything worked out, change the displayed page / login state?")
	return err1
}
