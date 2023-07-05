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
	//UnitTest           *J
	//UnitTestName       string
	Key                string
	Map                *JSMapStruct
	registryFileCached Path
	unitTestDirCached  Path
	referenceDirCached Path
}

// Get registry for a test case, constructing one if necessary
//
// Must be thread safe
func (j *J) registry() *HashCodeRegistry {
	var key = j.Filename
	Todo("If multiple threads are using the same registry, that's a problem")
	mutex.Lock()
	var registry = sClassesMap[key]

	if registry == nil {
		registry = new(HashCodeRegistry)
		registry.Key = key

		// Don't let other threads modify the map while we are modifying it or creating the registry's jsmap
		sClassesMap[key] = registry
		// See if there is a file it was saved to
		registry.Map = JSMapFromFileIfExistsM(registry.file(j))

	}

	//// Copy some values from the unit test to the registry... though this duplication has already caused one
	//// tricky bug
	//registry.UnitTest = j
	//registry.UnitTestName = strings.TrimPrefix(j.Name(), "Test")

	mutex.Unlock()
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

func (r *HashCodeRegistry) VerifyHash(j *J, testName string, currentHash int32, invalidateOldHash bool) bool {
	Todo("redundant to include testName as well as J")
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
func (r *HashCodeRegistry) SaveTestResults(j *J) {
	// If we're going to replace the hash in any case, delete any existing reference directory,
	// since its old contents may correspond to an older hash code
	if j.InvalidateOldHash {
		r.referenceDir(j).DeleteDirectoryM("/generated/")
	}

	var res = j.GetTestResultsDir()

	if Alert("bad names") {
		testName := j.BaseName()
		Pr("testName:", testName)
	}

	Pr(r.referenceDir(j).Info("SaveTestResults, reference dir"))

	if !r.referenceDir(j).Exists() {
		Todo("This sometimes fails due to our unit tests not being threadsafe")
		err := res.MoveTo(r.referenceDir(j))
		CheckOk(err)
	} else {
		Pr(res.Info("SaveTestResults, reference dir already exists"))
		err := res.DeleteDirectory("unit_test")
		CheckOk(err)
	}
}

func (r *HashCodeRegistry) referenceDir(j *J) Path {
	Todo("reference dir, other things should be part of J")
	if r.referenceDirCached.Empty() {
		var g = j.GetTestResultsDir()
		r.referenceDirCached = g.Parent().JoinM(g.Base() + "_REF")
	}
	return r.referenceDirCached
}
