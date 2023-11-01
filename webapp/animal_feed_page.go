package webapp

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/webapp/gen/webapp_data"
	. "github.com/jpsember/golang-base/webserv"
)

const FeedPageName = "feed"

var FeedPageTemplate = &FeedPageStruct{}

type FeedPageStruct struct {
	animList       AnimalList
	animListWidget ListWidget
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
		const itemPrefix = "feed_item:"

		// Construct the list item widget

		cardListener := func(sess *SessionStruct, widget Widget, args WidgetArgs) {
			Alert("Is this an element id, or an item index?")
			p.attemptSelectAnimal(sess, p.animListWidget.CurrentElement())
		}
		cardWidget := NewAnimalCard(m, itemPrefix, cardListener, "", nil)

		m.Add(cardWidget)

		animalList := NewAnimalList(getAnimals(), cardWidget, itemPrefix)

		p.animListWidget = m.AddList(animalList, cardWidget)
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
