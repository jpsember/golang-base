package webapp

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/webapp/gen/webapp_data"
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
	al := p.animalList()
	Todo("do something with", al)
	m.Close()

}

func (p ManagerPage) newAnimalListener(sess Session, widget Widget) error {
	NewCreateAnimalPage(sess, p.parentPage).Generate()
	return nil
}

func (p ManagerPage) animalList() AnimalList {

	alist := p.session.OptSessionData(SessionKey_MgrList)
	if alist == nil {
		alist = p.constructAnimalList()
		p.session.PutSessionData(SessionKey_MgrList, alist)
	}
	return alist.(AnimalList)
}

func (p ManagerPage) constructAnimalList() AnimalList {
	animalList := NewAnimalList()
	Pr("empty elements:", animalList.GetPageElements(0))

	managerId := SessionUser(p.session).Id()
	animalList.elements = getManagerAnimals(managerId)
	Pr("animals for manager", managerId, ":", INDENT, animalList.elements)
	return animalList
}

func getManagerAnimals(managerId int) []int {
	Todo("?A compound index on managerId+animalId would help here, but probably not worth it for now")
	var result []int
	iter := AnimalIterator(0)
	for iter.HasNext() {
		anim := iter.Next().(Animal)
		if anim.ManagerId() == managerId {
			result = append(result, anim.Id())
		}
	}
	return result
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
