package jt

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/files"
	. "github.com/jpsember/golang-base/json"
	"strings"
)

var _ = Pr

type HashCodeRegistry struct {
	Key               string
	Map               *JSMap
	_file             Path
	_dir              Path
	InvalidateOldHash bool
	UnitTest      *J
	_referenceDir Path
	UnitTestName  string
}

// Get registry for a test case, constructing one if necessary
//
// (must be thread safe?)
func RegistryFor(j *J) *HashCodeRegistry {
	var key = j.Filename
	var registry = sClassesMap[key]
	if registry == nil {
		registry = new(HashCodeRegistry)
		registry.UnitTest = j
		registry.Key = key
		sClassesMap[key] = registry
		// See if there is a file it was saved to
		registry.Map = JSMapFromFileIfExists(registry.file())
		registry.UnitTestName = strings.TrimPrefix(j.TB.Name(), "Test")
	}
	return registry
}

func (r *HashCodeRegistry) file() Path {
	if r._file.Empty() {
		r._file = r.unitTestDirectory().JoinM(strings.ReplaceAll(r.Key, ".", "_") + ".json")
	}
	return r._file
}

func (r *HashCodeRegistry) unitTestDirectory() Path {
	if r._dir.Empty() {
		path := r.UnitTest.GetModuleDir().JoinM("unit_test")
		path.MkDirsM()
		r._dir = path
	}
	return r._dir
}

func (r *HashCodeRegistry) VerifyHash(testName string, currentHash int32, invalidateOldHash bool) bool {
	var expectedHash = r.Map.OptInt32(testName, 0)
	if expectedHash == 0 || invalidateOldHash {
		Todo("synchronize access to map?")
		r.Map.Put(testName, currentHash)
		r.write()
		expectedHash = currentHash
	}
	return currentHash == expectedHash
}

func (r *HashCodeRegistry) write() {
	var path = r.file()
	var content = PrintJSEntity(r.Map, true)
	path.WriteString(content)
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
		r.referenceDir().DeleteDirectory("/generated/")
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
  Todo("get rid of underscores")
	if r._referenceDir.Empty() {
		var g = r.UnitTest.GetTestResultsDir()
		r._referenceDir = g.Parent().JoinM(g.Base() + "_REF")
	}
	return r._referenceDir
}
