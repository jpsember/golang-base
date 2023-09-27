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

var ManagerPageTemplate = &ManagerPageStruct{}

func (p ManagerPage) Name() string {
	return managerPageName
}

func (p ManagerPage) ConstructPage(s Session, args PageArgs) Page {
	user := OptSessionUser(s)
	if user.UserClass() == UserClassManager {
		if args.CheckDone() {
			t := &ManagerPageStruct{}
			t.manager = SessionUser(s)
			t.generateWidgets(s)
			return t
		}
	}
	return nil
}

func (p ManagerPage) Args() []string { return nil }

const managerPageName = "manager"
const manager_id_prefix = managerPageName + "."
const manager_card_id = manager_id_prefix + "card"

func (p ManagerPage) generateWidgets(sess Session) {
	m := GenerateHeader(sess, p)
	debug := m.StartConstruction()
	AddUserHeaderWidget(sess)

	// Row of buttons at top.
	m.Open()
	{
		m.Label("New Animal").AddButton(p.newAnimalListener)
	}
	m.Close()

	// Construct a list, and a card to use as the list item widget

	var cardWidget AnimalCard
	{
		w := NewAnimalCard(m, DefaultAnimal,
			func(sess Session, widget AnimalCard, arg string) {
				animalId := sess.Context().(int)
				p.attemptSelectAnimal(sess, animalId)
			}, "", nil)
		cardWidget = w
		m.Add(w)
	}

	managerId := SessionUser(sess).Id()
	animalList := NewAnimalList(getManagerAnimals(managerId), cardWidget)

	Todo("!document how the list forwards clicks related to items on to the list listener")
	m.AddList(animalList, cardWidget)
	m.EndConstruction(debug)
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
	s.SwitchToPage(EditAnimalPageTemplate, PageArgsWith(animal.Id()))
	return true
}

func (p ManagerPage) newAnimalListener(sess Session, widget Widget, arg string) {
	sess.SwitchToPage(CreateAnimalPageTemplate, nil)
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

const action_prefix_animal_card = "animal_id_"
