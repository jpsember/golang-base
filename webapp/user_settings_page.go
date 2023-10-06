package webapp

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/webapp/gen/webapp_data"
	. "github.com/jpsember/golang-base/webserv"
)

// There are a lot of common elements with this page and the create_user page.

type UserSettingsPageStruct struct {
	editor       DataEditor
	strict       bool
	user         User
	resetPwdFlag bool
	pwdVerify    Widget
}

type UserSettingsPage = *UserSettingsPageStruct

var UserSettingsPageTemplate = &UserSettingsPageStruct{}

func (p UserSettingsPage) prepare(session Session) {
	p.user = OptSessionUser(session)

	b := p.user.ToBuilder()
	if p.resetPwdFlag {
		b.SetPassword("")
	}
	p.editor = NewDataEditor(b)
	p.generateWidgets(session)
}

func (p UserSettingsPage) Name() string { return "usersettings" }

func (p UserSettingsPage) Args() []string { return nil }

func (p UserSettingsPage) ConstructPage(s Session, args PageArgs) Page {
	pr := PrIf("user_settings, ConstructPage", true)
	pr("args:", args)
	t := &UserSettingsPageStruct{}
	t.resetPwdFlag = args.ReadIf("resetpwd")
	if args.CheckDone() {
		t.prepare(s)
		return t
	}
	return nil
}

func (p UserSettingsPage) generateWidgets(s Session) {
	m := GenerateHeader(s, p)
	m.Label("Sign Up Page").Size(SizeLarge).AddHeading()

	m.Col(6).Open()
	{
		m.Col(12)
		{
			s.PushStateProvider(p.editor.WidgetStateProvider)
			Todo("?How do I add static text?  I.e., non-editable text field?")
			m.Label(p.user.Name()).AddText()
			m.Label("Password").Id(User_Password).AddPassword(p.listenerValidatePwd)

			// The verify password isn't part of the User object
			s.PushStateProvider(NewStateProvider("", NewJSMap()))
			// We supply a specific id just for ease of debugging
			p.pwdVerify = m.Label("Password Again").Id("pwd_verify").AddPassword(p.listenerValidatePwdVerify)
			s.PopStateProvider()

      // Removing some elements to concentrate on the problem
			if false {
				m.Label("Email").Id(User_Email).AddInput(p.validateEmail)
				m.Size(SizeTiny).Label("We will never share your email address with anyone.").AddText()
			}
			s.PopStateProvider()
		}
		m.Col(6)
		m.AddSpace()
		if false {
			m.Label("Ok").AddButton(p.okListener)
		}
	}
	m.Close()

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
		s.AddPostRequestEvent(func() { s.Validate(p.pwdVerify) })
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
		value1 := p.editor.GetString(User_Password)
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

	// Construct a user by parsing the signupstate map
	//b := DefaultUser.Parse(p.editor.State).(User).ToBuilder()
	b := p.editor.Read().(User).ToBuilder()

	hash, salt := HashPassword(b.Password())
	b.SetPasswordHash(hash)
	b.SetPasswordSalt(salt)

	pr("updated user:", INDENT, b)
	p.user = b.Build()
	UpdateUser(p.user)

	s.SwitchToPage(DefaultPageForUser(p.user), nil)
}
