package webapp

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/webserv"
)

func CreateLandingPage(sess Session) {

	m := sess.WidgetManager()

	m.Col(12)
	m.Label("Landing Page").Size(SizeLarge).AddHeading()
	m.Col(6).Open()
	{
		m.Col(12)
		m.Label("User name").Id("user_name").Listener(userNameListener).AddInput()
		Todo("!Option for password version of input field")
		m.Label("Password").Id("user_pwd").Listener(userPwdListener).AddInput()
		m.Col(6)
		m.AddSpace()
		m.Listener(signInListener)
		m.Id("sign_in").Label("Sign In").AddButton()
	}
	m.Close()

}

func getWidget(sess Session, id string) Widget {
	return sess.WidgetManager().Get(id)
}

func userNameListener(sess any, widget Widget) error {
	pr := PrIf(true)
	pr("userNameListener", WidgetId(widget))
	s := sess.(Session)

	//Todo("some redundancy here, as the id and value are found in the ajax args...")
	//wid := s.GetWidgetId()

	// It is here in the listener that we read the 'client requested' value for the widget
	// from the ajax parameters, and write it to the state.  We could send it through a validation here...

	// Maybe a two-stage validation, one that allows empty fields?

	value := s.GetValueString()
	pr("value:", value)
	value, err := ValidateUserName(value, true)
	pr("validated:", value, "error:", err)

	Todo("Utility function for the following boilerplate?")
	// We want to update the state even if the name is illegal, so user can see what he typed in
	s.State.Put(WidgetId(widget), value)

	if err != nil {
		s.SetWidgetProblem(widget, err.Error())
		Todo("Must repaint after problem")
		s.Repaint(widget)
	} else {
		s.ClearWidgetProblem(widget)
		s.Repaint(widget)
	}
	return err
}

func userPwdListener(sess any, widget Widget) error {
	Pr("userPwdListener", WidgetId(widget))
	s := sess.(Session)
	wid := s.GetWidgetId()
	s.State.Put(wid, s.GetValueString())
	Todo("if clearing the problem, it should repaint")
	s.ClearWidgetProblem(widget)
	s.Repaint(widget)
	return nil
}

func signInListener(sess any, widget Widget) error {

	s := sess.(Session)

	pr := PrIf(true)

	pr("state:", INDENT, s.State)

	browserUserName := getWidget(s, "user_name")
	browserPassword := getWidget(s, "user_pwd")

	Todo("have utility method to read widget value from state")

	userName := s.State.OptString("user_name", "")
	pwd := s.State.OptString("user_pwd", "")

	_, err := ValidateUserName(userName, false)
	if err != nil {
		s.SetWidgetProblem(browserUserName, err.Error())
	} else {
		s.ClearWidgetProblem(browserUserName)
	}

	s.ClearWidgetProblem(browserPassword)
	//if userName == "" {
	//	s.SetWidgetProblem(browserUserName, "Please enter your name")
	//	s.Repaint(browserUserName)
	//}
	if pwd == "" {
		s.SetWidgetProblem(browserPassword, "Please enter your password")
		s.Repaint(browserPassword)
	}
	Todo("if everything worked out, change the displayed page / login state?")
	return nil
}
