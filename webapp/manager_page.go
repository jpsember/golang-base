package webapp

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/webapp/gen/webapp_data"
	. "github.com/jpsember/golang-base/webserv"
)

type ManagerPageStruct struct {
	manager    User
	listWidget ListWidget
}

type ManagerPage = *ManagerPageStruct

func NewManagerPage(session Session) ManagerPage {
	t := &ManagerPageStruct{}
	if session != nil {
		t.manager = SessionUser(session)
		t.generateWidgets(session)
	}
	return t
}

var ManagerPageTemplate = NewManagerPage(nil)

func (p ManagerPage) Name() string {
	return ManagerPageName
}

func (p ManagerPage) ConstructPage(s Session, args PageArgs) Page {
	user := OptSessionUser(s)
	if user.UserClass() == UserClassManager {
		if args.CheckDone() {
			return NewManagerPage(s)
		}
	}
	return nil
}
func (p ManagerPage) Args() []string { return EmptyStringSlice }

const ManagerPageName = "manager"

const manager_id_prefix = ManagerPageName + "."
const (
	id_manager_list = manager_id_prefix + "list"
)

func (p ManagerPage) generateWidgets(sess Session) {
	Todo("?Think about ways of cleaning up the click listener which is not tied to a widget")
	m := GenerateHeader(sess, p)

	Todo("?If we are generating a new page, we shouldn't try to store the error in the old one")
	// Row of buttons at top.
	m.Open()
	{
		m.Label("New Animal").AddButton(p.newAnimalListener)
	}
	m.Close()

	// Set click listener for the card list
	sess.SetClickListener(p.clickListener)

	al := p.animalList(sess)
	p.listWidget = m.Id(id_manager_list).AddList(al, p.renderItem, p.listListener)
}

func (p ManagerPage) animalList(s Session) AnimalList {
	key := SessionKey_MgrList
	alist := s.OptSessionData(key)
	if alist == nil {
		alist = p.constructAnimalList(s)
		s.PutSessionData(key, alist)
		Todo("!We should maybe just store the mgrlist in the ManagerPage struct")
	}
	return alist.(AnimalList)
}

func (p ManagerPage) constructAnimalList(s Session) AnimalList {
	managerId := SessionUser(s).Id()
	animalList := NewAnimalList(getManagerAnimals(managerId))
	return animalList
}

func (p ManagerPage) newAnimalListener(sess Session, widget Widget) {
	sess.SwitchToPage(NewCreateAnimalPage(sess))
}

func (p ManagerPage) listListener(sess Session, widget ListWidget) error {
	Pr("listener event:", widget.Id())
	return nil
}

func getManagerAnimals(managerId int) []int {
	Todo("?A compound index on managerId+animalId would help here, but probably not worth it for now")
	result := []int{}
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

func (p ManagerPage) renderItem(session Session, widget ListWidget, elementId int, m MarkupBuilder) {
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
	RenderAnimalCard(session, anim, m, "Edit", action_prefix_animal_card, action_prefix_animal_card)
	m.CloseTag()
}

const action_prefix_animal_card = "animal_id_"

func (p ManagerPage) clickListener(sess Session, message string) {
	if id_str, f := TrimIfPrefix(message, action_prefix_animal_card); f {
		id, err := ParseAsPositiveInt(id_str)
		if ReportIfError(err) {
			return
		}
		animal, err := ReadActualAnimal(id)
		if err != nil || animal.Id() == 0 {
			Alert("#50trouble reading animal for clickListener message", message)
			return
		}
		if animal.ManagerId() != p.manager.Id() {
			Alert("#50wrong manager for animal", message, animal)
			return
		}
		sess.SetClickListener(nil)
		sess.SwitchToPage(NewEditAnimalPage(sess, animal.Id()))
		return
	}

	if p.listWidget.HandleClick(sess, message) {
		return
	}

}
