package webapp

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/webapp/gen/webapp_data"
	. "github.com/jpsember/golang-base/webserv"
)

type AnimalFeedPageStruct struct {
	BasicPage
}

type AnimalFeedPage = *AnimalFeedPageStruct

const feed_id_prefix = "feed."
const (
	id_feed_list = feed_id_prefix + "list"
)

func NewAnimalFeedPage(sess Session, parentWidget Widget) AnimalFeedPage {
	t := &AnimalFeedPageStruct{
		NewBasicPage(sess, parentWidget),
	}
	t.devLabel = "animal_feed_page"
	return t
}

func (p AnimalFeedPage) Generate() {
	// Set click listener for this page
	p.session.SetClickListener(p.clickListener)

	m := p.GenerateHeader()

	// If no animals found, add some
	if DevDatabase && !HasAnimals() {
		GenerateRandomAnimals()
	}

	al := p.animalList()
	m.Id(id_feed_list).AddList(al, p.renderItem, p.listListener)
}

func (p AnimalFeedPage) animalList() AnimalList {
	key := SessionKey_FeedList
	alist := p.session.OptSessionData(key)
	if alist == nil {
		alist = p.constructAnimalList()
		p.session.PutSessionData(key, alist)
	}
	return alist.(AnimalList)
}

func (p AnimalFeedPage) constructAnimalList() AnimalList {
	animalList := NewAnimalList(getAnimals())
	return animalList
}

func (p AnimalFeedPage) newAnimalListener(sess Session, widget Widget) {
	NewCreateAnimalPage(sess, p.parentPage).Generate()
}

func (p AnimalFeedPage) listListener(sess Session, widget ListWidget) error {
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

func (p AnimalFeedPage) renderItem(widget ListWidget, elementId int, m MarkupBuilder) {
	anim, err := ReadActualAnimal(elementId)
	if ReportIfError(err, "renderItem in animal feed page:", elementId) {
		return
	}
	m.OpenTag(`div class="col-sm-3"`)
	RenderAnimalCard(p.session, anim, m, "Edit", action_prefix_animal_card)
	m.CloseTag()
}

func (p AnimalFeedPage) clickListener(sess Session, message string) {
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
		Todo("Open a 'ViewAnimal' page instead")
		NewEditAnimalPage(sess, sess.PageWidget, anim.Id()).Generate()
		return
	}

	Todo("Pages, and perhaps Sessions, should have embeddings to simplify expressions like this one:")
	listWidget := p.session.WidgetManager().Get(id_feed_list).(ListWidget)

	if listWidget.HandleClick(sess, message) {

		return
	}
	Alert("#50Ignoring click:", message)

}
