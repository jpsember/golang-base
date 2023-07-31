package base

import (
	"sync"
)

// A thread-safe map implementation.
type ConcurrentMap[K comparable, V any] struct {
	wrappedMap map[K]V
	lock       sync.RWMutex
}

func NewConcurrentMap[K comparable, V any]() *ConcurrentMap[K, V] {
	r := ConcurrentMap[K, V]{
		wrappedMap: make(map[K]V),
	}
	return &r
}

// Get value for key, returning i) default value if key doesn't exist, ii) whether it existed
func (m *ConcurrentMap[K, V]) OptValue(key K, defaultValue V) (result V, ok bool) {
	m.lock.RLock()
	val, ok := m.wrappedMap[key]
	if !ok {
		val = defaultValue
	}
	m.lock.RUnlock()
	return val, ok
}

func (m *ConcurrentMap[K, V]) Get(key K) V {
	m.lock.RLock()
	val := m.wrappedMap[key]
	m.lock.RUnlock()
	return val
}

func (m *ConcurrentMap[K, V]) Put(key K, value V) V {
	m.lock.Lock()
	oldValue := m.wrappedMap[key]
	m.wrappedMap[key] = value
	m.lock.Unlock()
	return oldValue
}

// Store key/value pair if key doesn't already exist.  Returns old value if it existed, else new one; and true if already existed.
func (m *ConcurrentMap[K, V]) Provide(key K, value V) (V, bool) {
	m.lock.Lock()
	oldValue, ok := m.wrappedMap[key]
	if !ok {
		m.wrappedMap[key] = value
		oldValue = value
	}
	m.lock.Unlock()
	return oldValue, ok
}
