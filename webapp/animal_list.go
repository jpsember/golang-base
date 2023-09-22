// Implementation of ListInterface for lists of animal cards.
package webapp

import (
	. "github.com/jpsember/golang-base/base"
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
	if Experiment {
		animalIds = ClampedSlice(animalIds, 0, 1)
	}
	b.ElementIds = animalIds
	b.ElementsPerPage = 12
	return t
}

func (a AnimalList) ItemStateProvider(s Session, elementId int) WidgetStateProvider {
	anim := ReadAnimalIgnoreError(elementId)
	CheckState(anim.Id() != 0, "no animal available")
	a.cardWidget.SetAnimal(anim)
	return NewStateProvider(a.cardWidget.ChildIdPrefix, anim.ToJson())
}
