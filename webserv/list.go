// A datatype that represents a list of items, that is displayable a page at a time, with controls
// to jump to next/previous pages and whatnot.

package webserv

import (
	. "github.com/jpsember/golang-base/base"
)

type ListInterface interface {
	GetPageElements() []int
	CurrentPage() int
	TotalPages() int
	SetCurrentPage(pageNumber int)

	// ItemStateProvider constructs a state provider xi for rendering item i.
	// Child widgets within the item widget that already have explicit state providers
	// will *not* use Xi.
	ItemStateProvider(s Session, elementId int) WidgetStateProvider
}

type ListItemRenderer func(session Session, widget ListWidget, elementId int, m MarkupBuilder)

type BasicListStruct struct {
	ElementsPerPage int
	ElementIds      []int
	currentPage     int
}

type BasicList = *BasicListStruct

func (b BasicList) GetPageElements() []int {
	k := b.ElementsPerPage
	pgStart := b.CurrentPage() * k
	pgEnd := pgStart + k
	return ClampedSlice(b.ElementIds, pgStart, pgEnd)
}

func (b BasicList) CurrentPage() int {
	return b.currentPage
}

func (b BasicList) SetCurrentPage(pageNumber int) {
	j := Clamp(pageNumber, 0, b.TotalPages()-1)
	if j != pageNumber {
		BadArg("Attempt to set current page to", pageNumber, "; total is:", b.TotalPages())
	}
	b.currentPage = pageNumber
}

func (b BasicList) TotalPages() int {
	CheckState(b.ElementsPerPage > 0, "no ElementsPerPage")
	numElements := len(b.ElementIds)
	remainder := numElements % b.ElementsPerPage
	completePages := numElements / b.ElementsPerPage
	if numElements == 0 || remainder > 0 {
		completePages++
	}
	return completePages
}

// A default state provider.
func (b BasicList) ItemStateProvider(s Session, elementId int) WidgetStateProvider {
	BadState("No ItemStateProvider available")
	return nil
}
