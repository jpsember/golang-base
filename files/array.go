package files

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/json"
)

var _ = Pr

type Array[T any] struct {
	wrappedArray []T
}

func NewArray[T any]() *Array[T] {
	m := new(Array[T])
	// Make a slice that has length zero, but whose underlying array has a capacity of 10
	m.wrappedArray = make([]T, 0, 10)
	return m
}

func (array *Array[T]) Size() int { return len(array.wrappedArray) }

func (array *Array[T]) Add(value T) {
	array.wrappedArray = append(array.wrappedArray, value)
}

func (array *Array[T]) Array() []T {
	return array.wrappedArray
}

func (array *Array[T]) IsEmpty() bool {
	return len(array.wrappedArray) == 0
}

func (array *Array[T]) Pop() T {
	var w = array.wrappedArray
	i := len(w)
	if i == 0 {
		BadStateWithSkip(1, "Pop of empty array")
	}
	var result = w[i-1]
	array.wrappedArray = w[:i-1]
	return result
}

func (array *Array[T]) ToJson() *JSMap {
	var m *JSMap = NewJSMap()
	m.Put("", "Array")
	m.Put("cap", cap(array.wrappedArray))
	m.Put("len", len(array.wrappedArray))
	m.Put("size", len(array.wrappedArray))
	var lst = NewJSList()
	for _, x := range array.wrappedArray {
		{
			lst.Add(ToString(x))
		}
	}
	m.Put("[]", lst)
	return m
}
