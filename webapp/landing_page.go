package webapp

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/webapp/gen/webapp_data"
	. "github.com/jpsember/golang-base/webserv"
)

// ------------------------------------------------------------------------------------
// Page implementation
// ------------------------------------------------------------------------------------

const LandingPageName = "signin"

var LandingPageTemplate = &LandingPageStruct{}

// ------------------------------------------------------------------------------------

type LandingPageStruct struct {
	editor         DataEditor
	strict         bool
	nameWidget     Widget
	passwordWidget Widget
}

type LandingPage = *LandingPageStruct

func NewLandingPage(sess Session) Page {
	t := &LandingPageStruct{}
	if sess != nil {
		t.editor = NewDataEditor(NewLandingState())
		t.generateWidgets(sess)
	}
	return t
}

func (p LandingPage) Name() string {
	return LandingPageName
}

func (p LandingPage) ConstructPage(s Session, args PageArgs) Page {
	if args.CheckDone() {
		user := OptSessionUser(s)
		if user.Id() == 0 {
			return NewLandingPage(s)
		}
	}
	return nil
}
func (p LandingPage) Args() []string { return nil }

func (p LandingPage) generateWidgets(sess Session) {
	m := GenerateHeader(sess, p)

	m.Label("gallery").Align(AlignRight).Size(SizeTiny).AddButton(p.galleryListener)
	m.Col(6)
	m.Open()
	{
		m.Col(12)
		p.nameWidget = m.Label("User name").Id(LandingState_Name).AddInput(p.validateUserName)
		Alert("validateUserName is NOT being called if the value hasn't changed")
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

var AutoActivateUser = DevDatabase && Alert("?Automatically activating user")

func (p LandingPage) signInListener(sess Session, widget Widget, arg string) {
	pr := PrIf("LandingPage.signInListener", true)
	pr("state:", INDENT, p.editor.State)

	// Re-validate all the widgets in 'strict' mode.
	p.strict = true
	errcount := sess.ValidateAndCountErrors(sess.PageWidget)
	p.strict = false

	pr("error count:", errcount)
	if errcount != 0 {
		return
	}

	var user = DefaultUser
	prob := ""
	for {
		//errcount := sess.WidgetErrorCount(sess.PageWidget)
		//pr("errcount:", errcount)
		//if errcount != 0 {
		//	break
		//}
		userName := sess.WidgetStringValue(p.nameWidget)
		var err error
		user, err = ReadUserWithName(userName)
		ReportIfError(err)
		userId := user.Id()

		Todo("verify password match", p.passwordWidget)

		prob = AttemptSignIn(sess, userId)
		break
	}
	pr("problem is:", prob)
	if prob != "" {
		sess.SetProblem(p.nameWidget, prob)
	}
}

func (p LandingPage) signUpListener(s Session, widget Widget, arg string) {
	s.SwitchToPage(SignUpPageTemplate, nil)
}

func (p LandingPage) galleryListener(sess Session, widget Widget, arg string) {
	sess.SwitchToPage(GalleryPageTemplate, nil)
}

func (p LandingPage) forgotPwdListener(sess Session, widget Widget, arg string) {

	Todo("Where is the email widget?")
	//for {
	//
	//	w := sess.Widget(id_user_email)
	//	userEmail := sess.WidgetStringValue(w)
	//
	//	if userEmail == "" {
	//		sess.SetWidgetProblem(id_user_email, "Please enter your email address.")
	//		return
	//	}
	//
	//	sess.SetWidgetProblem(id_user_name, "An email has been sent with a link to change your password.")
	//
	//	//user, err := webapp_data.ReadUserWithName(userName)
	//	//userId := user.Id()
	//	//
	//	//if err != nil {
	//	//	Alert("Not revealing that 'no such user exists' in forgot password logic")
	//	//}
	//	//if userId != 0 {
	//	//	Todo("Send email")
	//	//}
	//	//break
	//}

}
