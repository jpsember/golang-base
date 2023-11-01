package webapp

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/webapp/gen/webapp_data"
	. "github.com/jpsember/golang-base/webserv"
	. "github.com/jpsember/golang-base/webserv/gen/webserv_data"
	"strings"
)

// ------------------------------------------------------------------------------------
// Page implementation
// ------------------------------------------------------------------------------------

const ForgotPasswordPageName = "forgotpwd"

var ForgotPasswordPageTemplate = &ForgotPasswordPageStruct{}

// ------------------------------------------------------------------------------------

type ForgotPasswordPageStruct struct {
	editor      DataEditor
	strict      bool
	emailWidget Widget
}

type ForgotPasswordPage = *ForgotPasswordPageStruct

func NewForgotPasswordPage(sess Session) Page {
	t := &ForgotPasswordPageStruct{}
	if sess != nil {
		// We can use the SignUpState in this page's editor.
		t.editor = NewDataEditor(NewSignUpState())
		t.generateWidgets(sess)
	}
	return t
}

func (p ForgotPasswordPage) Name() string {
	return ForgotPasswordPageName
}

func (p ForgotPasswordPage) ConstructPage(s Session, args PageArgs) Page {
	if args.CheckDone() {
		user := OptSessionUser(s)
		if user.Id() == 0 {
			return NewForgotPasswordPage(s)
		}
	}
	return nil
}

func (p ForgotPasswordPage) Args() []string { return nil }

func (p ForgotPasswordPage) generateWidgets(s Session) {
	m := GenerateHeader(s, p)
	AddUserHeaderWidget(s)

	m.Open()
	{
		m.Col(12)
		p.emailWidget = m.Label("Email").Id(SignUpState_Email).AddInput(p.validateEmail)
	}
	m.Close()
	m.Open()
	{
		m.Label("Send Change Password Link").Listener(p.sendLinkListener).AddBtn()
	}
	m.Close()
}

func (p ForgotPasswordPage) validateEmail(s Session, widget InputWidget, value string) (string, error) {
	return ValidateEmailAddress(value, 0)
}

func (p ForgotPasswordPage) reportError(sess Session, widget Widget, err error, errorMessage ...any) bool {
	if err == nil {
		return false
	}
	msg := internalErrMsg
	if len(errorMessage) != 0 {
		msg = ToString(errorMessage)
	}
	sess.SetProblem(widget, msg)
	return true
}

var userNotFoundError = Error("No such user")
var internalErrMsg = "Sorry, an internal error occurred."

func (p ForgotPasswordPage) sendLinkListener(sess Session, widget Widget, args WidgetArgs) {
	pr := PrIf("ForgotPasswordPage.sendLinkListener", false)

	p.strict = true
	errcount := sess.ValidateAndCountErrors(sess.PageWidget)
	p.strict = false

	pr("error count:", errcount)
	if errcount != 0 {
		return
	}

	emailAddr := sess.WidgetStringValue(p.emailWidget)
	Todo("?Massage email address to put in lower case etc")

	user, err2 := ReadUserWithEmail(emailAddr)
	if p.reportError(sess, p.emailWidget, err2, internalErrMsg) {
		return
	}

	if user.Id() == 0 {
		p.reportError(sess, p.emailWidget, userNotFoundError, "Sorry, that email doesn't belong to any registered users.")
		return
	}

	var forgottenPassword ForgottenPassword
	{
		var err error

		var oldRecord ForgottenPassword
		oldRecord, err = ReadForgottenPasswordWithUserId(user.Id())
		if p.reportError(sess, p.emailWidget, err) {
			return
		}

		Alert("!Is it safe to call delete with a zero id?")
		if true || oldRecord.Id() != 0 {
			DeleteForgottenPassword(oldRecord.Id())
		}

		forgottenPassword, err = CreateForgottenPasswordWithSecret(RandomSessionId())
		if p.reportError(sess, p.emailWidget, err, p.emailWidget) {
			return
		}

		forgottenPassword = forgottenPassword.ToBuilder().SetUserId(user.Id()).SetCreationTimeMs(CurrentTimeMs()).Build()
		err = UpdateForgottenPassword(forgottenPassword)
		if p.reportError(sess, p.emailWidget, err, err) {
			return
		}
	}

	var bodyText string
	{
		s := strings.Builder{}
		//<html><head><meta charset=3D"utf-8"><title>Capital vs. labor under Biden</t=
		//itle>

		s.WriteString(`
<html>
<body>
<p>
Hello, ` + user.Name() + `!
</p>

<p>
Click here to <a href="` + ProjStructure.BaseUrl() + `/resetpassword/` + forgottenPassword.Secret() + `">reset your password</a>.
</p>

</body>
</html>
`)
		bodyText = s.String()
	}

	m := NewEmail().SetToAddress(user.Email()).SetHtml(true).SetSubject("Reset Password Link").SetBody(bodyText)
	SharedEmailManager().Send(m)

	sess.SwitchToPage(CheckMailPageTemplate, nil)
}
