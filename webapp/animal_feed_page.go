package webapp

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/webapp/gen/webapp_data"
	. "github.com/jpsember/golang-base/webserv"
)

const FeedPageName = "feed"

var FeedPageTemplate = NewFeedPage(nil)

func (p FeedPage) Construct(s Session, args PageArgs) Page {
	if args.CheckDone() {
		if SessionUserIs(s, UserClassDonor) {
			Todo("Maybe don't set the url string until a page is accepted?")
			return NewFeedPage(s)
		}
	}
	return nil
}

func (p FeedPage) Args() []string { return EmptyStringSlice }

type FeedPageStruct struct {
	listWidget ListWidget
}

type FeedPage = *FeedPageStruct

func NewFeedPage(s Session) FeedPage {
	t := &FeedPageStruct{}
	return t
}

const feed_id_prefix = FeedPageName + "."
const (
	id_feed_list = feed_id_prefix + "list"
)

func (p FeedPage) Name() string { return FeedPageName }

func (p FeedPage) GenerateWidgets(s Session) {
	// Set click listener for this page
	s.SetClickListener(p.clickListener)

	m := GenerateHeader(s, p)
	al := p.animalList(s)
	p.listWidget = m.Id(id_feed_list).AddList(al, p.renderItem, p.listListener)
}

func (p FeedPage) animalList(s Session) AnimalList {
	key := SessionKey_FeedList
	alist := s.OptSessionData(key)
	if alist == nil {
		Todo("!store the animalList in the FeedPage struct")
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
	sess.SwitchToPage(NewCreateAnimalPage(sess))
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

func (p FeedPage) renderItem(session Session, widget ListWidget, elementId int, m MarkupBuilder) {
	anim, err := ReadActualAnimal(elementId)
	if ReportIfError(err, "renderItem in animal feed page:", elementId) {
		return
	}
	m.OpenTag(`div class="col-sm-3"`)
	RenderAnimalCard(session, anim, m, "Edit", action_prefix_animal_card, action_prefix_animal_card)
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
		sess.SwitchToPage(NewViewAnimalPage(sess, anim.Id()))
		return
	}

	if p.listWidget.HandleClick(sess, message) {
		return
	}
	Alert("#50Ignoring click:", message)
}
