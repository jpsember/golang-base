package webapp

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/webapp/gen/webapp_data"
	. "github.com/jpsember/golang-base/webserv"
	"strings"
)

type PageRequesterStruct struct {
	BaseObject
	registry map[string]Page
}

func NewPageRequester() PageRequester {
	t := &PageRequesterStruct{
		registry: make(map[string]Page),
	}
	t.SetName("PageRequester")
	return t
}

type PageRequester = *PageRequesterStruct

// Get the name of the default page for a user
func (r PageRequester) DefaultPage(user User) string {
	userId := 0
	if user != nil {
		userId = user.Id()
	}

	var result string
	if userId == 0 || !IsUserLoggedIn(user.Id()) {
		result = LandingPageName
	} else {
		switch user.UserClass() {
		case UserClassDonor:
			result = FeedPageName
		case UserClassManager:
			result = ManagerPageName
		default:
			NotSupported("page for", user.UserClass())
		}
	}
	return result
}

func (r PageRequester) Process(s Session, path string) Page {
	r.AlertVerbose()
	pr := r.Log

	p := NewPathParse(path)
	pr("Process path:", p)

	user := OptSessionUser(s)

	requestedPageName, _ := p.ReadIf()

	if requestedPageName == "" /*|| !HasKey(r.registry, requestedPageName) */ {
		requestedPageName = r.DefaultPage(user)
	}

	pr("getting template from registry for:", requestedPageName)

	templatePage := r.registry[requestedPageName]
	if templatePage == nil {
		pr("...could not find any page for:", Quoted(requestedPageName))
		return nil
	}
	page := templatePage.Construct(s)
	pr("constructed page:", page)
	return page
	//s.SetURLExpression(requestedPageName)
	//Todo("Maybe the Generate function should be in the abstract Page type?")
	//page.GetBasicPage().Generate()
	//pr("generated page")
	//
	//return true
	//
	//var resultPath string
	//
	//if !IsUserLoggedIn(user.Id()) {
	//	resultPath = "/"
	//} else {
	//	resultPath = "/feed"
	//}
	//Pr(resultPath)
	//
	//return false
}

func (r PageRequester) RegisterPage(template Page) {
	key := template.Name()
	if HasKey(r.registry, key) {
		BadState("duplicate page in registry:", key)
	}
	r.registry[key] = template
}

type PathParseStruct struct {
	text   string
	parts  []string
	cursor int
}

type PathParse = *PathParseStruct

func NewPathParse(text string) PathParse {
	t := &PathParseStruct{
		text: text,
	}
	t.parse()
	return t
}

func (p PathParse) String() string { return p.JSMap().String() }

func (p PathParse) JSMap() JSMap {
	x := NewJSMap()
	x.Put("text", p.text)
	x.Put("parts", JSListWith(p.Parts()))
	return x
}

func (p PathParse) Parts() []string {
	return p.parts
}

func (p PathParse) HasNext() bool {
	return p.cursor < len(p.parts)
}

func (p PathParse) Peek() string {
	if p.HasNext() {
		return p.parts[p.cursor]
	}
	return ""
}

func (p PathParse) PeekInt() (int, bool) {
	x := p.Peek()
	if x != "" {
		val, err := ParseInt(x)
		if err == nil {
			return int(val), true
		}
	}
	return -1, false
}

func (p PathParse) ReadIf() (string, bool) {
	if p.HasNext() {
		return p.Read(), true
	}
	return "", false
}

func (p PathParse) ReadIfInt() (int, bool) {
	x, flag := p.PeekInt()
	if flag {
		p.cursor++
	}
	return x, flag
}

func (p PathParse) Read() string {
	CheckState(p.HasNext())
	x := p.Peek()
	p.cursor++
	return x
}

func (p PathParse) parse() {
	if p.parts != nil {
		return
	}
	c := strings.TrimSpace(p.text)
	strings.TrimSuffix(c, "/")
	substr := strings.Split(c, "/")
	var f []string
	for _, x := range substr {
		x := strings.TrimSpace(x)
		if x == "" || x == "/" {
			continue
		}
		f = append(f, x)
	}
	p.parts = f
}
