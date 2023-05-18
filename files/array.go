package files

import (
	. "github.com/jpsember/golang-base/base"
)

var _ = Pr

type Array[T any] struct {
	wrappedArray []T
	size         int
}

func NewArray[T any]() *Array[T] {
	m := new(Array[T])
	m.wrappedArray = make([]T, 10)
	return m
}

func (array *Array[T]) Size() int { return array.size }

func (array *Array[T]) Add(value T) {
	var i = array.size
	var sl = array.wrappedArray
	if cap(sl) <= i {
		sl = append(sl, value)
		array.wrappedArray = sl
	} else {
		sl[i] = value
	}
	array.size = i + 1
}

func (array *Array[T]) Array() []T {
	return array.wrappedArray[:array.size]
}

func (array *Array[T]) IsEmpty() bool {
	return array.size == 0
}

func (array *Array[T]) Pop() T {
	if array.IsEmpty() {
		BadStateWithSkip(1, "Pop of empty array")
	}
	var w = array.wrappedArray
	i := array.size - 1
	var result = w[i]
	array.size--
	return result
}
