package webapp

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/webapp/gen/webapp_data"
	. "github.com/jpsember/golang-base/webserv"
)

const (
	id_user_name       = "user_name"
	id_user_pwd_verify = "user_pwd_verify"
	id_user_pwd        = "user_pwd"
	id_user_email      = "user_email"
	id_sign_up         = "sign_up"
)

type SignUpPageStruct struct {
	sess         Session
	parentWidget Widget
}

type SignUpPage = *SignUpPageStruct

func NewSignUpPage(sess Session, parentWidget Widget) SignUpPage {
	t := &SignUpPageStruct{
		sess:         sess,
		parentWidget: parentWidget,
	}
	return t
}

func (p SignUpPage) Generate() {

	s := p.sess.State
	s.DeleteEach(id_user_name, id_user_pwd, id_user_pwd_verify, id_user_email)
	Todo("Delete auxilliary versions of these as well; i.e., try to log in as a non-existent user, then switch to sign up")
	m := p.sess.WidgetManager()
	m.With(p.parentWidget)

	m.Col(12)
	m.Label("Sign Up Page").Size(SizeLarge).AddHeading()
	m.Col(6).Open()
	{
		m.Col(12)
		m.Label("User name").Id(id_user_name).Listener(p.validateUserName).AddInput()
		m.Label("Password").Id(id_user_pwd).Listener(p.validateUserPwd).AddPassword()
		m.Label("Password Again").Id(id_user_pwd_verify).Listener(p.validateMatchPwd).AddPassword()
		m.Label("Email").Id(id_user_email).Listener(p.validateEmail).AddInput()
		m.Size(SizeTiny).Label("We will never share your email address with anyone.").AddText()
		m.Col(6)
		m.AddSpace()
		m.Listener(p.signUpListener)
		m.Id(id_sign_up).Label("Sign Up").AddButton()
	}
	m.Close()
}

func getWidget(sess Session, id string) Widget {
	return sess.WidgetManager().Get(id)
}

func (p SignUpPage) validateUserName(s Session, widget Widget) {
	// It is here in the listener that we read the 'client requested' value for the widget
	// from the ajax parameters, and write it to the state.  We will validate it here.
	auxValidateUserName(s, widget, s.GetValueString(), VALIDATE_EMPTYOK)
}

func auxValidateUserName(s Session, widget Widget, value string, flag ValidateFlag) {
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
}

func (p SignUpPage) validateUserPwd(s Session, widget Widget) {
	value := s.GetValueString()
	auxValidateUserPwd(s, widget, value, VALIDATE_EMPTYOK)
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

func (p SignUpPage) validateMatchPwd(s Session, widget Widget) {
	value := s.GetValueString()
	p.auxValidateMatchPwd(s, widget, value, VALIDATE_EMPTYOK)
}

func (p SignUpPage) auxValidateMatchPwd(s Session, widget Widget, value string, flag ValidateFlag) {
	if flag.Has(VALIDATE_EMPTYOK) && value == "" {
		return
	}
	var err error
	value1 := s.State.OptString(id_user_pwd, "")
	err, value = replaceWithTestInput(err, value, "a", value1)
	if value1 != value {
		err = Ternary(value == "", ErrorEmptyUserPassword, ErrorUserPasswordsDontMatch)
	}

	s.State.Put(WidgetId(widget), value)
	s.SetWidgetProblem(widget, err)
}

func (p SignUpPage) validateEmail(s Session, widget Widget) {
	value := s.GetValueString()
	auxValidateEmail(s, widget, value, VALIDATE_EMPTYOK)
}

func auxValidateEmail(s Session, widget Widget, value string, flag ValidateFlag) {
	if flag.Has(VALIDATE_EMPTYOK) && value == "" {
		return
	}
	var err error
	value, err = ValidateEmailAddress(value, flag)
	s.State.Put(WidgetId(widget), value)
	s.SetWidgetProblem(widget, err)
}

func (p SignUpPage) signUpListener(s Session, widget Widget) {
	pr := PrIf(true)
	pr("state:", INDENT, s.State)

	auxValidateUserName(s, getWidget(s, id_user_name), s.State.OptString(id_user_name, ""), 0)
	auxValidateUserPwd(s, getWidget(s, id_user_pwd), s.State.OptString(id_user_pwd, ""), 0)
	p.auxValidateMatchPwd(s, getWidget(s, id_user_pwd_verify), s.State.OptString(id_user_pwd_verify, ""), 0)
	auxValidateEmail(s, getWidget(s, id_user_email), s.State.OptString(id_user_email, ""), 0)

	b := NewUser()
	b.SetName(s.State.OptString(id_user_name, ""))
	b.SetPassword(s.State.OptString(id_user_pwd, ""))
	b.SetEmail(s.State.OptString(id_user_email, ""))

	ub, err := Db().CreateUser(b)

	Pr("created user:", INDENT, ub)
	if err == UserExistsError {
		s.SetWidgetIdProblem(id_user_name, "This user already exists")
		return
	}

	CheckOk(err)

	Todo("add support for WaitingActivation")
	Todo("if everything worked out, change the displayed page / login state?")
}
