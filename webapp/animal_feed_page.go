package webapp

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/webapp/gen/webapp_data"
	. "github.com/jpsember/golang-base/webserv"
)

type AnimalFeedPageStruct struct {
	sess         Session
	parentWidget Widget
}

type AnimalFeedPage = *AnimalFeedPageStruct

func NewAnimalFeedPage(sess Session, parentWidget Widget) AnimalFeedPage {
	t := &AnimalFeedPageStruct{
		sess:         sess,
		parentWidget: parentWidget,
	}
	return t
}

func (p AnimalFeedPage) Generate() {
	//SetWidgetDebugRendering()

	m := p.sess.WidgetManager()
	m.With(p.parentWidget)

	m.Size(SizeMicro).Align(AlignRight).Label("AnimalFeedPage, user:" + SessionUser(p.sess).Name()).AddHeading()

	// If no animals found, add some
	if !HasAnimals() {
		GenerateRandomAnimals()
	}

	m.Col(4)
	for i := 1; i < 12; i++ {
		anim, err := ReadAnimal(i)
		if err != nil {
			Pr("what do we do with unexpected errors?", INDENT, err)
		}
		if anim == nil {
			continue
		}
		cardId := "animal_" + IntToString(int(anim.Id()))
		OpenAnimalCardWidget(m, cardId, anim, buttonListener)
	}

}
