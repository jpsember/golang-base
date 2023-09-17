package webapp

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/webserv"
)

func ProcessUserHeaderClick(sess Session, message string) bool {
	Todo("!Figure out how to automatically register click listeners (on a page basis) for things such as the user header")
	pr := PrIf(false)
	pr("UserHeaderClick? Messsage:", message)
	if _, f := TrimIfPrefix(message, HEADER_WIDGET_BUTTON_PREFIX); f {
		user := SessionUser(sess)
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
