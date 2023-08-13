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
		m.Label("Password").Id("user_pwd").Listener(userPwdListener).AddPassword()
		m.Label("Password Again").Id("user_pwd_verify").Listener(verifyUserPwdListener).AddPassword()
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
	pr := PrIf(false)
	pr("validateUserName")

	// It is here in the listener that we read the 'client requested' value for the widget
	// from the ajax parameters, and write it to the state.  We will validate it here.

	pr("value:", value)
	value, err := ValidateUserName(value, emptyOk)
	pr("validated:", value, "error:", err)

	// We want to update the state even if the name is illegal, so user can see what he typed in
	s.State.Put(WidgetId(widget), value)
	s.SetWidgetProblem(widget, err)
	return err
}

func userNameListener(s Session, widget Widget) error {
	pr := PrIf(false)
	pr("userNameListener", WidgetId(widget))

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
	s.SetWidgetProblem(widget, err)
	return err
}

func userPwdListener(s Session, widget Widget) error {
	value := s.GetValueString()
	return validateUserPwd(s, widget, value, true)
}

func validateMatchingPassword(s Session, widget Widget, value string, emptyOk bool) error {
	if emptyOk && value == "" {
		return nil
	}
	var err error
	value1 := s.State.OptString("user_pwd", "")
	if value1 != value {
		err = Ternary(value == "", ErrorEmptyUserPassword, ErrorUserPasswordsDontMatch)
	}
	s.State.Put(WidgetId(widget), value)
	s.SetWidgetProblem(widget, err)
	return err
}

func verifyUserPwdListener(s Session, widget Widget) error {
	value := s.GetValueString()
	return validateMatchingPassword(s, widget, value, true)
}

func signInListener(s Session, widget Widget) error {
	pr := PrIf(false)
	pr("state:", INDENT, s.State)

	browserUserName := getWidget(s, "user_name")
	browserPassword := getWidget(s, "user_pwd")

	err1 := validateUserName(s, browserUserName, s.State.OptString("user_name", ""), false)
	err2 := validateUserPwd(s, browserPassword, s.State.OptString("user_pwd", ""), false)
	err3 := validateMatchingPassword(s, getWidget(s, "user_pwd_verify"), s.State.OptString("user_pwd_verify", ""), false)

	pr("user_name err:", err1, "user_pwd err:", err2, "user_pwd_verify err:", err3)
	Todo("if everything worked out, change the displayed page / login state?")
	return nil
}
