// Implementation of ListInterface for lists of animal cards.
package webapp

import (
	. "github.com/jpsember/golang-base/base"
	"github.com/jpsember/golang-base/webapp/gen/webapp_data"
	. "github.com/jpsember/golang-base/webserv"
)

type AnimalListStruct struct {
	BasicListStruct
	PrepareState func(s Session, animal webapp_data.Animal)
	statePrefix  string
}

type AnimalList = *AnimalListStruct

func NewAnimalList(animalIds []int, statePrefix string) AnimalList {
	CheckArg(animalIds != nil)
	t := &AnimalListStruct{
		statePrefix: statePrefix,
	}
	b := &t.BasicListStruct
	b.ElementIds = animalIds
	b.ElementsPerPage = 12
	return t
}

func (a AnimalList) ItemStateProvider(s Session, elementId int) WidgetStateProvider {
	anim := ReadAnimalIgnoreError(elementId)
	CheckState(anim.Id() != 0, "no animal specified")

	CheckState(a.PrepareState != nil)
	a.PrepareState(s, anim)

	Todo("we need some way to get the Card's childIdPrefix to this state provider")
	Todo("We need a callback to allow client to plug the animal into the card that is the widget item")
	childIdPrefix := a.statePrefix
	return NewStateProvider(childIdPrefix, anim.ToJson().AsJSMap())
}
