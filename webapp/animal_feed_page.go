package webapp

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/webapp/gen/webapp_data"
	. "github.com/jpsember/golang-base/webserv"
)

// ------------------------------------------------------------------------------------
// Page implementation
// ------------------------------------------------------------------------------------

const FeedPageName = "feed"

var FeedPageTemplate = NewFeedPage(nil)

func (p FeedPage) Session() Session {
	return p.session
}

func (p FeedPage) Construct(s Session) Page {
	return NewLandingPage(s)
}

// ------------------------------------------------------------------------------------

type FeedPageStruct struct {
	session Session
}

type FeedPage = *FeedPageStruct

func NewFeedPage(s Session) FeedPage {
	t := &FeedPageStruct{
		session: s,
	}
	return t
}

const feed_id_prefix = FeedPageName + "."
const (
	id_feed_list = feed_id_prefix + "list"
)

func (p FeedPage) Name() string { return FeedPageName }

func (p FeedPage) Generate() {
	s := p.session
	// Set click listener for this page
	s.SetClickListener(p.clickListener)

	Todo("How do we modify the client's URL to e.g. set `/manager`?")

	m := GenerateHeader(p)

	// If no animals found, add some
	if DevDatabase && !HasAnimals() {
		GenerateRandomAnimals()
	}

	al := p.animalList()
	m.Id(id_feed_list).AddList(al, p.renderItem, p.listListener)
}

func (p FeedPage) animalList() AnimalList {
	key := SessionKey_FeedList
	s := p.Session()
	alist := s.OptSessionData(key)
	if alist == nil {
		alist = p.constructAnimalList()
		s.PutSessionData(key, alist)
	}
	return alist.(AnimalList)
}

func (p FeedPage) constructAnimalList() AnimalList {
	animalList := NewAnimalList(getAnimals())
	return animalList
}

func (p FeedPage) newAnimalListener(sess Session, widget Widget) {
	NewCreateAnimalPage(sess).Generate()
}

func (p FeedPage) listListener(sess Session, widget ListWidget) error {
	Pr("listener event:", widget.Id())
	return nil
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

func (p FeedPage) renderItem(widget ListWidget, elementId int, m MarkupBuilder) {
	anim, err := ReadActualAnimal(elementId)
	if ReportIfError(err, "renderItem in animal feed page:", elementId) {
		return
	}
	m.OpenTag(`div class="col-sm-3"`)
	RenderAnimalCard(p.Session(), anim, m, "Edit", action_prefix_animal_card, action_prefix_animal_card)
	m.CloseTag()
}

func (p FeedPage) clickListener(sess Session, message string) {
	if id_str, f := TrimIfPrefix(message, action_prefix_animal_card); f {
		id, err1 := ParseAsPositiveInt(id_str)
		if ReportIfError(err1, "AnimalFeedPage parsing", message) {
			return
		}
		anim, err := ReadActualAnimal(id)
		if ReportIfError(err, "AnimalFeed message", message) {
			return
		}
		sess.SetClickListener(nil)
		NewViewAnimalPage(sess, anim.Id()).Generate()
		return
	}

	Todo("Pages, and perhaps Sessions, should have embeddings to simplify expressions like this one:")
	listWidget := p.Session().WidgetManager().Get(id_feed_list).(ListWidget)

	if listWidget.HandleClick(sess, message) {

		return
	}
	Alert("#50Ignoring click:", message)

}
