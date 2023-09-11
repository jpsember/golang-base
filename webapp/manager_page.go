package webapp

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/webserv"
)

type ManagerPageStruct struct {
	BasicPage
}

type ManagerPage = *ManagerPageStruct

func NewManagerPage(sess Session, parentWidget Widget) ManagerPage {
	t := &ManagerPageStruct{
		NewBasicPage(sess, parentWidget),
	}
	t.devLabel = "manager_page"
	return t
}

func (p ManagerPage) Generate() {
	Todo("I suspect we don't need to store the session in the page, as we only ever call the Generate function")
	m := p.GenerateHeader()

	Todo("If we are generating a new page, we shouldn't try to store the error in the old one")
	// Row of buttons at top.
	m.Open()
	{
		m.Label("New Animal").AddButton(p.newAnimalListener)
	}
	m.Close()

	Todo("ability to store some user-specific data types in the session other than the state")

	// Scrolling list of animals for this manager.
	m.Open()
	Todo("?Scrolling list of manager's animals")
	p.animalList(p.session)
	m.Close()

}

func (p ManagerPage) newAnimalListener(sess Session, widget Widget) error {
	NewCreateAnimalPage(sess, p.parentPage).Generate()
	return nil
}

func (p ManagerPage) animalList(sess Session) AnimalList {

	alist := sess.OptSessionData(UserKey_MgrList)
	if alist == nil {
		als := NewAnimalList()
		sess.PutSessionData(UserKey_MgrList, als)
		Pr("empty elements:", als.GetPageElements(0))
		alist = als
	}
	return alist.(AnimalList)
}

type AnimalListStruct struct {
	elements []int
}

type AnimalList = *AnimalListStruct

func NewAnimalList() AnimalList {
	t := &AnimalListStruct{}
	return t
}
func (a AnimalList) ElementsPerPage() int {
	return 12
}

func (a AnimalList) GetPageElements(pageNumber int) []int {
	k := a.ElementsPerPage()
	pgStart := pageNumber * k
	pgEnd := pgStart + k

	return ClampedSlice(a.elements, pgStart, pgEnd)
}

func ClampedSlice[K any](slice []K, start int, end int) []K {
	start = Clamp(start, 0, len(slice))
	end = Clamp(end, start, len(slice))
	return slice[start:end]
}
