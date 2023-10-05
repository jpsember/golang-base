package webapp

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/webapp/gen/webapp_data"
	. "github.com/jpsember/golang-base/webserv"
)

//  https://jeff.org/resetpassword/tGj_b_QJM7LRGDMewmnf9NbrXF27J-R69-CIQTZpFp0=

// ------------------------------------------------------------------------------------
// Page implementation
// ------------------------------------------------------------------------------------

const ResetPasswordPageName = "resetpassword"

var ResetPasswordPageTemplate = &ResetPasswordPageStruct{}

// ------------------------------------------------------------------------------------

type ResetPasswordPageStruct struct {
}

type ResetPasswordPage = *ResetPasswordPageStruct

func (p ResetPasswordPage) Name() string {
	return ResetPasswordPageName
}

func (p ResetPasswordPage) ConstructPage(s Session, args PageArgs) Page {
	pr := PrIf("ConstructPage", true)
	pr("args:", args)

	if args.Done() {
		return nil
	}
	secret := args.Next()
	if !args.CheckDone() {
		return nil
	}
	pr("looking for record with secret", secret)
	fg, err := ReadForgottenPasswordWithSecret(secret)
	if ReportIfError(err) {
		pr("error:", err)
		return nil
	}
	if fg.Id() == 0 {
		pr("no record found for secret")
		return nil
	}

	// Log the user in, delete the reset secret, and put them on the change password page
	prob := AttemptSignIn(s, fg.UserId())
	if prob != "" {
		Alert("#50trouble signing in:", prob, "user:", fg.UserId())
		return nil
	}

	Todo("Include an argument to clear the password")
	page := UserSettingsPageTemplate.ConstructPage(s, NewPageArgs(nil))
	CheckState(page != nil)
	return page
}

func (p ResetPasswordPage) Args() []string { return nil }

func (p ResetPasswordPage) generateWidgets(s Session) {
	//m := GenerateHeader(s, p)
	//AddUserHeaderWidget(s)
	//
	//m.Open()
	//{
	//	m.Col(12)
	//	p.emailWidget = m.Label("Email").Id(SignUpState_Email).AddInput(p.validateEmail)
	//}
	//m.Close()
	//m.Open()
	//{
	//	m.Label("Send Change Password Link").AddButton(p.sendLinkListener)
	//}
	//m.Close()
}

//func (p ResetPasswordPage) validateEmail(s Session, widget InputWidget, value string) (string, error) {
//	return ValidateEmailAddress(value, 0)
//}

func (p ResetPasswordPage) reportError(sess Session, widget Widget, err error, errorMessage ...any) bool {
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

//var userNotFoundError = Error("No such user")
//var internalErrMsg = "Sorry, an internal error occurred."
//
//func (p ResetPasswordPage) sendLinkListener(sess Session, widget Widget, arg string) {
//	pr := PrIf("ResetPasswordPage.sendLinkListener", false)
//
//	p.strict = true
//	errcount := sess.ValidateAndCountErrors(sess.PageWidget)
//	p.strict = false
//
//	pr("error count:", errcount)
//	if errcount != 0 {
//		return
//	}
//
//	emailAddr := sess.WidgetStringValue(p.emailWidget)
//	Todo("?Massage email address to put in lower case etc")
//
//	user, err2 := ReadUserWithEmail(emailAddr)
//	if p.reportError(sess, p.emailWidget, err2, internalErrMsg) {
//		return
//	}
//
//	if user.Id() == 0 {
//		p.reportError(sess, p.emailWidget, userNotFoundError, "Sorry, that email doesn't belong to any registered users.")
//		return
//	}
//	Todo("send an email with a reset password link")
//
//	var forgottenPassword ForgottenPassword
//	{
//		var err error
//
//		var oldRecord ForgottenPassword
//		oldRecord, err = ReadForgottenPasswordWithUserId(user.Id())
//		if p.reportError(sess, p.emailWidget, err) {
//			return
//		}
//
//		Alert("!Is it safe to call delete with a zero id?")
//		if true || oldRecord.Id() != 0 {
//			DeleteForgottenPassword(oldRecord.Id())
//		}
//
//		forgottenPassword, err = CreateForgottenPasswordWithSecret(RandomSessionId())
//		if p.reportError(sess, p.emailWidget, err, p.emailWidget) {
//			return
//		}
//
//		forgottenPassword = forgottenPassword.ToBuilder().SetUserId(user.Id()).SetCreationTimeMs(CurrentTimeMs()).Build()
//		err = UpdateForgottenPassword(forgottenPassword)
//		if p.reportError(sess, p.emailWidget, err, err) {
//			return
//		}
//	}
//
//	Todo("html email is not rendering properly")
//
//	var bodyText string
//	{
//		s := strings.Builder{}
//		//<html><head><meta charset=3D"utf-8"><title>Capital vs. labor under Biden</t=
//		//itle>
//
//		s.WriteString(`
//<html><head><title>Reset Password Link</title></head><body>
//Hello, ` + user.Name() + `!
//
//Click here to <a href="` + ProjStructure.BaseUrl() + `/reset_password/` + forgottenPassword.Secret() + `>Reset password</a>
//
//</body>
//`)
//		bodyText = s.String()
//	}
//
//	m := NewEmail().SetToAddress(user.Email()).SetSubject("Reset Password Link").SetBody(bodyText)
//	SharedEmailManager().Send(m)
//
//	sess.SwitchToPage(CheckMailPageTemplate, nil)
//}
