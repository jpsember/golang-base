package webapp

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/webapp/gen/webapp_data"
	. "github.com/jpsember/golang-base/webserv"
)

const FeedPageName = "feed"

var FeedPageTemplate = &FeedPageStruct{}

type FeedPageStruct struct {
	animList AnimalList
}

type FeedPage = *FeedPageStruct

func (p FeedPage) Name() string { return FeedPageName }

func (p FeedPage) ConstructPage(s Session, args PageArgs) Page {
	if args.CheckDone() {
		if SessionUserIs(s, UserClassDonor) {
			t := &FeedPageStruct{}
			t.generateWidgets(s)
			return t
		}
	}
	return nil
}

func (p FeedPage) Args() []string { return nil }

func (p FeedPage) generateWidgets(s Session) {
	m := GenerateHeader(s, p)
	debug := m.StartConstruction()

	AddUserHeaderWidget(s)

	{
		m.PushStateMap(NewJSMap())
		// For now, write the code as one big function; split up later once structure is more apparent.
		var cardWidget AnimalCard
		{
			// Construct the list item widget
			w := NewAnimalCard(m, DefaultAnimal,
				func(sess Session, widget AnimalCard, arg string) {
					animalId := sess.Context().(int)
					Pr("card listener, animal id:", animalId, "arg:", arg)
					p.attemptSelectAnimal(sess, animalId)
				},
				"", nil)

			cardWidget = w
		}
		m.Add(cardWidget)

		animalList := NewAnimalList(getAnimals(), cardWidget)
		m.AddList(animalList, cardWidget)
		m.PopStateMap()
	}
	m.EndConstruction(debug)
}

func getAnimals() []int {
	var result []int
	{
		iter := AnimalIterator(0)
		for iter.HasNext() {
			anim := iter.Next().(Animal)
			result = append(result, anim.Id())
		}
	}
	return result
}

func (p FeedPage) attemptSelectAnimal(s Session, id int) bool {
	animal, err := ReadActualAnimal(id)
	if err != nil || animal.Id() == 0 {
		Alert("#50trouble reading animal:", id)
		return false
	}
	s.SwitchToPage(ViewAnimalPageTemplate, PageArgsWith(animal.Id()))
	return true
}
