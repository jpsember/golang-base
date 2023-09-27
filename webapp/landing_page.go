package webapp

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/webapp/gen/webapp_data"
	. "github.com/jpsember/golang-base/webserv"
)

// ------------------------------------------------------------------------------------
// Page implementation
// ------------------------------------------------------------------------------------

const LandingPageName = "signin"

var LandingPageTemplate = &LandingPageStruct{}

// ------------------------------------------------------------------------------------

const (
	id_user_name       = "user_name"
	id_user_pwd_verify = "user_pwd_verify"
	id_user_pwd        = "user_pwd"
	id_user_email      = "user_email"
)

type LandingPageStruct struct {
}

type LandingPage = *LandingPageStruct

func NewLandingPage(sess Session) Page {
	t := &LandingPageStruct{}
	if sess != nil {
		t.generateWidgets(sess)
	}
	return t
}

func (p LandingPage) Name() string {
	return LandingPageName
}

func (p LandingPage) ConstructPage(s Session, args PageArgs) Page {
	if args.CheckDone() {
		user := OptSessionUser(s)
		if user.Id() == 0 {
			return NewLandingPage(s)
		}
	}
	return nil
}
func (p LandingPage) Args() []string { return nil }

type ExpStruct struct {
	BasicListStruct
	itemStates map[int]JSMap
}

type Exp = *ExpStruct

func NewExp() Exp {
	t := &ExpStruct{}
	t.ElementIds = []int{17, 42, 93, 61, 18, 29, 70}
	t.ElementsPerPage = 3
	t.itemStates = make(map[int]JSMap)
	return t
}

// ItemStateProvider constructs a state provider xi for rendering item i.
// Child widgets within the item widget that already have explicit state providers
// will *not* use Xi.
func (x Exp) ItemStateProvider(s Session, elementId int) WidgetStateProvider {
	pr := PrIf("Exp.ItemStateProvider", false)
	j := x.itemStates[elementId]
	if j == nil {
		j = NewJSMap().Put("alpha", "#"+IntToString(elementId)).Put("charlie", true)
		x.itemStates[elementId] = j
	}
	result := NewStateProvider("", j)
	Todo("!figure out if padding ints in BasePrinter is desired behavior.")
	pr("element id:", elementId, "returning:", result)
	return result
}

func (p LandingPage) generateWidgets(sess Session) {
	CheckState(sess != nil, "There is no session!")
	s := sess.State
	CheckState(s != nil, "there is no State for the session!")
	s.DeleteEach(id_user_name, id_user_pwd, id_user_pwd_verify, id_user_email)

	m := GenerateHeader(sess, p)

	if false && Alert("!Doing an experiment with lists") {

		// Construct a widget to serve as the item widget

		itemWidget := m.Open()
		{
			m.Id("alpha").Col(2).AddText()
			m.Id("abba").Col(3).AddInput(func(sess Session, widget InputWidget, value string) (string, error) {
				Pr("abba listener, id:", widget.Id(), "value:", QUO, value, "context:", sess.Context())
				return value, nil
			})

			m.Id("bravo").Col(2).Label("Hello").AddButton(func(sess Session, widget Widget, message string) {
				Pr("Landing page, button listener, id:", widget.Id(), "message:", QUO, message, "context:", sess.Context())
			})
			m.Id("charlie").Col(5).Label("Option:").AddCheckbox(
				func(sess Session, widget CheckboxWidget, state bool) (bool, error) {
					Pr("Landing page, checkbox listener, id:", widget.Id(), "state:", state, "context:", sess.Context())
					return state, nil
				})
		}
		m.Close()
		//itemWidget.SetVisible(false)

		y := m.AddList(NewExp(), itemWidget)
		y.WithPageControls = false
	}

	Todo("Refactor to eliminate id_user_xxx")
	m.Label("gallery").Align(AlignRight).Size(SizeTiny).AddButton(p.galleryListener)
	m.Col(6)
	m.Open()
	{
		m.Col(12)
		m.Label("User name").Id(id_user_name).AddInput(p.validateUserName)
		m.Label("Password").Id(id_user_pwd).AddPassword(p.validateUserPwd)
		m.Open()
		m.Col(6)
		{
			m.Label("Sign In").AddButton(p.signInListener)
			m.Label("I forgot my password").Size(SizeTiny).AddButton(p.forgotPwdListener)
		}
		m.Close()
	}
	m.Close()
	m.Open()
	{
		m.Label("Sign Up").AddButton(p.signUpListener)
	}
	m.Close()
}

func (p LandingPage) validateUserName(s Session, widget InputWidget, name string) (string, error) {
	return ValidateUserName(name, VALIDATE_EMPTYOK)
}

func (p LandingPage) validateUserPwd(s Session, widget InputWidget, content string) (string, error) {
	return ValidateUserPassword(content, VALIDATE_ONLY_NONEMPTY)
}

var AutoActivateUser = DevDatabase && Alert("?Automatically activating user")

func (p LandingPage) signInListener(sess Session, widget Widget, arg string) {
	pr := PrIf("", false)
	s := sess.State
	pr("signInListener; state:", INDENT, s)
	userName := s.OptString(id_user_name, "")
	pwd := s.OptString(id_user_pwd, "")

	var err1 error
	userName, err1 = ValidateUserName(userName, VALIDATE_ONLY_NONEMPTY)
	var err2 error
	pwd, err2 = ValidateUserPassword(pwd, VALIDATE_ONLY_NONEMPTY)

	pr("id_user_name problem:", err1)
	pr("id_user_pwd  problem:", err2)
	sess.SetWidgetProblem(id_user_name, err1)
	sess.SetWidgetProblem(id_user_pwd, err2)

	var user = DefaultUser
	prob := ""
	for {
		errcount := WidgetErrorCount(sess.PageWidget, sess.State)
		pr("errcount:", errcount)
		if errcount != 0 {
			break
		}

		var err error
		user, err = ReadUserWithName(userName)
		ReportIfError(err)
		userId := user.Id()

		prob = AttemptSignIn(sess, userId)
		break
	}
	pr("problem is:", prob)
	if prob != "" {
		sess.SetWidgetProblem(id_user_name, prob)
	}
}

func (p LandingPage) signUpListener(s Session, widget Widget, arg string) {
	s.SwitchToPage(SignUpPageTemplate, nil)
}

func (p LandingPage) galleryListener(sess Session, widget Widget, arg string) {
	sess.SwitchToPage(GalleryPageTemplate, nil)
}

func (p LandingPage) forgotPwdListener(sess Session, widget Widget, arg string) {

	for {

		userEmail := sess.StringValue(id_user_email)

		if userEmail == "" {
			sess.SetWidgetProblem(id_user_email, "Please enter your email address.")
			return
		}

		sess.SetWidgetProblem(id_user_name, "An email has been sent with a link to change your password.")

		//user, err := webapp_data.ReadUserWithName(userName)
		//userId := user.Id()
		//
		//if err != nil {
		//	Alert("Not revealing that 'no such user exists' in forgot password logic")
		//}
		//if userId != 0 {
		//	Todo("Send email")
		//}
		//break
	}

}
