package webapp

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/webapp/gen/webapp_data"
	. "github.com/jpsember/golang-base/webserv"
)

const FeedPageName = "feed"

var FeedPageTemplate = NewFeedPage(nil)

func (p FeedPage) ConstructPage(s Session, args PageArgs) Page {
	if args.CheckDone() {
		if SessionUserIs(s, UserClassDonor) {
			return NewFeedPage(s)
		}
	}
	return nil
}

func (p FeedPage) Args() []string { return EmptyStringSlice }

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

var fewWidgets = false && Alert("Rendering only a few of the usual widgets")

func (p FeedPage) generateWidgets(s Session) {
	m := GenerateHeader(s, p)
	debug := m.StartConstruction()

	if !fewWidgets {
		AddUserHeaderWidget(s)
	}

	// Construct widget to use in list
	cardWidget := p.constructListItemWidget(s)
	listWidget := m.AddList(p.animalList(s), cardWidget, cardWidget.StateProviderFunc())
	if fewWidgets {
		listWidget.WithPageControls = false
	}
	listWidget.Listener = p.listListener
	m.EndConstruction(debug)
}

func (p FeedPage) constructListItemWidget(s Session) NewCard {
	m := s.WidgetManager()

	cardListener := func(sess Session, widget NewCard) {
		p.attemptSelectAnimal(sess, widget.Animal().Id())
	}

	// Construct the list item widget by adding it to the page (which adds its children as well).  Then, detach the item.
	//
	// These list cards have no buttons.
	w := NewNewCard(m.AllocateAnonymousId("feed_card"), DefaultAnimal,
		cardListener, //
		"", nil)      //
	m.Add(w)
	m.Detach(w)
	return w
}

func (p FeedPage) animalList(s Session) AnimalList {
	if p.animList == nil {
		p.animList = p.constructAnimalList()
	}
	return p.animList
}

func (p FeedPage) constructAnimalList() AnimalList {
	animalList := NewAnimalList(getAnimals())
	return animalList
}

func (p FeedPage) newAnimalListener(sess Session, widget Widget) {
	sess.SwitchToPage(NewCreateAnimalPage(sess))
}

func getAnimals() []int {
	var result []int
	{
		iter := AnimalIterator(0)
		for iter.HasNext() {
			anim := iter.Next().(Animal)
			result = append(result, anim.Id())
		}
		if fewWidgets {
			if len(result) > 2 {
				result = result[0:2]
			}

		}
	}
	return result
}

func (p FeedPage) listListener(sess Session, widget *ListWidgetStruct, itemId int, args string) {
	p.attemptSelectAnimal(sess, itemId)
}

func (p FeedPage) attemptSelectAnimal(s Session, id int) bool {
	animal, err := ReadActualAnimal(id)
	if err != nil || animal.Id() == 0 {
		Alert("#50trouble reading animal:", id)
		return false
	}
	s.SwitchToPage(NewViewAnimalPage(s, animal.Id()))
	return true
}
