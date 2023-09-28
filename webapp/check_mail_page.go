package webapp

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/webserv"
)

var CheckMailPageTemplate = &CheckMailPageStruct{}

type CheckMailPageStruct struct{}

type CheckMailPage = *CheckMailPageStruct

func (p CheckMailPage) Name() string {
	_ = Pr
	return "checkmail"
}

func (p CheckMailPage) ConstructPage(s Session, args PageArgs) Page {
	if args.CheckDone() {
		user := OptSessionUser(s)
		if user.Id() == 0 {
			return newCheckMailPage(s)
		}
	}
	return nil
}

func (p CheckMailPage) Args() []string { return nil }

func newCheckMailPage(s Session) Page {
	p := &CheckMailPageStruct{}
	p.generateWidgets(s)
	return p
}

func (p CheckMailPage) generateWidgets(s Session) {
	m := GenerateHeader(s, p)
	AddUserHeaderWidget(s)
	m.Open()
	{
		m.Label("Check your email for a link to reset your password.").Size(SizeLarge).AddText()
	}
	m.Close()
}
