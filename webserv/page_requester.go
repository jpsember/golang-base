package webserv

import (
	. "github.com/jpsember/golang-base/base"
)

type AbstractUser interface {
	Name() string
	Id() int
}

type PageRequesterStruct struct {
	PageRequesterInterface
	BaseObject
	// Maps are thread safe for reading.  We won't modify the map once the map has been initialized.
	registry map[string]Page
}

type PageRequesterInterface interface {
	UserForSession(s Session) AbstractUser
	DefaultPageForUser(user AbstractUser) Page
}

func NewPageRequester(fn PageRequesterInterface) PageRequester {
	t := &PageRequesterStruct{
		PageRequesterInterface: fn,
		registry:               make(map[string]Page),
	}
	t.SetName("PageRequester")
	return t
}

type PageRequester = *PageRequesterStruct

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
func (r PageRequester) DefaultPagePage(user AbstractUser) Page {
	p := r.DefaultPageForUser(user)
	CheckArg(p != nil)
	return p
}

func (r PageRequester) Process(s Session, path string) Page {
	//r.AlertVerbose()
	pr := r.Log

	p := NewPathParse(path)
	pr("Process path:", p)

	user := r.UserForSession(s)

	defPageForUser := r.DefaultPagePage(user)
	requestedPageName := p.Read()

	if requestedPageName == "" {
		requestedPageName = defPageForUser.Name()
	}

	pr("getting template from registry for:", requestedPageName)

	templatePage := r.PageWithName(requestedPageName)
	if templatePage == nil {
		if requestedPageName != "" {
			pr("...could not find any page for:", Quoted(requestedPageName))
		}
		requestedPageName = defPageForUser.Name()
	}
	templatePage = r.PageWithNameM(requestedPageName)

	remainingArgs := NewPageArgs(p.RemainingArgs())
	pr("remaining args:", remainingArgs)
	page := templatePage.ConstructPage(s, remainingArgs)
	if page == nil {
		page = r.DefaultPagePage(user)
		page = page.ConstructPage(s, NewPageArgs(nil))
	}
	CheckState(page != nil, "requested page is nil")
	return page
}

// PageRequester must be threadsafe (once all the pages have been registered).
func (r PageRequester) RegisterPage(template Page) {
	key := template.Name()
	if HasKey(r.registry, key) {
		BadState("duplicate page in registry:", key)
	}
	r.registry[key] = template
}
