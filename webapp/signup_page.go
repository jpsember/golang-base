package webapp

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/webapp/gen/webapp_data"
	. "github.com/jpsember/golang-base/webserv"
	"sync"
)

// ------------------------------------------------------------------------------------
// This is the canonical boilerplate that I will turn into a goland live template,
// to simplify creating new pages:
// ------------------------------------------------------------------------------------

type SignUpPageStruct struct {
	editor DataEditor
	strict bool
}

type SignUpPage = *SignUpPageStruct

var SignUpPageTemplate = &SignUpPageStruct{}

var CreateUserLock sync.Mutex

func newSignUpPage(session Session) SignUpPage {
	t := &SignUpPageStruct{}
	t.editor = NewDataEditor(NewSignUpState())
	t.generateWidgets(session)
	return t
}

func (p SignUpPage) Name() string   { return "signup" }
func (p SignUpPage) Args() []string { return nil }
func (p SignUpPage) ConstructPage(s Session, args PageArgs) Page {
	// Use the PageArgs to verify that the construction parameters are valid.
	if args.CheckDone() {
		if OptSessionUser(s).Id() == 0 {
			return newSignUpPage(s)
		}
	}
	return nil
}

func (p SignUpPage) generateWidgets(s Session) {
	m := GenerateHeader(s, p)
	m.Label("Sign Up Page").Size(SizeLarge).AddHeading()

	s.PushEditor(p.editor)
	m.Col(6).Open()
	{
		m.Col(12)
		m.Label("User name").Id(SignUpState_Name).AddInput(p.listenerValidateName)
		m.Label("Password").Id(SignUpState_Password).AddPassword(p.listenerValidatePwd)
		m.Label("Password Again").Id(SignUpState_PasswordVerify).AddPassword(p.listenerValidatePwdVerify)
		m.Label("Email").Id(SignUpState_Email).AddInput(p.validateEmail)
		m.Size(SizeTiny).Label("We will never share your email address with anyone.").AddText()
		m.Col(6)
		m.AddSpace()
		m.Label("Sign Up").Listener(p.signUpListener).AddBtn()
	}
	m.Close()

	s.PopEditor()
}

// ------------------------------------------------------------------------------------

func (p SignUpPage) validateFlag() ValidateFlag {
	return Ternary(p.strict, 0, VALIDATE_EMPTYOK)
}

func (p SignUpPage) listenerValidateName(s Session, widget InputWidget, value string) (string, error) {
	return ValidateUserName(value, p.validateFlag())
}

func (p SignUpPage) listenerValidatePwd(s Session, widget InputWidget, value string) (string, error) {
	pr := PrIf(">listenerValidatePwd", false)
	flag := p.validateFlag()
	pr("Validating, value:", QUO, value)
	validated, err := ValidateUserPassword(value, flag)
	pr("after validating:", validated, "err:", err)
	if !p.strict {
		// We must DELAY this additional validation until after the current validation has completed,
		// else the most recent password value isn't the one we will read

		// If this becomes a common idiom, we will add a function s.PostValidate(...)
		s.AddPostRequestEvent(func() { s.Validate(s.Get(SignUpState_PasswordVerify)) })
	}
	return validated, err
}

func (p SignUpPage) listenerValidatePwdVerify(s Session, widget InputWidget, value string) (string, error) {
	pr := PrIf(">listenerValidatePwdVerify", false)
	pr("verify value  :", QUO, value)
	var err error
	flag := p.validateFlag()
	if flag.Has(VALIDATE_EMPTYOK) && value == "" {
	} else {
		value1 := p.editor.GetString(SignUpState_Password)
		pr("password value:", QUO, value1)
		pr("editor state:", INDENT, p.editor.JSMap)
		err, value = replaceWithTestInput(err, value, "a", value1)
		if value1 != value {
			err = Ternary(value == "", ErrorEmptyUserPassword, ErrorUserPasswordsDontMatch)
		}
	}
	pr("returning:", QUO, value, "err:", err)
	return value, err
}

func (p SignUpPage) validateEmail(s Session, widget InputWidget, value string) (string, error) {
	return ValidateEmailAddress(value, p.validateFlag())
}

func (p SignUpPage) signUpListener(s Session, widget Widget, args WidgetArgs) {
	pr := PrIf("signupListener", false)
	pr("state:", INDENT, p.editor.JSMap)

	// Re-validate all the widgets in 'strict' mode.
	p.strict = true
	errcount := s.ValidateAndCountErrors(s.PageWidget)
	p.strict = false

	pr("after validating page;")
	pr("state:", INDENT, p.editor.JSMap)

	pr("error count:", errcount)
	if errcount != 0 {
		return
	}

	// Construct a user by parsing the signupstate map
	b := DefaultUser.Parse(p.editor).(User).ToBuilder()
	b.SetUserClass(UserClassDonor)

	hash, salt := HashPassword(b.Password())
	b.SetPasswordHash(hash)
	b.SetPasswordSalt(salt)

	ub, err := p.attemptCreateUniqueUser(b)
	if err != nil {
		switch err {
		case nameExists:
			s.SetWidgetProblem(SignUpState_Name, err.Error())
		case emailExists:
			s.SetWidgetProblem(SignUpState_Email, err.Error())
		default:
			s.SetWidgetProblem(SignUpState_Name, "Sorry, an error occurred.")
		}
		return
	}

	Pr("created user:", INDENT, ub)

	Todo("add support for WaitingActivation")
	AttemptSignIn(s, ub.Id())
}

var nameExists = Error("A user with this name already exists")
var emailExists = Error("A user with this email already exists")

func (p SignUpPage) attemptCreateUniqueUser(b UserBuilder) (User, error) {
	CreateUserLock.Lock()
	defer CreateUserLock.Unlock()

	var existing User

	existing, _ = ReadUserWithName(b.Name())
	if existing.Id() != 0 {
		return nil, nameExists
	}
	existing, _ = ReadUserWithEmail(b.Email())
	if existing.Id() != 0 {
		return nil, emailExists
	}

	user, err := CreateUserWithName(b.Name())
	if err != nil {
		return nil, err
	}
	if user.Id() == 0 {
		return nil, nameExists
	}
	return user, nil
}
