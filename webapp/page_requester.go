package webapp

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/webapp/gen/webapp_data"
	. "github.com/jpsember/golang-base/webserv"
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
func (r PageRequester) DefaultPagePage(user User) Page {
	nm := r.DefaultPage(user)
	return r.PageWithNameM(nm)
}
func (r PageRequester) PageWithName(nm string) Page {
	return r.registry[nm]
}
func (r PageRequester) PageWithNameM(nm string) Page {
	pg := r.registry[nm]
	if pg == nil {
		BadArg("No page found with name:", nm)
	}
	return pg
}

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

	requestedPageName := p.Read()

	if requestedPageName == "" /*|| !HasKey(r.registry, requestedPageName) */ {
		requestedPageName = r.DefaultPage(user)
	}

	pr("getting template from registry for:", requestedPageName)

	templatePage := r.PageWithName(requestedPageName)
	if templatePage == nil {
		if requestedPageName != "" {
			pr("...could not find any page for:", Quoted(requestedPageName))
		}
		requestedPageName = r.DefaultPage(user)
	}
	templatePage = r.PageWithNameM(requestedPageName)

	page := templatePage.Construct(s)

	// Ask the page to confirm it is ok

	if len(s.PendingURLArgs2) != 0 {
		BadState("PendingURLArgs is not empty")
	}
	page = page.Request(s, p)
	pr("requested page:", page, "url expr:", s.PendingURLExpr2)
	if page == nil {
		page = r.DefaultPagePage(user)
	}
	CheckState(page != nil, "requested page is nil")

	// Construct a non-template version of the page to return
	page = page.Construct(s, s.PendingURLArgs2...)
	Todo("how do we transmit the url args to the page construction args?")
	return page
}

func (r PageRequester) RegisterPages(template ...Page) {
	for _, t := range template {
		r.RegisterPage(t)
	}
}

func (r PageRequester) RegisterPage(template Page) {
	key := template.Name()
	if HasKey(r.registry, key) {
		BadState("duplicate page in registry:", key)
	}
	r.registry[key] = template
}
