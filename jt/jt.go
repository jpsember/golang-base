package jt

import (
	"hash/fnv"
	"os"
	"path/filepath"
	"strings"
	"testing"

	. "github.com/jpsember/golang-base/base"
	"github.com/jpsember/golang-base/files"
	. "github.com/jpsember/golang-base/json"
)

var _ = Pr

func determineUnittestFilename(location string) string {
	var result string
	for {
		var i = strings.LastIndexByte(location, ':')
		if i < 0 {
			break
		}
		var s = location[0:i]
		var s2 = strings.TrimSuffix(s, "_test.go")
		if s == s2 {
			break
		}
		if Empty(s2) {
			break
		}
		result = s2
		break
	}
	CheckState(NonEmpty(result), "failed parsing:", location)
	return result
}

func New(t testing.TB) *J {
	return &J{
		TB:       t,
		Filename: determineUnittestFilename(CallerLocation(3)),
	}
}

// Deprecated: this constructor will cause the old hash code to be thrown out
func Newz(t testing.TB) *J {
	return &J{
		TB:                t,
		Filename:          determineUnittestFilename(CallerLocation(3)),
		InvalidateOldHash: true,
	}
}

type J struct {
	testing.TB
	Filename          string
	verbose           bool
	testResultsDir    string
	unitTestDir       string
	moduleDir         string
	baseNameCached    string
	InvalidateOldHash bool
}

func (j *J) Verbose() bool {
	return j.verbose
}

// Get the base name of the test, which is the name of the unit test with the suffix 'Test' removed
func (j *J) BaseName() string {
	if Empty(j.baseNameCached) {
		var testName = j.TB.Name()
		var baseName, found = strings.CutPrefix(testName, "Test")
		CheckState(found, "Unexpected test name:", testName)
		j.baseNameCached = baseName
	}
	return j.baseNameCached
}

func (j *J) GetUnitTestDir() string {
	if Empty(j.unitTestDir) {
		var dir = filepath.Join(j.GetModuleDir(), "unit_test")
		var err = os.MkdirAll(dir, os.ModePerm)
		CheckOk(err)
		j.unitTestDir = dir
	}
	return j.unitTestDir
}

func (j *J) GetTestResultsDir() string {
	if Empty(j.testResultsDir) {
		var genDir = filepath.Join(j.GetUnitTestDir(), "generated")
		{
			var err = os.MkdirAll(genDir, os.ModePerm)
			CheckOk(err)
		}

		var testResultsDir = filepath.Join(genDir, j.Filename, j.BaseName())
		// Delete any existing contents of this directory
		// Make sure it contains '/generated/' (pretty sure it does) to avoid crazy deletion
		files.DeleteDir(testResultsDir, "/generated/")
		{
			var err = os.MkdirAll(testResultsDir, os.ModePerm)
			CheckOk(err)
		}

		j.testResultsDir = testResultsDir
	}
	return j.testResultsDir
}

func (j *J) SetVerbose() {
	j.verbose = true
}

func (j *J) GetModuleDir() string {
	if Empty(j.moduleDir) {
		var path, err = files.AscendToDirectoryContainingFile("", "go.mod")
		CheckOk(err)
		j.moduleDir = path.String()
	}
	return j.moduleDir
}

func (j *J) Log(message ...any) {
	if j.Verbose() {
		Pr(message...)
	}
}

func (j *J) GenerateMessage(message ...any) {
	var text = ToString(message)

	j.generateMessageTo("message.txt", text)
}

func (j *J) generateMessageTo(filename string, content string) {
	if j.Verbose() {
		var q strings.Builder
		for _, s := range strings.Split(content, "\n") {
			q.WriteString("\u21e8")
			q.WriteString(s)
			q.WriteString("\u21e6\n")
		}
		j.Log("Content:", INDENT, q)
	}
	var path = filepath.Join(j.GetTestResultsDir(), filename)
	files.WriteString(path, content)
}

func (j *J) AssertMessage(message ...any) {
	j.generateMessageTo("message.txt", ToString(message...))
	j.AssertGenerated()
}

var hasher = fnv.New32a()

func HashOfJSMap(jsonMap *JSMap) int32 {
	return HashOfString(PrintJSEntity(jsonMap, false))
}

func HashOfString(str string) int32 {
	return HashOfBytes([]byte(str))
}

func HashOfBytes(b []byte) int32 {
	hasher.Reset()
	hasher.Write(b)
	return int32((hasher.Sum32()&0xffff)%9000 + 1000)
}

// Construct hash of generated directory, and verify it has the expected value.
func (j *J) AssertGenerated() {

	var jsonMap = dirSummary(j.GetTestResultsDir())
	var currentHash = HashOfJSMap(jsonMap)
	// /**
	//   - Construct hash of generated directory, and verify it has the expected value
	//     */
	//     public void assertGeneratedDirectoryHash() {
	//     if (mUnitTest.verbose())
	//     createInspectionDir();
	//     try {
	//     JSMap jsonMap = MyTestUtils.dirSummary(generatedDir());
	//     // Convert hash code to one using exactly four digits
	//     int currentHash = (jsonMap.HashOfJSMap() & 0xffff) % 9000 + 1000;
	var registry = RegistryFor(j)

	if !registry.VerifyHash(j.Name(), currentHash, j.InvalidateOldHash) {
		var summary = ToString("\nUnexpected hash value for directory contents:", CR, DASHES, CR)
		Pr(summary)
		Todo("showDiffs")
		//showDiffs()
		j.Fail()
	}
	registry.SaveTestResults()
}

func dirSummary(dir string) *JSMap {
	return auxDirSummary(dir, true)
}

func auxDirSummary(dir string, calculateFileHashes bool) *JSMap {
	var jsMap = NewJSMap()

	var entries, err = os.ReadDir(dir)
	CheckOk(err)

	for _, ent := range entries {

		var filename = ent.Name()
		var value any

		value = "?"
		if ent.IsDir() {
			var subdirSummary = auxDirSummary(filepath.Join(dir, filename), calculateFileHashes)
			Todo("have JSMap.size or empty method")
			value = subdirSummary
		} else if calculateFileHashes {
			var bytes []byte
			bytes, err = files.ReadBytes(filepath.Join(dir, filename))
			CheckOk(err)
			value = HashOfBytes(bytes)
		}
		jsMap.Put(filename, value)
	}

	return jsMap
}
