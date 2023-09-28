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
	m := GenerateHeader(s, p)
	m.Label("Sign Up Page").Size(SizeLarge).AddHeading()

	s.PushStateProvider(p.editor.WidgetStateProvider)
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

	s.PopStateProvider()
}

// ------------------------------------------------------------------------------------

func (p SignUpPage) validateFlag() ValidateFlag {
	return Ternary(p.strict, 0, VALIDATE_EMPTYOK)
}

func (p SignUpPage) listenerValidateName(s Session, widget InputWidget, value string) (string, error) {
	return ValidateUserName(value, p.validateFlag())
}

func (p SignUpPage) listenerValidatePwd(s Session, widget InputWidget, value string) (string, error) {
	pr := PrIf(">listenerValidatePwd", false)
	flag := p.validateFlag()
	pr("Validating, value:", QUO, value)
	validated, err := ValidateUserPassword(value, flag)
	pr("after validating:", validated, "err:", err)
	if !p.strict {
		// We must DELAY this additional validation until after the current validation has completed,
		// else the most recent password value isn't the one we will read

		// If this becomes a common idiom, we will add a function s.PostValidate(...)
		s.AddPostRequestEvent(func() { s.Validate(s.Get(SignUpState_PasswordVerify)) })
	}
	return validated, err
}

func (p SignUpPage) listenerValidatePwdVerify(s Session, widget InputWidget, value string) (string, error) {
	pr := PrIf(">listenerValidatePwdVerify", false)
	pr("verify value  :", QUO, value)
	var err error
	flag := p.validateFlag()
	if flag.Has(VALIDATE_EMPTYOK) && value == "" {
	} else {
		value1 := p.editor.GetString(SignUpState_Password)
		pr("password value:", QUO, value1)
		pr("editor state:", INDENT, p.editor.State)
		err, value = replaceWithTestInput(err, value, "a", value1)
		if value1 != value {
			err = Ternary(value == "", ErrorEmptyUserPassword, ErrorUserPasswordsDontMatch)
		}
	}
	pr("returning:", QUO, value, "err:", err)
	return value, err
}

func (p SignUpPage) validateEmail(s Session, widget InputWidget, value string) (string, error) {
	return ValidateEmailAddress(value, p.validateFlag())
}

func (p SignUpPage) signUpListener(s Session, widget Widget, arg string) {
	pr := PrIf("signupListener", false)
	pr("state:", INDENT, p.editor.State)

	// Re-validate all the widgets in 'strict' mode.
	p.strict = true
	errcount := s.ValidateAndCountErrors(s.PageWidget)
	p.strict = false

	pr("after validating page;")
	pr("state:", INDENT, p.editor.State)

	pr("error count:", errcount)
	if errcount != 0 {
		return
	}

	Todo("don't create the user until we are sure the other fields have no errors")

	// Construct a user by parsing the signupstate map
	b := DefaultUser.Parse(p.editor.State).(User).ToBuilder()
	b.SetUserClass(UserClassDonor)

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
