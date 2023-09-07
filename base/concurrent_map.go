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
	r := ConcurrentMap[K, V]{}
	r.Clear()
	return &r
}

func NewConcurrentMapWith[K comparable, V any](contents map[K]V) *ConcurrentMap[K, V] {
	r := ConcurrentMap[K, V]{}
	r.wrappedMap = contents
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

func (m *ConcurrentMap[K, V]) GetAll() ([]K, []V) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return GetMapKeysAndValues(m.wrappedMap)
}

func GetMapKeysAndValues[K comparable, V any](mp map[K]V) ([]K, []V) {
	ln := len(mp)
	keys := make([]K, ln)
	values := make([]V, ln)
	i := 0
	for k, v := range mp {
		keys[i] = k
		values[i] = v
		i++

		var q any
		q = v

		jv, ok := q.(DataClass)
		if ok {
			m := jv.ToJson().AsJSMap()
			m.Delete("data")
		}

	}

	return keys, values
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

func (m *ConcurrentMap[K, V]) Clear() {
	m.lock.Lock()
	m.wrappedMap = make(map[K]V)
	m.lock.Unlock()
}

func (m *ConcurrentMap[K, V]) Size() int {
	m.lock.RLock()
	result := len(m.wrappedMap)
	m.lock.RUnlock()
	return result
}
