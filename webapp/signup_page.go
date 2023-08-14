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
	return auxValidateUserName(s, widget, s.GetValueString(), VALIDATE_EMPTYOK)
}

func auxValidateUserName(s Session, widget Widget, value string, flag ValidateFlag) error {
	pr := PrIf(false)
	pr("auxValidateUserName")

	// It is here in the listener that we read the 'client requested' value for the widget
	// from the ajax parameters, and write it to the state.  We will validate it here.

	pr("value:", value)
	value, err := ValidateUserName(value, flag)
	pr("validated:", value, "error:", err)

	// We want to update the state even if the name is illegal, so user can see what he typed in
	s.State.Put(WidgetId(widget), value)
	s.SetWidgetProblem(widget, err)
	return err
}

func validateUserPwd(s Session, widget Widget) error {
	value := s.GetValueString()
	return auxValidateUserPwd(s, widget, value, VALIDATE_EMPTYOK)
}

func auxValidateUserPwd(s Session, widget Widget, value string, flag ValidateFlag) error {
	pr := PrIf(false)
	pr("auxValidateUserPwd:", value)
	value, err := ValidateUserPassword(value, flag)
	pr("afterward:", value, "err:", err)
	s.State.Put(WidgetId(widget), value)
	s.SetWidgetProblem(widget, err)
	return err
}

func validateMatchPwd(s Session, widget Widget) error {
	value := s.GetValueString()
	return auxValidateMatchPwd(s, widget, value, VALIDATE_EMPTYOK)
}

func auxValidateMatchPwd(s Session, widget Widget, value string, flag ValidateFlag) error {
	if flag.Has(VALIDATE_EMPTYOK) && value == "" {
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
	return auxValidateEmail(s, widget, value, VALIDATE_EMPTYOK)
}

func auxValidateEmail(s Session, widget Widget, value string, flag ValidateFlag) error {
	if flag.Has(VALIDATE_EMPTYOK) && value == "" {
		return nil
	}
	value, err := ValidateEmailAddress(value, flag)

	s.State.Put(WidgetId(widget), value)
	s.SetWidgetProblem(widget, err)

	return err
}

func signUpListener(s Session, widget Widget) error {
	pr := PrIf(false)
	pr("state:", INDENT, s.State)

	auxValidateUserName(s, getWidget(s, id_user_name), s.State.OptString(id_user_name, ""), 0)
	auxValidateUserPwd(s, getWidget(s, id_user_pwd), s.State.OptString(id_user_pwd, ""), 0)
	auxValidateMatchPwd(s, getWidget(s, id_user_pwd_verify), s.State.OptString(id_user_pwd_verify, ""), 0)
	auxValidateEmail(s, getWidget(s, id_user_email), s.State.OptString(id_user_email, ""), 0)

	Todo("if everything worked out, change the displayed page / login state?")
	return nil
}
