package webapp

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/webapp/gen/webapp_data"
	. "github.com/jpsember/golang-base/webserv"
)

// ------------------------------------------------------------------------------------
// This is the canonical boilerplate that I will turn into a goland live template,
// to simplify creating new pages:
// ------------------------------------------------------------------------------------

type SignUpPageStruct struct {
	editor DataEditor
	strict bool
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
		m.Label("User name").Id(SignUpState_Name).AddInput(p.listenerValidateName)
		m.Label("Password").Id(SignUpState_Password).AddPassword(p.listenerValidatePwd)
		m.Label("Password Again").Id(SignUpState_PasswordVerify).AddPassword(p.listenerValidatePwdVerify)
		m.Label("Email").Id(SignUpState_Email).AddInput(p.validateEmail)
		m.Size(SizeTiny).Label("We will never share your email address with anyone.").AddText()
		m.Col(6)
		m.AddSpace()
		m.Label("Sign Up").AddButton(p.signUpListener)
	}
	m.Close()

	s.WidgetManager().PopStateProvider()
}

// ------------------------------------------------------------------------------------

func getWidget(sess Session, id string) Widget {
	return sess.WidgetManager().Get(id)
}

func (p SignUpPage) listenerValidateName(s Session, widget InputWidget, value string) (string, error) {
	// It is here in the listener that we read the 'client requested' value for the widget
	// from the ajax parameters, and write it to the state.  We will validate it here.
	return ValidateUserName(value, Ternary(p.strict, 0, VALIDATE_EMPTYOK))
}

func (p SignUpPage) auxValidateUserName(s Session, widgetId string, flag ValidateFlag) (string, error) {
	pr := PrIf("auxValidateUserName", true)
	value := p.editor.GetString(widgetId)
	pr("value:", value)
	value, err := ValidateUserName(value, flag)
	pr("validated:", value, "error:", err)
	//	s.UpdateValueAndProblemId(widgetId, value, err)
	return value, err
}

func (p SignUpPage) listenerValidatePwd(s Session, widget InputWidget, value string) (string, error) {
	pr := PrIf("listenerValidatePwd", true)
	// This assumes that the widget state is stored in our editor.
	flag := Ternary(p.strict, 0, VALIDATE_EMPTYOK)
	//widgetId := widget.Id()
	pr(VERT_SP, "Validating, value:", QUO, value)
	//
	//	return p.auxValidateUserPwd(s, widget.Id(), Ternary(p.strict, 0, VALIDATE_EMPTYOK))
	//}
	//
	//func (p SignUpPage) auxValidateUserPwd(s Session, widgetId string, flag ValidateFlag) (string, error) {
	//edValue := p.editor.GetString(widgetId)
	//pr("editor.GetString:", QUO, edValue)
	//pr("widgetId:", widgetId, "pwd:", value)
	validated, err := ValidateUserPassword(value, flag)
	pr("after validating:", validated, "err:", err)
	//	s.UpdateValueAndProblemId(widgetId, value, err)
	return validated, err
}

func (p SignUpPage) listenerValidatePwdVerify(s Session, widget InputWidget, value string) (string, error) {
	return p.auxValidateMatchPwd(s, widget.Id(), Ternary(p.strict, 0, VALIDATE_EMPTYOK))
}

func (p SignUpPage) auxValidateMatchPwd(s Session, widgetId string, flag ValidateFlag) (string, error) {
	var err error
	pr := PrIf("auxValidateMatchPwd", true)
	pr("widgetId:", widgetId, "flag:", flag)
	value := p.editor.GetString(widgetId)
	pr("flag.Has(VALIDATE_EMPTYOK):", flag.Has(VALIDATE_EMPTYOK))
	if flag.Has(VALIDATE_EMPTYOK) && value == "" {
	} else {
		value1 := p.editor.GetString(SignUpState_Password)
		err, value = replaceWithTestInput(err, value, "a", value1)
		if value1 != value {
			err = Ternary(value == "", ErrorEmptyUserPassword, ErrorUserPasswordsDontMatch)
		}
		pr("returning:", QUO, value, "err:", err)
	}
	//	if flag.Has(VALIDATE_UPDATE_WIDGETS) {
	//		s.UpdateValueAndProblemId(widgetId, value, err)
	//	}
	//s.SetWidgetProblem(widgetId, err)
	return value, err
}

func (p SignUpPage) validateEmail(s Session, widget InputWidget, value string) (string, error) {
	return p.auxValidateEmail(s, widget.Id(), Ternary(p.strict, 0, VALIDATE_EMPTYOK))
}

func (p SignUpPage) auxValidateEmail(s Session, widgetId string, flag ValidateFlag) (string, error) {
	Todo("would be simpler to pass in the widget, not the widget id")
	val, err := ValidateEmailAddress(p.editor.GetString(widgetId), flag)
	//	if flag.Has(VALIDATE_UPDATE_WIDGETS) {
	//		s.UpdateValueAndProblemId(widgetId, val, err)
	//	}
	return val, err
}

func (p SignUpPage) signUpListener(s Session, widget Widget, arg string) {
	pr := PrIf("signupListener", true)
	pr("state:", INDENT, p.editor.State)

	// We need to basically call all the same validators that we do in the callbacks,
	// and we have to update the widget values and errors ourselves (something the callback handler
	// does automatically).

	// Better approach: set a flag that says we are doing 'strict' validation
	p.strict = true
	s.Validate(s.PageWidget)

	//Todo("Have a push state thing to set VALIDATE_UPDATE_WIDGETS here?")
	//p.auxValidateUserName(s, SignUpState_Name, VALIDATE_UPDATE_WIDGETS)
	//p.auxValidateUserPwd(s, SignUpState_Password, VALIDATE_UPDATE_WIDGETS)
	//p.auxValidateMatchPwd(s, SignUpState_PasswordVerify, VALIDATE_UPDATE_WIDGETS)
	//p.auxValidateEmail(s, SignUpState_Email, VALIDATE_UPDATE_WIDGETS)
	p.strict = false
	errcount := WidgetErrorCount(s.PageWidget, s.State)
	Todo("Maybe have Validate(...) return the error count?")
	pr("error count:", errcount)
	if errcount != 0 {
		return
	}

	Todo("don't create the user until we are sure the other fields have no errors")

	// Construct a user by parsing the signupstate map
	b := DefaultUser.Parse(p.editor.State).(User).ToBuilder()
	b.SetUserClass(UserClassDonor)

	//b := NewUser()
	//b.SetName(s.State.OptString(id_user_name, ""))
	//b.SetPassword(s.State.OptString(id_user_pwd, ""))
	//b.SetEmail(s.State.OptString(id_user_email, ""))

	ub, err := CreateUserWithName(b)
	if ReportIfError(err, "CreateUserWithName", b) {
		return
	}
	if ub.Id() == 0 {
		s.SetWidgetProblem(SignUpState_Name, "A user with this name already exists.")
		return
	}

	Pr("created user:", INDENT, ub)

	Todo("add support for WaitingActivation")
	AttemptSignIn(s, ub.Id())
}
