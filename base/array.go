package base

import (
	"sort"
)

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

func (array *Array[T]) NonEmpty() bool {
	return !array.IsEmpty()
}

func (array *Array[T]) Last() T {
	return array.Get(array.Size() - 1)
}

func (array *Array[T]) Append(items ...T) {
	array.wrappedArray = append(array.wrappedArray, items...)
}

func (array *Array[T]) Get(i int) T {
	return array.wrappedArray[i]
}

func (array *Array[T]) Set(i int, value T) {
	array.wrappedArray[i] = value
}

// Attempt to sort the array
func (array *Array[T]) Sort() error {
	if array.Size() < 2 {
		return nil
	}
	// Not sure why; have to cast argument to 'any'
	a, ok := any(array.wrappedArray).([]string)
	if ok {
		sort.Strings(a)
		return nil
	}
	return Error("Not sortable")
}

func (array *Array[T]) String() string {
	var bp BasePrinter
	bp.AppendString("[")
	for i, x := range array.wrappedArray {
		if i != 0 {
			bp.AppendString(", ")
		}
		bp.Append(x)
	}
	bp.AppendString("]")
	return bp.String()
}
