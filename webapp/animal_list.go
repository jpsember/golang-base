// Implementation of ListInterface for lists of animal cards.
package webapp

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/webserv"
)

type AnimalListStruct struct {
	BasicListStruct
	cardWidget AnimalCard
	itemPrefix string
}

type AnimalList = *AnimalListStruct

func NewAnimalList(animalIds []int, cardWidget AnimalCard, itemPrefix string) AnimalList {
	CheckArg(cardWidget != nil)
	CheckArg(animalIds != nil)
	t := &AnimalListStruct{
		cardWidget: cardWidget,
		itemPrefix: itemPrefix,
	}
	b := &t.BasicListStruct
	if Experiment {
		animalIds = ClampedSlice(animalIds, 0, 2)
	}
	b.ElementIds = animalIds
	b.ElementsPerPage = 12
	return t
}

func (a AnimalList) ItemStateProvider(s Session, elementId int) WidgetStateProvider {
	anim := ReadAnimalIgnoreError(elementId)
	CheckState(anim.Id() != 0, "no animal available")

	json := anim.ToJson().AsJSMap()
	return NewStateProvider(a.ItemPrefix(), json)
}

func (a AnimalList) ItemPrefix() string {
	return a.itemPrefix
}
