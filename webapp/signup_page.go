package webapp

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/webapp/gen/webapp_data"
	. "github.com/jpsember/golang-base/webserv"
)

const (
	//id_user_name       = "user_name"
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
	editor DataEditor
}
type SignUpPage = *SignUpPageStruct

var SignUpPageTemplate = &SignUpPageStruct{}

func newSignUpPage(session Session) SignUpPage {
	Todo("Use editor and dataclass to hold the state, e.g. user name, pwd, pwd verify")
	t := &SignUpPageStruct{}
	t.editor = NewDataEditor(NewSignUpState())
	t.generateWidgets(session)
	return t
}

func (p SignUpPage) Name() string   { return "signup" }
func (p SignUpPage) Args() []string { return nil }
func (p SignUpPage) ConstructPage(s Session, args PageArgs) Page {
	// Use the PageArgs to verify that the construction parameters are valid.
	if args.CheckDone() {
		if OptSessionUser(s).Id() == 0 {
			return newSignUpPage(s)
		}
	}
	return nil
}

func (p SignUpPage) generateWidgets(s Session) {

	s.DeleteStateErrors()
	m := GenerateHeader(s, p)
	m.Label("Sign Up Page").Size(SizeLarge).AddHeading()

	s.WidgetManager().PushStateProvider(p.editor.WidgetStateProvider)
	m.Col(6).Open()
	{
		m.Col(12)
		m.Label("User name").Id(SignUpState_UserName).AddInput(p.validateUserName)
		m.Label("Password").Id(SignUpState_UserPwd).AddPassword(p.validateUserPwd)
		m.Label("Password Again").Id(SignUpState_UserPwdVerify).AddPassword(p.validateMatchPwd)
		m.Label("Email").Id(SignUpState_UserEmail).AddInput(p.validateEmail)
		m.Size(SizeTiny).Label("We will never share your email address with anyone.").AddText()
		m.Col(6)
		m.AddSpace()
		m.Id(id_sign_up).Label("Sign Up").AddButton(p.signUpListener)
	}
	m.Close()

	s.WidgetManager().PopStateProvider()
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

func (p SignUpPage) auxValidateUserName(s Session, widgetId string, flag ValidateFlag) {
	pr := PrIf("", true)
	value := p.editor.GetString(widgetId)
	pr("auxValidateUserName")
	pr("value:", value)
	value, err := ValidateUserName(value, flag)
	pr("validated:", value, "error:", err)
	s.SetWidgetProblem(widgetId, err)
}

func (p SignUpPage) validateUserPwd(s Session, widget InputWidget, value string) (string, error) {
	// This assumes that the widget state is stored in our editor.
	return value, p.auxValidateUserPwd(s, widget.Id(), VALIDATE_EMPTYOK)
}

func (p SignUpPage) auxValidateUserPwd(s Session, widgetId string, flag ValidateFlag) error {
	pr := PrIf("", false)
	value := p.editor.GetString(widgetId)
	pr("auxValidateUserPwd:", value)
	value, err := ValidateUserPassword(value, flag)
	pr("afterward:", value, "err:", err)
	return err
}

func (p SignUpPage) validateMatchPwd(s Session, widget InputWidget, value string) (string, error) {
	return p.auxValidateMatchPwd(s, widget.Id(), VALIDATE_EMPTYOK)
}

func (p SignUpPage) auxValidateMatchPwd(s Session, widgetId string, flag ValidateFlag) (string, error) {
	value := p.editor.GetString(widgetId)
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
	return p.auxValidateEmail(s, widget.Id(), VALIDATE_EMPTYOK)
}

func (p SignUpPage) auxValidateEmail(s Session, widgetId string, flag ValidateFlag) (string, error) {
	return ValidateEmailAddress(p.editor.GetString(widgetId), flag)
}

func (p SignUpPage) signUpListener(s Session, widget Widget, arg string) {
	pr := PrIf("", false)
	pr("signUpListener, state:", INDENT, s.State)

	p.auxValidateUserName(s, SignUpState_UserName, 0)
	p.auxValidateUserPwd(s, SignUpState_UserPwd, 0)
	p.auxValidateMatchPwd(s, SignUpState_UserPwdVerify, 0)
	p.auxValidateEmail(s, SignUpState_UserEmail, 0)

	errcount := WidgetErrorCount(s.PageWidget, s.State)
	pr("error count:", errcount)
	if errcount != 0 {
		return
	}

	Todo("don't create the user until we are sure the other fields have no errors")

	// Construct a user by parsing the signupstate map
	b := DefaultUser.Parse(p.editor.State).(User)

	//b := NewUser()
	//b.SetName(s.State.OptString(id_user_name, ""))
	//b.SetPassword(s.State.OptString(id_user_pwd, ""))
	//b.SetEmail(s.State.OptString(id_user_email, ""))

	ub, err := CreateUserWithName(b)
	if ReportIfError(err, "CreateUserWithName", b) {
		return
	}
	if ub.Id() == 0 {
		s.SetWidgetProblem(SignUpState_UserName, "A user with this name already exists.")
		return
	}

	Pr("created user:", INDENT, ub)

	Todo("add support for WaitingActivation")
	Todo("if everything worked out, change the displayed page / login state?")
}
