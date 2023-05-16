package jt

import (
	. "github.com/jpsember/golang-base/base"
	"github.com/jpsember/golang-base/files"
	. "github.com/jpsember/golang-base/json"
	"strings"
)

var _ = Pr

type HashCodeRegistry struct {
	Key               string
	Map               *JSMap
	_file             string
	_dir              string
	InvalidateOldHash bool
	_generatedDir     string
	UnitTest          *J
	_referenceDir     string
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
	}
	return registry
}

func (r *HashCodeRegistry) file() string {
	if r._file == "" {
		r._file = r.unitTestDirectory() + "/" + strings.ReplaceAll(r.Key, ".", "_") + ".json"
	}
	return r._file
}

func (r *HashCodeRegistry) unitTestDirectory() string {
	if r._dir == "" {
		d, err := files.MkDirs(files.PathJoin(r.UnitTest.GetModuleDir(), "unit_test"))
		CheckOk(err, "failed to make directory")
		r._dir = d
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
	var file = r.file()
	var content = PrintJSEntity(r.Map, true)
	files.WriteString(file, content)
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
		Todo("Delete old reference directory.........  CAREFULLY")
		//Files.S.deleteDirectory(referenceDir());
	}

	if !files.DirExists(r.referenceDir()) {
		Todo("files.MoveDirectory")
		//files.MoveDirectory(r.GeneratedDir(), r.referenceDir())
	} else {
		Todo("Delete generated dir, same as reference")
		//files.DeleteDirectory(r.generatedDir())
	}

}

func (r *HashCodeRegistry) GeneratedDir() string {

	if r._generatedDir == "" {
		var unitTestDir = r.unitTestDirectory()

		// If no .gitignore file exists, create one (creating the directory as well if necessary);
		// it will have the entry GENERATED_DIR_NAME

		var GENERATED_DIR_NAME = "generated"
		var gitIgnoreFile = unitTestDir + "/.gitignore"
		if !files.Exists(gitIgnoreFile) {
			files.WriteString(gitIgnoreFile, GENERATED_DIR_NAME+"\n")
		}

		var projectDir = files.PathJoin(unitTestDir, GENERATED_DIR_NAME)
		var className = strings.TrimSuffix(r.UnitTest.Filename, "_test.go")
		var testName = strings.TrimSuffix(r.Key, "Test")
		r._generatedDir = files.PathJoin(projectDir, className+"/"+testName)
		Todo("remake dirs")
		//files.RemakeDirs(r._generatedDir)
	}
	return r._generatedDir
}

func (r *HashCodeRegistry) referenceDir() string {
	if r._referenceDir == "" {
		r._referenceDir = files.PathJoin(files.Parent(r.GeneratedDir()), files.GetName(r.GeneratedDir())+"_REF")
	}
	return r._referenceDir
}
