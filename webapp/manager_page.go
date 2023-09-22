package webapp

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/webapp/gen/webapp_data"
	. "github.com/jpsember/golang-base/webserv"
)

type ManagerPageStruct struct {
	manager User
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
const manager_card_id = manager_id_prefix + "card"

func (p ManagerPage) generateWidgets(sess Session) {
	m := GenerateHeader(sess, p)

	if !Experiment {
		AddUserHeaderWidget(sess)

		// Row of buttons at top.
		m.Open()
		{
			m.Label("New Animal").AddButton(p.newAnimalListener)
		}
		m.Close()
	}
	// Construct a list, and a card to use as the list item widget

	// For now, write the code as one big function; split up later once structure is more apparent.
	var cardWidget AnimalCard
	{
		cardListener := func(sess Session, widget AnimalCard) {
			Pr("listener for card, id:", widget.Id())
			p.attemptSelectAnimal(sess, widget.Animal().Id())
		}

		// Construct the list item widget by adding it to the page (which adds its children as well).  Then, detach the item.
		w := NewAnimalCard(manager_card_id, DefaultAnimal, cardListener, "hey", cardListener)
		cardWidget = w
		m.Add(w)
		m.Detach(w)
	}

	managerId := SessionUser(sess).Id()
	animalList := NewAnimalList(getManagerAnimals(managerId), cardWidget)

	if Experiment {
		m.Id("experiment")
	}
	Todo("Consider *requiring* a listener (at least a nil one) for AddList")
	Todo("document how the list forwards clicks related to items on to the list listener")
	listWidget := m.AddList(animalList, cardWidget)
	listWidget.Listener = p.listListener
}

func (p ManagerPage) listListener(sess Session, widget *ListWidgetStruct, itemId int, args string) {
	pr := PrIf(Experiment)
	pr("ManagerPage listListener, itemId:", itemId, "args:", args)
	p.attemptSelectAnimal(sess, itemId)
}

func (p ManagerPage) newAnimalListener(sess Session, widget Widget) {
	sess.SwitchToPage(NewCreateAnimalPage(sess))
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
	Alert("finish refactoring for new card")
	/*
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
	*/
}

const action_prefix_animal_card = "animal_id_"

func (p ManagerPage) clickListener(sess Session, message string) bool {
	Todo("This explicit handler probably not required")

	for {

		Alert("Do we still need this code?")
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
	s.SwitchToPage(NewEditAnimalPage(s, animal.Id()))
	return true
}
