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
	listWidget ListWidget
	animList   AnimalList
}

type FeedPage = *FeedPageStruct

func NewFeedPage(s Session) FeedPage {
	t := &FeedPageStruct{}
	if s != nil {
		t.generateWidgets(s)
	}
	return t
}

const feed_id_prefix = FeedPageName + "."
const (
	id_feed_list = feed_id_prefix + "list"
)

func (p FeedPage) Name() string { return FeedPageName }

func (p FeedPage) generateWidgets(s Session) {
	// Set click listener for this page
	s.SetClickListener(p.clickListener)

	m := GenerateHeader(s, p)
	if !Alert("not adding user header for now") {
		m.AddUserHeader()
	}

	// Set click listener for the card list
	s.SetClickListener(p.clickListener)

	// Construct widget to use in list
	cardWidget := p.constructListItemWidget(s)
	p.listWidget = m.Id(id_feed_list).AddList(p.animalList(s), cardWidget, cardWidget.StateProviderFunc())
	p.listWidget.WithPageControls = !Alert("!disabling page controls")
}

func (p FeedPage) constructListItemWidget(s Session) NewCard {
	m := s.WidgetManager()

	cardListener := func(sess Session, widget NewCard) {
		p.attemptSelectAnimal(sess, widget.Animal().Id())
	}

	// Construct the list item widget by adding it to the page (which adds its children as well).  Then, detach the item.
	w := NewNewCard(m.AllocateAnonymousId("donor_item"), DefaultAnimal,
		cardListener, "hey", cardListener)
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
		if Alert("!trimming animal list to a couple of items") {
			if len(result) > 2 {
				result = result[0:2]
			}
		}
	}
	return result
}

func (p FeedPage) renderItem(session Session, widget ListWidget, elementId int, m MarkupBuilder) {
	anim, err := ReadActualAnimal(elementId)
	if ReportIfError(err, "renderItem in animal feed page:", elementId) {
		return
	}
	Todo("How do we render a card widget as a list item though?")
	m.OpenTag(`div class="col-sm-3"`)
	RenderAnimalCard(session, anim, m, "View", action_prefix_animal_card, action_prefix_animal_card)
	m.CloseTag()
}

func (p FeedPage) clickListener(sess Session, message string) bool {

	if ProcessUserHeaderClick(sess, message) {
		return true
	}

	if id_str, f := TrimIfPrefix(message, action_prefix_animal_card); f {
		id, err1 := ParseAsPositiveInt(id_str)
		if ReportIfError(err1, "AnimalFeedPage parsing", message) {
			return true
		}
		p.attemptSelectAnimal(sess, id)
		return true
		//anim, err := ReadActualAnimal(id)
		//if ReportIfError(err, "AnimalFeed message", message) {
		//	return true
		//}
		//sess.SetClickListener(nil)
		//sess.SwitchToPage(NewViewAnimalPage(sess, anim.Id()))
		//return true
	}

	if p.listWidget.HandleClick(sess, message) {
		return true
	}
	return false
}

func (p FeedPage) attemptSelectAnimal(s Session, id int) bool {
	animal, err := ReadActualAnimal(id)
	if err != nil || animal.Id() == 0 {
		Alert("#50trouble reading animal:", id)
		return false
	}
	Todo("clear click listener on switch page?")
	s.SetClickListener(nil)
	s.SwitchToPage(NewViewAnimalPage(s, animal.Id()))
	return true
}
