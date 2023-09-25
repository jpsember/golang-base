package webapp

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/webserv"
)

// This adds the webserv UserHeaderWidget, and adds our app's click listener to it.
func AddUserHeaderWidget(s Session) {
	m := s.WidgetManager()
	m.AddUserHeader(ourProcessUserHeaderClick)
}

func ourProcessUserHeaderClick(sess Session, widget Widget, message string) {
	pr := PrIf("UserHeaderClick", false)
	pr("message:", message)
	user := OptSessionUser(sess)
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
