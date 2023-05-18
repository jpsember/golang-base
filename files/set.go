package files

import (
	. "github.com/jpsember/golang-base/base"
)

var _ = Pr

type Set[KEY comparable] struct {
	wrappedMap map[KEY]bool
}

func NewSet[KEY comparable]() *Set[KEY] {
	m := new(Set[KEY])
	m.wrappedMap = make(map[KEY]bool)
	return m
}

func (set *Set[KEY]) Add(value KEY) bool {
	found := set.Contains(value)
	if !found {
		set.wrappedMap[value] = true
	}
	return !found
}

func (set *Set[KEY]) Contains(value KEY) bool {
	_, found := set.wrappedMap[value]
	return found
}
