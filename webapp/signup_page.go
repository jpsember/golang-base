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

// ------------------------------------------------------------------------------------
// This is the canonical boilerplate that I will turn into a goland live template,
// to simplify creating new pages:
// ------------------------------------------------------------------------------------

type SignUpPageStruct struct {
}
type SignUpPage = *SignUpPageStruct

func NewSignUpPage(session Session) SignUpPage {
	t := &SignUpPageStruct{}
	return t
}

var SignUpPageTemplate = NewSignUpPage(nil)

const SignUpPageName = "signup"

func (p SignUpPage) Name() string   { return SignUpPageName }
func (p SignUpPage) Args() []string { return EmptyStringSlice }

func (p SignUpPage) Construct(s Session, args PageArgs) Page {
	if args.CheckDone() {
		if OptSessionUser(s).Id() == 0 {
			return NewSignUpPage(s)
		}
	}
	return nil
}

func (p SignUpPage) GenerateWidgets(s Session) {
	s.DeleteStateErrors()
	m := GenerateHeader(s, p)
	m.Label("Sign Up Page").Size(SizeLarge).AddHeading()
	m.Col(6).Open()
	{
		m.Col(12)
		m.Label("User name").Id(id_user_name).AddInput(p.validateUserName)
		m.Label("Password").Id(id_user_pwd).AddPassword(p.validateUserPwd)
		m.Label("Password Again").Id(id_user_pwd_verify).AddPassword(p.validateMatchPwd)
		m.Label("Email").Id(id_user_email).AddInput(p.validateEmail)
		m.Size(SizeTiny).Label("We will never share your email address with anyone.").AddText()
		m.Col(6)
		m.AddSpace()
		m.Id(id_sign_up).Label("Sign Up").AddButton(p.signUpListener)
	}
	m.Close()
}

// ------------------------------------------------------------------------------------

func getWidget(sess Session, id string) Widget {
	return sess.WidgetManager().Get(id)
}

func (p SignUpPage) validateUserName(s Session, widget InputWidget, value string) (string, error) {
	// It is here in the listener that we read the 'client requested' value for the widget
	// from the ajax parameters, and write it to the state.  We will validate it here.
	return ValidateUserName(value, VALIDATE_EMPTYOK)
}

func (p SignUpPage) auxValidateUserName(s Session, widgetId string, value string, flag ValidateFlag) {
	pr := PrIf(true)
	pr("auxValidateUserName")
	pr("value:", value)
	value, err := ValidateUserName(value, flag)
	pr("validated:", value, "error:", err)
	s.SetWidgetProblem(widgetId, err)
}

func (p SignUpPage) validateUserPwd(s Session, widget InputWidget, value string) (string, error) {
	return value, p.auxValidateUserPwd(s, widget, value, VALIDATE_EMPTYOK)
}

func (p SignUpPage) auxValidateUserPwd(s Session, widget Widget, value string, flag ValidateFlag) error {
	pr := PrIf(false)
	pr("auxValidateUserPwd:", value)
	value, err := ValidateUserPassword(value, flag)
	pr("afterward:", value, "err:", err)
	return err
}

func (p SignUpPage) validateMatchPwd(s Session, widget InputWidget, value string) (string, error) {
	return p.auxValidateMatchPwd(s, widget, value, VALIDATE_EMPTYOK)
}

func (p SignUpPage) auxValidateMatchPwd(s Session, widget InputWidget, value string, flag ValidateFlag) (string, error) {
	if flag.Has(VALIDATE_EMPTYOK) && value == "" {
		return value, nil
	}
	var err error
	value1 := s.State.OptString(id_user_pwd, "")
	err, value = replaceWithTestInput(err, value, "a", value1)
	if value1 != value {
		err = Ternary(value == "", ErrorEmptyUserPassword, ErrorUserPasswordsDontMatch)
	}
	return value, err
}

func (p SignUpPage) validateEmail(s Session, widget InputWidget, value string) (string, error) {
	return p.auxValidateEmail(s, widget, value, VALIDATE_EMPTYOK)
}

func (p SignUpPage) auxValidateEmail(s Session, widget Widget, value string, flag ValidateFlag) (string, error) {
	return ValidateEmailAddress(value, flag)
}

func (p SignUpPage) signUpListener(s Session, widget Widget) {
	pr := PrIf(true)
	pr("signUpListener, state:", INDENT, s.State)

	p.auxValidateUserName(s, id_user_name, s.State.OptString(id_user_name, ""), 0)
	p.auxValidateUserPwd(s, getWidget(s, id_user_pwd), s.State.OptString(id_user_pwd, ""), 0)
	p.auxValidateMatchPwd(s, getWidget(s, id_user_pwd_verify).(InputWidget), s.State.OptString(id_user_pwd_verify, ""), 0)
	p.auxValidateEmail(s, getWidget(s, id_user_email), s.State.OptString(id_user_email, ""), 0)

	b := NewUser()
	b.SetName(s.State.OptString(id_user_name, ""))
	b.SetPassword(s.State.OptString(id_user_pwd, ""))
	b.SetEmail(s.State.OptString(id_user_email, ""))

	errcount := WidgetErrorCount(s.PageWidget, s.State)
	pr("error count:", errcount)
	if errcount != 0 {
		return
	}

	Todo("don't create the user until we are sure the other fields have no errors")

	ub, err := CreateUserWithName(b)
	if ReportIfError(err, "CreateUserWithName", b) {
		return
	}
	if ub.Id() == 0 {
		s.SetWidgetProblem(id_user_name, "A user with this name already exists.")
		return
	}

	Pr("created user:", INDENT, ub)

	Todo("add support for WaitingActivation")
	Todo("if everything worked out, change the displayed page / login state?")
}
