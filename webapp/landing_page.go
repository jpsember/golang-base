package webapp

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/webapp/gen/webapp_data"
	. "github.com/jpsember/golang-base/webserv"
)

var LandingPageTemplate = &LandingPageStruct{}

type LandingPageStruct struct {
	editor         DataEditor
	strict         bool
	nameWidget     Widget
	passwordWidget Widget
}

type LandingPage = *LandingPageStruct

func (p LandingPage) Name() string {
	return "signin"
}

func (p LandingPage) ConstructPage(s Session, args PageArgs) Page {
	if args.CheckDone() {
		user := OptSessionUser(s)
		if user.Id() == 0 {
			return newLandingPage(s)
		}
	}
	return nil
}

func (p LandingPage) Args() []string { return nil }

func newLandingPage(s Session) Page {
	t := &LandingPageStruct{}
	if s != nil {
		t.editor = NewDataEditor(NewLandingState())
		t.generateWidgets(s)
	}
	return t
}

func (p LandingPage) generateWidgets(sess Session) {
	m := GenerateHeader(sess, p)

	m.Label("gallery").Align(AlignRight).Size(SizeTiny).AddButton(p.galleryListener)
	m.Col(6)
	m.Open()
	{
		m.Col(12)
		p.nameWidget = m.Label("User name").Id(LandingState_Name).AddInput(p.validateUserName)
		p.passwordWidget = m.Label("Password").Id(LandingState_Password).AddPassword(p.validateUserPwd)
		m.Open()
		m.Col(6)
		{
			m.Label("Sign In").AddButton(p.signInListener)
			m.Label("I forgot my password").Size(SizeTiny).AddButton(p.forgotPwdListener)
		}
		m.Close()
	}
	m.Close()
	m.Open()
	{
		m.Label("Sign Up").AddButton(p.signUpListener)
	}
	m.Close()
}

func (p LandingPage) validateUserName(s Session, widget InputWidget, name string) (string, error) {
	Pr("validate user name for session")
	return ValidateUserName(name, p.validateFlag())
}

func (p LandingPage) validateUserPwd(s Session, widget InputWidget, content string) (string, error) {
	return ValidateUserPassword(content, p.validateFlag()|VALIDATE_ONLY_NONEMPTY)
}

func (p LandingPage) validateFlag() ValidateFlag {
	return Ternary(p.strict, 0, VALIDATE_EMPTYOK)
}

func (p LandingPage) signInListener(s Session, widget Widget, arg string) {
	pr := PrIf("LandingPage.signInListener", false)
	pr("state:", INDENT, p.editor.State)

	// Re-validate all the widgets in 'strict' mode.
	p.strict = true
	errcount := s.ValidateAndCountErrors(s.PageWidget)
	p.strict = false

	pr("error count:", errcount)
	if errcount != 0 {
		return
	}

	var user = DefaultUser
	prob := ""
	for {
		userName := s.WidgetStringValue(p.nameWidget)
		var err error
		user, err = ReadUserWithName(userName)
		ReportIfError(err)
		userId := user.Id()

		Todo("!verify the password matches the widget", p.passwordWidget)

		prob = AttemptSignIn(s, userId)
		break
	}
	pr("problem is:", prob)
	if prob != "" {
		s.SetProblem(p.nameWidget, prob)
	}
}

func (p LandingPage) signUpListener(s Session, widget Widget, arg string) {
	s.SwitchToPage(SignUpPageTemplate, nil)
}

func (p LandingPage) galleryListener(s Session, widget Widget, arg string) {
	s.SwitchToPage(GalleryPageTemplate, nil)
}

func (p LandingPage) forgotPwdListener(s Session, widget Widget, arg string) {
	s.SwitchToPage(ForgotPasswordPageTemplate, nil)
}
