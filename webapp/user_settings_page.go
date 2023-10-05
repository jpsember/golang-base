package webapp

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/webapp/gen/webapp_data"
	. "github.com/jpsember/golang-base/webserv"
)

// There are a lot of common elements with this page and the create_user page.

type UserSettingsPageStruct struct {
	editor DataEditor
	strict bool
	user   User
}

type UserSettingsPage = *UserSettingsPageStruct

var UserSettingsPageTemplate = &UserSettingsPageStruct{}

func newUserSettingsPage(session Session) UserSettingsPage {
	t := &UserSettingsPageStruct{}
	t.user = OptSessionUser(session)

	// Set the editor to the current user info

	m := t.user.ToJson().AsJSMap()
	ss := DefaultSignUpState.Parse(m).(SignUpState).ToBuilder()
	Todo("only clear the password if args told us to")
	ss.SetPassword("").SetPasswordVerify("")
	t.editor = NewDataEditor(ss)

	t.generateWidgets(session)
	return t
}

func (p UserSettingsPage) Name() string   { return "usersettings" }
func (p UserSettingsPage) Args() []string { return nil }
func (p UserSettingsPage) ConstructPage(s Session, args PageArgs) Page {
	if args.CheckDone() {
		return newUserSettingsPage(s)
	}
	return nil
}

func (p UserSettingsPage) generateWidgets(s Session) {
	m := GenerateHeader(s, p)
	m.Label("Sign Up Page").Size(SizeLarge).AddHeading()

	s.PushStateProvider(p.editor.WidgetStateProvider)
	m.Col(6).Open()
	{
		m.Col(12)
		Todo("How do I add static text?  I.e., non-editable text field?")
		m.Label(p.user.Name()).AddText()
		m.Label("Password").Id(SignUpState_Password).AddPassword(p.listenerValidatePwd)
		m.Label("Password Again").Id(SignUpState_PasswordVerify).AddPassword(p.listenerValidatePwdVerify)
		m.Label("Email").Id(SignUpState_Email).AddInput(p.validateEmail)
		m.Size(SizeTiny).Label("We will never share your email address with anyone.").AddText()
		m.Col(6)
		m.AddSpace()
		m.Label("Ok").AddButton(p.okListener)
	}
	m.Close()

	s.PopStateProvider()
}

// ------------------------------------------------------------------------------------

func (p UserSettingsPage) validateFlag() ValidateFlag {
	return Ternary(p.strict, 0, VALIDATE_EMPTYOK)
}

func (p UserSettingsPage) listenerValidatePwd(s Session, widget InputWidget, value string) (string, error) {
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

func (p UserSettingsPage) listenerValidatePwdVerify(s Session, widget InputWidget, value string) (string, error) {
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

func (p UserSettingsPage) validateEmail(s Session, widget InputWidget, value string) (string, error) {
	return ValidateEmailAddress(value, p.validateFlag())
}

func (p UserSettingsPage) okListener(s Session, widget Widget, arg string) {
	pr := PrIf("okListener", true)
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

	Die("get existing user, and modify that")
	// Construct a user by parsing the signupstate map
	b := DefaultUser.Parse(p.editor.State).(User).ToBuilder()

	hash, salt := HashPassword(b.Password())
	b.SetPasswordHash(hash)
	b.SetPasswordSalt(salt)

	//ub, err := p.attemptCreateUniqueUser(b)
	//if err != nil {
	//	switch err {
	//	case nameExists:
	//		s.SetWidgetProblem(SignUpState_Name, err.Error())
	//	case emailExists:
	//		s.SetWidgetProblem(SignUpState_Email, err.Error())
	//	default:
	//		s.SetWidgetProblem(SignUpState_Name, "Sorry, an error occurred.")
	//	}
	//	return
	//}

	//Pr("created user:", INDENT, ub)

	Todo("add support for WaitingActivation")
	//AttemptSignIn(s, ub.Id())
}

//func (p UserSettingsPage) attemptCreateUniqueUser(b UserBuilder) (User, error) {
//	CreateUserLock.Lock()
//	defer CreateUserLock.Unlock()
//
//	var existing User
//
//	existing, _ = ReadUserWithName(b.Name())
//	if existing.Id() != 0 {
//		return nil, nameExists
//	}
//	existing, _ = ReadUserWithEmail(b.Email())
//	if existing.Id() != 0 {
//		return nil, emailExists
//	}
//
//	user, err := CreateUserWithName(b.Name())
//	if err != nil {
//		return nil, err
//	}
//	if user.Id() == 0 {
//		return nil, nameExists
//	}
//	return user, nil
//}
