// Implementation of ListInterface for lists of animal cards.
package webapp

import (
	. "github.com/jpsember/golang-base/base"
	//"github.com/jpsember/golang-base/webapp/gen/webapp_data"
	. "github.com/jpsember/golang-base/webserv"
)

type AnimalListStruct struct {
	BasicListStruct
	cardWidget AnimalCard
}

type AnimalList = *AnimalListStruct

func NewAnimalList(animalIds []int, cardWidget AnimalCard) AnimalList {
	CheckArg(animalIds != nil)
	t := &AnimalListStruct{
		cardWidget: cardWidget,
	}
	b := &t.BasicListStruct
	b.ElementIds = animalIds
	b.ElementsPerPage = 12
	return t
}

func (a AnimalList) ItemStateProvider(s Session, elementId int) WidgetStateProvider {
	anim := ReadAnimalIgnoreError(elementId)
	CheckState(anim.Id() != 0, "no animal specified")
	a.cardWidget.SetAnimal(anim)
	return NewStateProvider(a.cardWidget.ChildIdPrefix, anim.ToJson())
}
