package jt

import (
	"hash/fnv"
	"math/rand"
	"os/exec"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
)

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
	CheckState(NonEmpty(result), "failed determining unit test filename for:", location)
	return result
}

var unitTestCounter atomic.Int32

func auxNew(t testing.TB) JTest {

	// Ideally we could determine ahead of time how many unit tests are being run in the current session;
	// but there doesn't seem to be a way to do that.  Instead, turn on verbosity iff this is the first
	// JTestStruct object being constructed.

	var testNumber = unitTestCounter.Add(1)

	return &JTestStruct{
		TB:       t,
		Filename: determineUnittestFilename(CallerLocation(2)),
		verbose:  testNumber == 1,
		rand:     NewJSRand().SetSeed(1965),
	}
}

func New(t testing.TB) JTest {
	return auxNew(t)
}

// Deprecated: this constructor will cause the old hash code to be thrown out
//
//goland:noinspection GoUnusedExportedFunction
func Newz(t testing.TB) JTest {
	r := auxNew(t)
	r.verbose = true
	r.InvalidateOldHash = true
	return r
}

type JTestStruct struct {
	testing.TB
	Filename           string
	verbose            bool
	testResultsDir     Path
	unitTestDir        Path
	moduleDir          Path
	generatedDir       Path
	baseNameCached     string
	InvalidateOldHash  bool
	rand               JSRand
	referenceDirCached Path
}
type JTest = *JTestStruct

func (j JTest) Verbose() bool {
	return j.verbose
}

// Get the base name of the test, which is the name of the unit test with the suffix 'Test' removed
func (j JTest) BaseName() string {
	if Empty(j.baseNameCached) {
		var testName = j.TB.Name()
		var baseName, found = strings.CutPrefix(testName, "Test")
		CheckState(found, "Unexpected test name:", testName)
		j.baseNameCached = baseName
	}
	return j.baseNameCached
}

func (j JTest) GetUnitTestDir() Path {
	if j.unitTestDir.Empty() {
		var dir = j.GetModuleDir().JoinM("unit_test")
		dir.MkDirsM()
		j.unitTestDir = dir
	}
	return j.unitTestDir
}

func (j JTest) getGeneratedDir() Path {
	if j.generatedDir.Empty() {
		var genDir = j.GetUnitTestDir().JoinM("generated")
		genDir.MkDirsM()
		j.generatedDir = genDir
	}
	return j.generatedDir
}

func (j JTest) GetTestResultsDir() Path {
	if j.testResultsDir.Empty() {
		var dir = j.getGeneratedDir().JoinM(j.Filename + "/" + j.BaseName())
		// Delete any existing contents of this directory
		// Make sure it contains '/generated/' (pretty sure it does) to avoid crazy deletion
		CheckOk(dir.RemakeDir("/generated/"))
		j.testResultsDir = dir
	}
	return j.testResultsDir
}

func (j JTest) SetVerbose() {
	j.verbose = true
}

func (j JTest) GetModuleDir() Path {
	if j.moduleDir.Empty() {
		j.moduleDir = CheckOkWith(AscendToDirectoryContainingFile("", "go.mod"))
	}
	return j.moduleDir
}

func (j JTest) Log(message ...any) {
	if j.Verbose() {
		Pr(message...)
	}
}

func (j JTest) GenerateMessage(message ...any) {
	var text = ToString(message...)
	j.generateMessageTo("message.txt", text)
}

func (j JTest) generateMessageTo(filename string, content string) {
	if j.Verbose() {
		var q strings.Builder
		for _, s := range strings.Split(content, "\n") {
			q.WriteString("\u21e8")
			q.WriteString(s)
			q.WriteString("\u21e6\n")
		}
		j.Log("Text:", INDENT, q)
	}
	var path = j.GetTestResultsDir().JoinM(filename)
	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	path.WriteStringM(content)
}

func (j JTest) AssertMessage(message ...any) {
	j.generateMessageTo("message.txt", ToString(message...))
	j.AssertGenerated()
}

var hasher = fnv.New32a()

func HashOfJSMap(jsonMap *JSMapStruct) int32 {
	return HashOfString(jsonMap.CompactString())
}

func HashOfString(str string) int32 {
	return HashOfBytes([]byte(str))
}

func HashOfBytes(b []byte) int32 {
	hasher.Reset()
	CheckOkWith(hasher.Write(b))
	return int32((hasher.Sum32()&0xffff)%9000 + 1000)
}

// Construct hash of generated directory, and verify it has the expected value.
func (j JTest) AssertGenerated() {

	var jsonMap = DirSummary(j.GetTestResultsDir())
	var currentHash = HashOfJSMap(jsonMap)
	var registry = j.registry()

	if !registry.VerifyHash(j, currentHash, j.InvalidateOldHash) {
		var summary = ToString("\nUnexpected hash value for directory contents:", CR)
		Pr(summary)
		j.showDiffs()
		j.Fail()
		return
	}
	j.saveTestResults()
}

func (j JTest) FailWithMessage(prefix string, message ...any) {
	Pr(JoinElementToList(prefix, message))
	j.Fail()
}

func (j JTest) AssertTrue(value bool, message ...any) bool {
	if !value {
		j.FailWithMessage("Expression is not true:", message...)
	}
	return value
}

// True asserts that the specified value is true.
func (j JTest) AssertFalse(value bool, message ...any) bool {
	if value {
		j.FailWithMessage("Expression is not false:", message...)
	}
	return value
}

func DirSummary(dir Path) JSMap {
	var jsMap = NewJSMap()
	var w = NewDirWalk(dir).WithDirNames()
	for _, ent := range w.Files() {
		var filename = ent.Base()
		var value any

		value = "?"
		if ent.IsDir() {
			var subdirSummary = DirSummary(dir.JoinM(filename))
			value = subdirSummary
		} else {
			bytes := CheckOkWith(dir.JoinM(filename).ReadBytes())
			value = HashOfBytes(bytes)
		}
		jsMap.Put(filename, value)
	}

	return jsMap
}

// Display diff of generated directory and its reference version
func (j JTest) showDiffs() {

	var refDir = j.referenceDir()
	if !refDir.IsDir() {
		return
	}
	var genDir = j.GetTestResultsDir()
	var relFiles = NewSet[Path]()

	var dirWalk = NewDirWalk(refDir).WithRecurse().OmitNames(`\.DS_Store`)
	relFiles.AddAll(dirWalk.FilesRelative())

	dirWalk = NewDirWalk(genDir).WithRecurse().OmitNames(`\.DS_Store`)
	relFiles.AddAll(dirWalk.FilesRelative())

	for _, fileReceived := range relFiles.Slice() {
		var fileRecAbs = genDir.JoinM(fileReceived.String())
		var fileRefAbs = refDir.JoinM(fileReceived.String())

		if fileRefAbs.Exists() && fileRecAbs.Exists() {
			var refBytes = fileRefAbs.ReadBytesM()
			var recBytes = fileRecAbs.ReadBytesM()
			if reflect.DeepEqual(refBytes, recBytes) {
				continue
			}
		}
		if !j.Verbose() && !Alert("only call this method if verbose") {
			continue
		}

		Pr(CR, DASHES)
		Pr(fileReceived)
		if !fileRefAbs.Exists() {
			Pr("...unexpected file")
			continue
		}
		if !fileRecAbs.Exists() {
			Pr("...file has disappeared")
			continue
		}

		// If it looks like a text file, call the 'diff' utility to display differences.
		// Otherwise, only do this (using binary mode) if in verbose mode
		//
		var ext = fileReceived.Extension()
		var isTextFile = TextFileExtensions.Contains(ext)

		var args = NewArray[string]()
		args.Append("diff")
		if isTextFile {
			args.Add("--text") // "Treat all files as text."
		}
		args.Append("-C", "2", fileRefAbs.String(), fileRecAbs.String())

		output, err := makeSysCall(args.Array())
		_ = err
		Pr(output)
	}
}

var TextFileExtensions = NewStringSet()

func init() {
	TextFileExtensions.AddAll(strings.Split("txt json java go", " "))
}

func makeSysCall(c []string) (string, error) {
	var cmd = c[0]
	var args = c[1:]
	out, err := exec.Command(cmd, args...).Output()
	var strout = string(out)
	return strout, err
}

func (j JTest) Seed(seed int) JTest {
	j.JSRand().SetSeed(seed)
	return j
}

func (j JTest) Rand() *rand.Rand {
	return j.JSRand().Rand()
}

func (j JTest) JSRand() JSRand {
	return j.rand
}

// Generate a directory structure based upon a JSMap script.  The target argument, if not an absolute directory,
// is assumed to be relative to the test's results directory.
// The jsmap has keys representing files or directories.  If the value is a string, it generates a random text file;
// and if it is a jsmap, it generates a directory recursively.
func (j JTest) GenerateSubdirs(target Path, jsmap JSMap) {
	var dir Path
	if target.IsAbs() {
		dir = target
	} else {
		dir = j.GetTestResultsDir().JoinM(target.String())
	}
	j.auxGenDir(dir, jsmap)
}

func (j JTest) auxGenDir(dir Path, jsmap JSMap) {
	dir.MkDirsM()
	for _, entry := range jsmap.Entries() {
		key := entry.Key
		val := entry.Value
		s, ok := val.(JSMap)
		if ok {
			j.auxGenDir(dir.JoinM(key), s)
		} else {
			targ := dir.JoinM(key)
			text := RandomText(j.JSRand(), 80, false) + "\n"
			targ.WriteStringM(text)
		}
	}
}

func (j JTest) referenceDir() Path {
	if j.referenceDirCached.Empty() {
		var g = j.GetTestResultsDir()
		j.referenceDirCached = g.Parent().JoinM(g.Base() + "_REF")
	}
	return j.referenceDirCached
}

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
func (j JTest) saveTestResults() {

	// If we're going to replace the hash in any case, delete any existing reference directory,
	// since its old contents may correspond to an older hash code
	if j.InvalidateOldHash {
		j.referenceDir().DeleteDirectoryM("/generated/")
	}

	var res = j.GetTestResultsDir()

	if !j.referenceDir().Exists() {
		CheckOk(res.MoveTo(j.referenceDir()))
	} else {
		CheckOk(res.DeleteDirectory("unit_test"))
	}
}

// A do-nothing method
func (j JTest) Nothing() {
}

func (j JTest) AssertEqual(a any, b any) JTest {
	if !reflect.DeepEqual(a, b) {
		BadArg("Values are not equal:", INDENT, a, CR, b)
	}
	return j
}
