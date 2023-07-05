package jt

import (
	. "github.com/jpsember/golang-base/base"
	"strings"
	"sync"
)

var _ = Pr

var mutex sync.RWMutex

type HashCodeRegistry struct {
	Key                string
	Map                *JSMapStruct
	registryFileCached Path
	unitTestDirCached  Path
}

// Get registry for a test case, constructing one if necessary
//
// Must be thread safe
func (j *J) registry() *HashCodeRegistry {
	var key = j.Filename
	mutex.RLock()
	var registry = sClassesMap[key]
	mutex.RUnlock()
	if registry == nil {
		mutex.Lock()
		registry = new(HashCodeRegistry)
		registry.Key = key
		// Don't let other threads modify the map while we are modifying it or creating the registry's jsmap
		sClassesMap[key] = registry
		// See if there is a file it was saved to
		registry.Map = JSMapFromFileIfExistsM(registry.file(j))
		mutex.Unlock()
	}
	return registry
}

func (r *HashCodeRegistry) file(j *J) Path {
	if r.registryFileCached.Empty() {
		r.registryFileCached = r.unitTestDirectory(j).JoinM(strings.ReplaceAll(r.Key, ".", "_") + ".json")
	}
	return r.registryFileCached
}

func (r *HashCodeRegistry) unitTestDirectory(j *J) Path {
	if r.unitTestDirCached.Empty() {
		path := j.GetModuleDir().JoinM("unit_test")
		path.MkDirsM()
		r.unitTestDirCached = path
	}
	return r.unitTestDirCached
}

func (r *HashCodeRegistry) VerifyHash(j *J, currentHash int32, invalidateOldHash bool) bool {
	testName := j.BaseName()
	var expectedHash = r.Map.OptInt32(testName, 0)
	if expectedHash == 0 || invalidateOldHash {
		// Don't let other threads modify or write the map
		mutex.Lock()
		r.Map.Put(testName, currentHash)
		r.write(j)
		mutex.Unlock()
		expectedHash = currentHash
	}
	return currentHash == expectedHash
}

func (r *HashCodeRegistry) write(j *J) {
	var path = r.file(j)
	var content = r.Map.String()
	path.WriteStringM(content)
}

var sClassesMap = make(map[string]*HashCodeRegistry)
