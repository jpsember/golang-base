package base

var _ = Pr

type Set[KEY comparable] struct {
	wrappedMap map[KEY]bool
}

func NewSet[KEY comparable]() *Set[KEY] {
	Todo("!Make Set follow ptr/struct pattern")
	m := new(Set[KEY])
	m.Clear()
	return m
}

// Add element to set.  Return true if it was not already in the set.
func (set *Set[KEY]) Add(value KEY) bool {
	found := set.Contains(value)
	if !found {
		set.wrappedMap[value] = true
	}
	return !found
}

func (set *Set[KEY]) Clear() {
	set.wrappedMap = make(map[KEY]bool)
}

func (set *Set[KEY]) Contains(value KEY) bool {
	_, found := set.wrappedMap[value]
	return found
}

func (set *Set[KEY]) AddAll(slice []KEY) {
	for _, v := range slice {
		set.Add(v)
	}
}

func (set *Set[KEY]) Slice() []KEY {
	arr := NewArray[KEY]()
	for v, _ := range set.wrappedMap {
		arr.Add(v)
	}
	return arr.Array()
}

func (set *Set[KEY]) Size() int {
	return len(set.wrappedMap)
}

func (set *Set[KEY]) WrappedMap() map[KEY]bool {
	return set.wrappedMap
}
