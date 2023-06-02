package jt

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/files"
	. "github.com/jpsember/golang-base/json"
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
	InvalidateOldHash  bool
	UnitTest           *J
	referenceDirCached Path
	UnitTestName       string
}

// Get registry for a test case, constructing one if necessary
//
// Must be thread safe
func (j *J) registry() *HashCodeRegistry {
	var key = j.Filename
	var registry = sClassesMap[key]
	if registry == nil {
		registry = new(HashCodeRegistry)
		registry.UnitTest = j
		registry.Key = key

		Todo("I think I need to lock the mutex for reading as well")
		// Don't let other threads modify the map while we are modifying it or creating the registry's jsmap
		mutex.Lock()
		sClassesMap[key] = registry
		// See if there is a file it was saved to
		registry.Map = JSMapFromFileIfExistsM(registry.file())
		mutex.Unlock()

		registry.UnitTestName = strings.TrimPrefix(j.TB.Name(), "Test")
	}
	return registry
}

func (r *HashCodeRegistry) file() Path {
	if r.registryFileCached.Empty() {
		r.registryFileCached = r.unitTestDirectory().JoinM(strings.ReplaceAll(r.Key, ".", "_") + ".json")
	}
	return r.registryFileCached
}

func (r *HashCodeRegistry) unitTestDirectory() Path {
	if r.unitTestDirCached.Empty() {
		path := r.UnitTest.GetModuleDir().JoinM("unit_test")
		path.MkDirsM()
		r.unitTestDirCached = path
	}
	return r.unitTestDirCached
}

func (r *HashCodeRegistry) VerifyHash(testName string, currentHash int32, invalidateOldHash bool) bool {
	var expectedHash = r.Map.OptInt32(testName, 0)
	if expectedHash == 0 || invalidateOldHash {
		// Don't let other threads modify or write the map
		mutex.Lock()
		r.Map.Put(testName, currentHash)
		r.write()
		mutex.Unlock()
		expectedHash = currentHash
	}
	return currentHash == expectedHash
}

func (r *HashCodeRegistry) write() {
	var path = r.file()
	var content = r.Map.String()
	path.WriteStringM(content)
}

var sClassesMap = make(map[string]*HashCodeRegistry)

/**
 * Called when the generated directory's hash has been successfully verified.
 *
 * 1) If a 'reference' copy of the directory doesn't exist, move generated
 * directory as it; otherwise, delete the generated directory (since it is the
 * same as the reference copy)
 *
 * 2) Update the hash code of the directory, if it differs from the previous
 * value (or no previous value exists).
 */
func (r *HashCodeRegistry) SaveTestResults() {

	// If we're going to replace the hash in any case, delete any existing reference directory,
	// since its old contents may correspond to an older hash code
	if r.InvalidateOldHash {
		r.referenceDir().DeleteDirectoryM("/generated/")
	}

	var res = r.UnitTest.GetTestResultsDir()

	if !r.referenceDir().Exists() {
		err := res.MoveTo(r.referenceDir())
		CheckOk(err)
	} else {
		err := res.DeleteDirectory("unit_test")
		CheckOk(err)
	}
}

func (r *HashCodeRegistry) referenceDir() Path {
	if r.referenceDirCached.Empty() {
		var g = r.UnitTest.GetTestResultsDir()
		r.referenceDirCached = g.Parent().JoinM(g.Base() + "_REF")
	}
	return r.referenceDirCached
}
