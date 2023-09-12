package webapp

import (
	. "github.com/jpsember/golang-base/base"
)

type AnimalListStruct struct {
	elements []int
}

type AnimalList = *AnimalListStruct

func NewAnimalList() AnimalList {
	t := &AnimalListStruct{}
	return t
}
func (a AnimalList) ElementsPerPage() int {
	return 12
}

func (a AnimalList) GetPageElements(pageNumber int) []int {
	k := a.ElementsPerPage()
	pgStart := pageNumber * k
	pgEnd := pgStart + k

	return ClampedSlice(a.elements, pgStart, pgEnd)
}
