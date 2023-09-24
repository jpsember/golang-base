package webapp

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/webapp/gen/webapp_data"
	. "github.com/jpsember/golang-base/webserv"
)

const FeedPageName = "feed"

var FeedPageTemplate = NewFeedPage(nil)

type FeedPageStruct struct {
	animList AnimalList
}

type FeedPage = *FeedPageStruct

func NewFeedPage(s Session) FeedPage {
	t := &FeedPageStruct{}
	if s != nil {
		t.generateWidgets(s)
	}
	return t
}

func (p FeedPage) Name() string { return FeedPageName }

func (p FeedPage) ConstructPage(s Session, args PageArgs) Page {
	if args.CheckDone() {
		if SessionUserIs(s, UserClassDonor) {
			return NewFeedPage(s)
		}
	}
	return nil
}

func (p FeedPage) Args() []string { return nil }

func (p FeedPage) generateWidgets(s Session) {
	m := GenerateHeader(s, p)
	debug := m.StartConstruction()

	AddUserHeaderWidget(s)

	// For now, write the code as one big function; split up later once structure is more apparent.
	var cardWidget AnimalCard
	{
		// Construct the list item widget
		w := NewAnimalCard(m.AllocateAnonymousId("feedcard"), DefaultAnimal,
			func(sess Session, widget AnimalCard, arg string) {
				animalId := sess.Context().(int)
				Pr("card listener, animal id:", animalId, "arg:", arg)
				p.attemptSelectAnimal(sess, animalId)
			},
			"", nil)

		cardWidget = w
	}
	m.Add(cardWidget)
	cardWidget.SetVisible(false)

	animalList := NewAnimalList(getAnimals(), cardWidget)

	if Experiment {
		m.Id("experiment")
	}
	listWidget := m.AddList(animalList, cardWidget)
	if Experiment {
		listWidget.WithPageControls = false
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
	Alert("switching to NewViewAnimalPage for animal:", INDENT, animal)
	Alert("I think some widgets are not being removed from the map")
	s.SwitchToPage(NewViewAnimalPage(s, animal.Id()))
	return true
}
