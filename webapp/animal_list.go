package webapp

import (
	. "github.com/jpsember/golang-base/base"
)

type AnimalListStruct struct {
	elements []int
	Page     int
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

func (a AnimalList) CurrentPage() int {
	Todo("Can the SetPage/CurrentPage be handled by an underlying basic page?")
	return a.Page
}
func (a AnimalList) TotalPages() int {
	numElements := len(a.elements)
	remainder := numElements % a.ElementsPerPage()
	completePages := numElements / a.ElementsPerPage()
	totalPages := completePages
	if remainder != 0 {
		totalPages++
	}
	return MaxInt(1, totalPages)
}

func (a AnimalList) SetPage(page int) {
	a.Page = page
}
