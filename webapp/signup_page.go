package webapp

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/webserv"
)

const (
	id_user_name       = "user_name"
	id_user_pwd_verify = "user_pwd_verify"
	id_user_pwd        = "user_pwd"
	id_user_email      = "user_email"
	id_sign_up         = "sign_up"
)

func GenerateSignUpView(sess Session) {

	m := sess.WidgetManager()

	m.Col(12)
	m.Label("Sign Up Page").Size(SizeLarge).AddHeading()
	m.Col(6).Open()
	{
		m.Col(12)
		m.Label("User name").Id(id_user_name).Listener(validateUserName).AddInput()
		m.Label("Password").Id(id_user_pwd).Listener(validateUserPwd).AddPassword()
		m.Label("Password Again").Id(id_user_pwd_verify).Listener(validateMatchPwd).AddPassword()
		m.Label("Email").Id(id_user_email).Listener(validateEmail).AddInput()
		m.Size(SizeTiny).Label("We will never share your email address with anyone.").AddText()
		m.Col(6)
		m.AddSpace()
		m.Listener(signUpListener)
		m.Id(id_sign_up).Label("Sign In").AddButton()
	}
	m.Close()

}

func getWidget(sess Session, id string) Widget {
	return sess.WidgetManager().Get(id)
}

func validateUserName(s Session, widget Widget) error {
	// It is here in the listener that we read the 'client requested' value for the widget
	// from the ajax parameters, and write it to the state.  We will validate it here.
	return auxValidateUserName(s, widget, s.GetValueString(), true)
}

func auxValidateUserName(s Session, widget Widget, value string, emptyOk bool) error {
	pr := PrIf(false)
	pr("auxValidateUserName")

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

func validateUserPwd(s Session, widget Widget) error {
	value := s.GetValueString()
	return auxValidateUserPwd(s, widget, value, true)
}

func auxValidateUserPwd(s Session, widget Widget, value string, emptyOk bool) error {
	pr := PrIf(false)
	pr("auxValidateUserPwd:", value)
	value, err := ValidateUserPassword(value, emptyOk)
	pr("afterward:", value, "err:", err)
	s.State.Put(WidgetId(widget), value)
	s.SetWidgetProblem(widget, err)
	return err
}

func validateMatchPwd(s Session, widget Widget) error {
	value := s.GetValueString()
	return auxValidateMatchPwd(s, widget, value, true)
}

func auxValidateMatchPwd(s Session, widget Widget, value string, emptyOk bool) error {
	if emptyOk && value == "" {
		return nil
	}
	var err error
	value1 := s.State.OptString(id_user_pwd, "")
	if value1 != value {
		err = Ternary(value == "", ErrorEmptyUserPassword, ErrorUserPasswordsDontMatch)
	}
	s.State.Put(WidgetId(widget), value)
	s.SetWidgetProblem(widget, err)
	return err
}

func validateEmail(s Session, widget Widget) error {
	value := s.GetValueString()
	return auxValidateEmail(s, widget, value, true)
}

func auxValidateEmail(s Session, widget Widget, value string, emptyOk bool) error {
	if emptyOk && value == "" {
		return nil
	}
	value, err := ValidateEmailAddress(value, emptyOk)

	s.State.Put(WidgetId(widget), value)
	s.SetWidgetProblem(widget, err)

	return err
}

func signUpListener(s Session, widget Widget) error {
	pr := PrIf(false)
	pr("state:", INDENT, s.State)

	auxValidateUserName(s, getWidget(s, id_user_name), s.State.OptString(id_user_name, ""), false)
	auxValidateUserPwd(s, getWidget(s, id_user_pwd), s.State.OptString(id_user_pwd, ""), false)
	auxValidateMatchPwd(s, getWidget(s, id_user_pwd_verify), s.State.OptString(id_user_pwd_verify, ""), false)
	auxValidateEmail(s, getWidget(s, id_user_email), s.State.OptString(id_user_email, ""), false)

	Todo("if everything worked out, change the displayed page / login state?")
	return nil
}
