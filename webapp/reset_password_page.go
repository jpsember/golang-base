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
	alert AlertWidget
}

type ResetPasswordPage = *ResetPasswordPageStruct

func (p ResetPasswordPage) Name() string {
	return ResetPasswordPageName
}

func (p ResetPasswordPage) ConstructPage(s Session, args PageArgs) Page {
	pr := PrIf("ResetPasswordPage.ConstructPage", true)
	pr("args:", args)

	Todo("!Have an alert page that returns to sign in, or home page")

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

	page := UserSettingsPageTemplate.ConstructPage(s, PageArgsWith("resetpwd"))
	CheckState(page != nil)
	return page
}

func (p ResetPasswordPage) Args() []string { return nil }

func (p ResetPasswordPage) generateWidgets(s Session) {
	m := GenerateHeader(s, p)

	m.Open()

	Todo("Have widgetManager support for Alerts")
	p.alert = NewAlertWidget("info", AlertInfo)
	//alertWidget.SetVisible(false)
	m.Add(p.alert)

	m.Close()
}

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
