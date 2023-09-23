package webapp

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/webserv"
)

// This adds the webserv UserHeaderWidget, and adds our app's click listener to it.
func AddUserHeaderWidget(s Session) {
	m := s.WidgetManager()
	hw := m.AddUserHeader()
	Todo("register a click listener with the user header", hw)
	//hw.SetClickListener(ourProcessUserHeaderClick)
}

func ourProcessUserHeaderClick(sess Session, message string) bool {
	pr := PrIf(false)
	pr("UserHeaderClick? Message:", message)
	if _, f := TrimIfPrefix(message, HEADER_WIDGET_BUTTON_PREFIX); f {
		user := OptSessionUser(sess)
		if user.Id() > 0 {
			LogOut(sess)
			sess.SwitchToPage(NewLandingPage(sess))
		} else {
			sess.SwitchToPage(NewLandingPage(sess))
		}
		return true
	}
	return false
}
