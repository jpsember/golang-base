package webapp

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/webserv"
)

// This adds the webserv UserHeaderWidget, and adds our app's click listener to it.
func AddUserHeaderWidget(s Session) {
	s.AddUserHeader(ourProcessUserHeaderClick)
}

func ourProcessUserHeaderClick(sess Session, widget Widget, args WidgetArgs) {
	pr := PrIf("UserHeaderClick", false)
	pr("args:", args)
	user := OptSessionUser(sess)
	valid, message := args.Read()
	if valid {
		switch message {
		case USER_HEADER_ACTION_SIGN_OUT:
			if user.Id() > 0 {
				LogOut(sess)
				sess.SwitchToPage(LandingPageTemplate, nil)
			}
			break
		case USER_HEADER_ACTION_SIGN_IN:
			if user.Id() == 0 {
				sess.SwitchToPage(LandingPageTemplate, nil)
			}
			break
		}
	}
}
