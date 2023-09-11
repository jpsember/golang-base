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

const manager_id_prefix = "manager."
const (
	id_manager_list = manager_id_prefix + "list"
)

func (p ManagerPage) Generate() {
	SetWidgetDebugRendering()
	m := p.GenerateHeader()

	Todo("?If we are generating a new page, we shouldn't try to store the error in the old one")
	// Row of buttons at top.
	m.Open()
	{
		m.Label("New Animal").AddButton(p.newAnimalListener)
	}
	m.Close()

	// Scrolling list of animals for this manager.
	m.Open()
	al := p.animalList()
	m.Id(id_manager_list).AddList(al, nil, p.listListener)

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
	managerId := SessionUser(p.session).Id()
	animalList.elements = getManagerAnimals(managerId)
	Pr("animals for manager", managerId, ":", INDENT, animalList.elements)
	return animalList
}

func (p ManagerPage) listListener(sess Session, widget ListWidget) error {
	Pr("listener event:", widget.Id())
	return nil
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
	Pr("returning:", result)
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
