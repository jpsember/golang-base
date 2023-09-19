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
	m.AddUserHeader()

	// Construct widget to use in list
	listItemWidget := p.constructListItemWidget(s)

	al := p.animalList(s)
	p.listWidget = m.Id(id_feed_list).AddList(al, listItemWidget, p.listItemStateProvider, p.listListener)
}

func (p FeedPage) constructListItemWidget(s Session) Widget {
	m := s.WidgetManager()
	Todo("We need a way to construct a widget that isn't attached to a container")
	w := m.Open()
	m.Id("foo_text").AddText()
	m.Close()
	return m.Detach(w)
}

func (p FeedPage) listItemStateProvider(sess Session, widget *ListWidgetStruct, elementId int) (string, JSMap) {
	json := NewJSMap()
	json.Put("foo_text", ToString("Item #", elementId))
	return "", json
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
		anim, err := ReadActualAnimal(id)
		if ReportIfError(err, "AnimalFeed message", message) {
			return true
		}
		sess.SetClickListener(nil)
		sess.SwitchToPage(NewViewAnimalPage(sess, anim.Id()))
		return true
	}

	if p.listWidget.HandleClick(sess, message) {
		return true
	}
	return false
}
