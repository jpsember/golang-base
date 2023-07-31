package jt

import (
	. "github.com/jpsember/golang-base/base"
	"strings"
	"sync"
)

type HashCodeRegistry struct {
	Key                string
	Map                JSMap
	registryFileCached Path
	unitTestDirCached  Path
	mutex              sync.RWMutex
}

// Get registry for a test case, constructing one if necessary
//
// Must be thread safe
func (j JTest) registry() *HashCodeRegistry {
	var key = j.Filename
	var registry = sClassesMap.Get(key)
	if registry == nil {
		registry = new(HashCodeRegistry)
		registry.Key = key
		// See if there is a file it was saved to
		registry.Map = JSMapFromFileIfExistsM(registry.file(j))
		registry, _ = sClassesMap.Provide(key, registry)
	}
	return registry
}

func (r *HashCodeRegistry) file(j JTest) Path {
	if r.registryFileCached.Empty() {
		r.registryFileCached = r.unitTestDirectory(j).JoinM(strings.ReplaceAll(r.Key, ".", "_") + ".json")
	}
	return r.registryFileCached
}

func (r *HashCodeRegistry) unitTestDirectory(j JTest) Path {
	if r.unitTestDirCached.Empty() {
		path := j.GetModuleDir().JoinM("unit_test")
		path.MkDirsM()
		r.unitTestDirCached = path
	}
	return r.unitTestDirCached
}

func (r *HashCodeRegistry) VerifyHash(j JTest, currentHash int32, invalidateOldHash bool) bool {
	testName := j.BaseName()
	// Don't let other threads read or write this HashCodeRegistry's map
	r.mutex.Lock()
	var expectedHash = r.Map.OptInt32(testName, 0)
	if expectedHash == 0 || invalidateOldHash {
		r.Map.Put(testName, currentHash)
		r.file(j).WriteStringM(r.Map.String())
		expectedHash = currentHash
	}
	r.mutex.Unlock()
	return currentHash == expectedHash
}

var sClassesMap = NewConcurrentMap[string, *HashCodeRegistry]()
