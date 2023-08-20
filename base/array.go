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

func (array *Array[T]) Clear() {
	array.wrappedArray = make([]T, 0)
}

func (array *Array[T]) Pop() T {
	var w = array.wrappedArray
	i := len(w)
	if i == 0 {
		BadState("<1 Pop of empty array")
	}
	var result = w[i-1]
	array.wrappedArray = w[:i-1]
	return result
}

func (array *Array[T]) NonEmpty() bool {
	return !array.IsEmpty()
}

func (array *Array[T]) First() T {
	return array.Get(0)
}

func (array *Array[T]) Last() T {
	return array.Get(array.Size() - 1)
}

// Remove a contiguous sequence of elements; adjust arguments into range, and do nothing if appropriate.
func (array *Array[T]) Remove(start int, count int) {
	w := array.wrappedArray
	end := start + count
	x := len(w)
	start = Clamp(start, 0, x)
	end = Clamp(end, start, x)
	if start < end {
		m := w[0:start]
		m = append(m, w[end:x]...)
		array.wrappedArray = m
	}
}

// Remove all elements at or beyond a particular position; adjust arguments
// into range, and do nothing if appropriate.
func (array *Array[T]) RemoveAllButFirstN(n int) {
	array.Remove(n, array.Size())
}

// Remove all elements except the last n, doing nothing if there are <= n elements.
func (array *Array[T]) RemoveAllButLastN(n int) {
	array.Remove(0, array.Size()-n)
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
