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
		m.Label("Send Change Password Link").AddButton(p.sendLinkListener)
	}
	m.Close()
}

func (p ForgotPasswordPage) validateEmail(s Session, widget InputWidget, value string) (string, error) {
	return ValidateEmailAddress(value, 0)
}

func (p ForgotPasswordPage) sendLinkListener(sess Session, widget Widget, arg string) {
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

	user, err := ReadUserWithEmail(emailAddr)
	ReportIfError(err)

	if user.Id() == 0 {
		Alert("#50No user found for email:", emailAddr)
		sess.SetProblem(p.emailWidget, "Sorry, that email doesn't belong to any registered users.")
		return
	}
	Todo("send an email with a reset password link")

	var bodyText string
	{
		Alert("Is it safe to modify the user here?")
		b := user.ToBuilder()
		b.SetResetPasswordSecret(RandomSessionId())
		user = b.Build()
		CheckOk(UpdateUser(user))
		s := strings.Builder{}
		s.WriteString("Hello, " + user.Name() + "!\n")
		s.WriteString("Click here to <a href=\"" + ProjStructure.BaseUrl() + "/reset_password/" + user.ResetPasswordSecret() + IntToString(user.Id()) + "\">Reset password</a>")
		bodyText = s.String()
	}

	m := NewEmail().SetToAddress(user.Email()).SetSubject("Reset Password Link").SetBody(bodyText)
	SharedEmailManager().Send(m)

	sess.SwitchToPage(CheckMailPageTemplate, nil)
}
