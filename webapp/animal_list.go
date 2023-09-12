package webapp

import (
	. "github.com/jpsember/golang-base/base"
	"github.com/jpsember/golang-base/webserv"
)

type AnimalListStruct struct {
	webserv.BasicListStruct
}

type AnimalList = *AnimalListStruct

func NewAnimalList(animalIds []int) AnimalList {
	CheckArg(animalIds != nil)
	t := &AnimalListStruct{}
	b := &t.BasicListStruct
	b.ElementIds = animalIds
	b.ElementsPerPage = 12
	return t
}
