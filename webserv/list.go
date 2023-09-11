// A datatype that represents a list of items, that is displayable a page at a time, with controls
// to jump to next/previous pages and whatnot.

package webserv

type ListInterface interface {
	ElementsPerPage() int
	GetPageElements(pageNumber int) []int
}

type ListItemRenderer func(widget ListWidget, elementId int, m MarkupBuilder)
