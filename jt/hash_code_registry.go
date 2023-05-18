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
	_generatedDir     Path
	UnitTest          *J
	_referenceDir     Path
	UnitTestName      string
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
		Halt("about to delete directory:", r.referenceDir())
		r.referenceDir().DeleteDirectory("/generated/")
	}

	if !r.referenceDir().Exists() {
		Halt("about to move:", r.GeneratedDir(), "to:", r.referenceDir())
		err := r.GeneratedDir().MoveTo(r.referenceDir())
		CheckOk(err)
	} else {
		Halt("claims reference dir doesn't exist:", r.referenceDir())

		err := r.GeneratedDir().DeleteDirectory("unit_test")
		CheckOk(err)
	}

}

func (r *HashCodeRegistry) GeneratedDir() Path {

	if r._generatedDir.Empty() {
		var unitTestDir = r.unitTestDirectory()

		// If no .gitignore file exists, create one (creating the directory as well if necessary);
		// it will have the entry GENERATED_DIR_NAME

		var GENERATED_DIR_NAME = "generated"
		var gitIgnoreFile = unitTestDir.JoinM(".gitignore")
		if !gitIgnoreFile.Exists() {
			gitIgnoreFile.WriteStringM(GENERATED_DIR_NAME + "\n")
		}

		var projectDir = unitTestDir.JoinM(GENERATED_DIR_NAME)
		var className = strings.TrimSuffix(r.UnitTest.Filename, "_test.go")
		var testName = r.UnitTestName
		//	strings.TrimSuffix(r.Key, "Test")
		r._generatedDir = projectDir.JoinM(className + "/" + testName)
		Pr(DASHES, CR, "Remaking generated dir")
		CheckOk(r._generatedDir.RemakeDir("unit_test"))
	}
	return r._generatedDir
}

func (r *HashCodeRegistry) referenceDir() Path {
	if r._referenceDir.Empty() {
		//Halt("about to construct ref dir:")
		r._referenceDir = r.GeneratedDir().Parent().JoinM(r.GeneratedDir().Base() + "_REF")
		Halt("built ref dir:", r._referenceDir)
	}
	return r._referenceDir
}
