package webapp

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/webapp/gen/webapp_data"
	. "github.com/jpsember/golang-base/webserv"
)

type ManagerPageStruct struct {
	manager    User
	listWidget ListWidget
	animList   AnimalList
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
	return managerPageName
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

const managerPageName = "manager"
const manager_id_prefix = managerPageName + "."
const (
	id_manager_list = manager_id_prefix + "list"
)

func (p ManagerPage) generateWidgets(sess Session) {
	Todo("?Think about ways of cleaning up the click listener which is not tied to a widget")
	m := GenerateHeader(sess, p)

	m.AddUserHeader()

	// Row of buttons at top.
	m.Open()
	{
		m.Label("New Animal").AddButton(p.newAnimalListener)
	}
	m.Close()

	// Set click listener for the card list
	sess.SetClickListener(p.clickListener)

	// Construct widget to use in list
	cardWidget := p.constructListItemWidget(sess)
	p.listWidget = m.Id(id_manager_list).AddList(p.animalList(sess), cardWidget, cardWidget.StateProviderFunc(), p.listListener)
}

func (p ManagerPage) constructListItemWidget(s Session) NewCard {
	m := s.WidgetManager()
	Todo("We need a way to construct a widget that isn't attached to a container")

	cardListener := func(sess Session, widget NewCard) {
		Pr("card listener, animal id:", widget.Animal().Id())
		p.attemptSelectAnimal(sess, widget.Animal().Id())
	}

	// Construct the list item widget by adding it to the page (which adds its children as well).  Then, detach the item.
	w := NewNewCard(m.AllocateAnonymousId("manager_item"), DefaultAnimal,
		cardListener, "hey", cardListener)
	m.Add(w)
	m.Detach(w)
	return w
}

func (p ManagerPage) animalList(s Session) AnimalList {
	alist := p.animList
	if alist == nil {
		alist = p.constructAnimalList(s)
		p.animList = alist
	}
	return alist
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

	m.OpenTag(`div class="col-sm-3"`)
	RenderAnimalCard(session, anim, m, "Edit", action_prefix_animal_card, action_prefix_animal_card)
	m.CloseTag()
}

const action_prefix_animal_card = "animal_id_"

func (p ManagerPage) clickListener(sess Session, message string) bool {

	for {
		if ProcessUserHeaderClick(sess, message) {
			break
		}

		if id_str, f := TrimIfPrefix(message, action_prefix_animal_card); f {
			id, err := ParseAsPositiveInt(id_str)
			if ReportIfError(err) {
				break
			}
			p.attemptSelectAnimal(sess, id)
			break
			//animal, err := ReadActualAnimal(id)
			//if err != nil || animal.Id() == 0 {
			//	Alert("#50trouble reading animal for clickListener message", message)
			//	break
			//}
			//if animal.ManagerId() != p.manager.Id() {
			//	Alert("#50wrong manager for animal", message, animal)
			//	break
			//}
			//sess.SetClickListener(nil)
			//sess.SwitchToPage(NewEditAnimalPage(sess, animal.Id()))
			//break
		}

		if p.listWidget.HandleClick(sess, message) {
			break
		}
		return false
	}
	return true

}

func (p ManagerPage) attemptSelectAnimal(s Session, id int) bool {
	animal, err := ReadActualAnimal(id)
	if err != nil || animal.Id() == 0 {
		Alert("#50trouble reading animal:", id)
		return false
	}
	if animal.ManagerId() != p.manager.Id() {
		Alert("#50wrong manager for animal", animal)
		return false
	}
	Todo("clear click listener on switch page?")
	s.SetClickListener(nil)
	s.SwitchToPage(NewEditAnimalPage(s, animal.Id()))
	return true
}
