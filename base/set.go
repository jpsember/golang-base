package base

var _ = Pr

type Set[KEY comparable] struct {
	wrappedMap map[KEY]bool
	locked     bool
}

func NewSet[KEY comparable]() *Set[KEY] {
	m := new(Set[KEY])
	m.Clear()
	return m
}

func (set *Set[KEY]) Lock() {
	set.locked = true
}

// Add element to set.  Return true if it was not already in the set.
func (set *Set[KEY]) Add(value KEY) bool {
	found := set.Contains(value)
	if !found {
		set.mutableWrappedMap()[value] = true
	}
	return !found
}

func (set *Set[KEY]) Clear() {
	set.mutableWrappedMap()
	set.wrappedMap = make(map[KEY]bool)
}

func (set *Set[KEY]) Contains(value KEY) bool {
	_, found := set.wrappedMap[value]
	return found
}

func (set *Set[KEY]) AddAll(slice []KEY) {
	k := set.mutableWrappedMap()
	for _, v := range slice {
		k[v] = true
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

func (set *Set[KEY]) mutableWrappedMap() map[KEY]bool {
	if set.locked {
		BadState("<2set is locked")
	}
	return set.wrappedMap
}
