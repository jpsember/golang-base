package webapp

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/webapp/gen/webapp_data"
	. "github.com/jpsember/golang-base/webserv"
)

type ManagerPageStruct struct {
	session Session
}

type ManagerPage = *ManagerPageStruct

func NewManagerPage(session Session) ManagerPage {
	t := &ManagerPageStruct{session: session}
	return t
}
func (p ManagerPage) Session() Session { return p.session }

var ManagerPageTemplate = NewManagerPage(nil)

func (p ManagerPage) Name() string {
	return ManagerPageName
}

func (p ManagerPage) Construct(s Session) Page {
	return NewManagerPage(s)
}

const ManagerPageName = "manager"

const manager_id_prefix = ManagerPageName + "."
const (
	id_manager_list = manager_id_prefix + "list"
)

func (p ManagerPage) Generate() {
	sess := p.session
	Todo("?Think about ways of cleaning up the click listener which is not tied to a widget")
	//SetWidgetDebugRendering()
	m := GenerateHeader(p)

	Todo("?If we are generating a new page, we shouldn't try to store the error in the old one")
	// Row of buttons at top.
	m.Open()
	{
		m.Label("New Animal").AddButton(p.newAnimalListener)
	}
	m.Close()

	// Set click listener for the card list
	sess.SetClickListener(p.clickListener)

	al := p.animalList()
	m.Id(id_manager_list).AddList(al, p.renderItem, p.listListener)
}

func (p ManagerPage) animalList() AnimalList {
	key := SessionKey_MgrList
	alist := p.Session().OptSessionData(key)
	if alist == nil {
		alist = p.constructAnimalList()
		p.Session().PutSessionData(key, alist)
	}
	return alist.(AnimalList)
}

func (p ManagerPage) constructAnimalList() AnimalList {
	managerId := SessionUser(p.Session()).Id()
	animalList := NewAnimalList(getManagerAnimals(managerId))
	return animalList
}

func (p ManagerPage) newAnimalListener(sess Session, widget Widget) {
	NewCreateAnimalPage(sess).Generate()
}

func (p ManagerPage) listListener(sess Session, widget ListWidget) error {
	Pr("listener event:", widget.Id())
	return nil
}

func getManagerAnimals(managerId int) []int {
	Todo("?A compound index on managerId+animalId would help here, but probably not worth it for now")
	var result []int
	{
		iter := AnimalIterator(0)
		for iter.HasNext() {
			anim := iter.Next().(Animal)
			if anim.ManagerId() == managerId {
				result = append(result, anim.Id())
			}
		}
	}
	if false && Alert("choosing a much larger random list") {
		iter := AnimalIterator(0)
		for iter.HasNext() {
			anim := iter.Next().(Animal)
			result = append(result, anim.Id())
		}
	}
	return result
}

func (p ManagerPage) renderItem(widget ListWidget, elementId int, m MarkupBuilder) {
	anim, err := ReadActualAnimal(elementId)
	if ReportIfError(err, "renderItem in manager page page:", elementId) {
		return
	}

	if false {
		m.OpenTag(`div class="card bg-light mb-3"`)
		m.A("animal ", elementId)
		m.CloseTag()
		return
	}

	//<div class="card bg-light mb-3 animal-card">

	m.OpenTag(`div class="col-sm-3"`)
	RenderAnimalCard(p.Session(), anim, m, "Edit", action_prefix_animal_card, action_prefix_animal_card)
	m.CloseTag()
}

const action_prefix_animal_card = "animal_id_"

func (p ManagerPage) clickListener(sess Session, message string) {
	if id_str, f := TrimIfPrefix(message, action_prefix_animal_card); f {
		id, err := ParseAsPositiveInt(id_str)
		if ReportIfError(err) {
			return
		}
		anim, err := ReadActualAnimal(id)
		if err != nil || anim.Id() == 0 {
			Alert("#50trouble reading animal for clickListener message", message)
			return
		}
		sess.SetClickListener(nil)

		NewEditAnimalPage(sess, anim.Id()).Generate()
	}
}
