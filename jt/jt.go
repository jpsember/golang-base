package jt

import (
	. "github.com/jpsember/golang-base/base"
	"github.com/jpsember/golang-base/data"
	. "github.com/jpsember/golang-base/files"
	. "github.com/jpsember/golang-base/json"
	"hash/fnv"
	"math/rand"
	"os/exec"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
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

var unitTestCounter atomic.Int32

func New(t testing.TB) *J {

	// Ideally we could determine ahead of time how many unit tests are being run in the current session;
	// but there doesn't seem to be a way to do that.  Instead, turn on verbosity iff this is the first
	// J object being constructed.

	var testNumber = unitTestCounter.Add(1)

	return &J{
		TB:       t,
		Filename: determineUnittestFilename(CallerLocation(1)),
		verbose:  testNumber == 1,
	}
}

// Deprecated: this constructor will cause the old hash code to be thrown out
//
//goland:noinspection GoUnusedExportedFunction
func Newz(t testing.TB) *J {
	return &J{
		TB:                t,
		Filename:          determineUnittestFilename(CallerLocation(1)),
		InvalidateOldHash: true,
		verbose:           true,
	}
}

type J struct {
	testing.TB
	Filename          string
	verbose           bool
	testResultsDir    Path
	unitTestDir       Path
	moduleDir         Path
	generatedDir      Path
	baseNameCached    string
	InvalidateOldHash bool
	rand              *rand.Rand
	randSeed          int
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

func (j *J) GetUnitTestDir() Path {
	if j.unitTestDir.Empty() {
		var dir = j.GetModuleDir().JoinM("unit_test")
		dir.MkDirsM()
		j.unitTestDir = dir
	}
	return j.unitTestDir
}

func (j *J) getGeneratedDir() Path {
	if j.generatedDir.Empty() {
		var genDir = j.GetUnitTestDir().JoinM("generated")
		genDir.MkDirsM()
		j.generatedDir = genDir
	}
	return j.generatedDir
}

func (j *J) GetTestResultsDir() Path {
	if j.testResultsDir.Empty() {
		var dir = j.getGeneratedDir().JoinM(j.Filename + "/" + j.BaseName())
		// Delete any existing contents of this directory
		// Make sure it contains '/generated/' (pretty sure it does) to avoid crazy deletion
		err := dir.RemakeDir("/generated/")
		CheckOk(err)
		j.testResultsDir = dir
	}
	return j.testResultsDir
}

func (j *J) SetVerbose() {
	j.verbose = true
}

func (j *J) GetModuleDir() Path {
	if j.moduleDir.Empty() {
		var path, err = AscendToDirectoryContainingFile("", "go.mod")
		CheckOk(err)
		j.moduleDir = path
	}
	return j.moduleDir
}

func (j *J) Log(message ...any) {
	if j.Verbose() {
		Pr(message...)
	}
}

func (j *J) GenerateMessage(message ...any) {
	var text = ToString(message...)
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
	var path = j.GetTestResultsDir().JoinM(filename)
	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	path.WriteStringM(content)
}

func (j *J) AssertMessage(message ...any) {
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
	_, e := hasher.Write(b)
	CheckOk(e)
	return int32((hasher.Sum32()&0xffff)%9000 + 1000)
}

// Construct hash of generated directory, and verify it has the expected value.
func (j *J) AssertGenerated() {

	var jsonMap = DirSummary(j.GetTestResultsDir())
	var currentHash = HashOfJSMap(jsonMap)
	var registry = j.registry()

	if !registry.VerifyHash(j.Name(), currentHash, j.InvalidateOldHash) {
		var summary = ToString("\nUnexpected hash value for directory contents:", CR)
		Pr(summary)
		j.showDiffs()
		j.Fail()
	}
	registry.SaveTestResults()
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
			bytes, err := dir.JoinM(filename).ReadBytes()
			CheckOk(err)
			value = HashOfBytes(bytes)
		}
		jsMap.Put(filename, value)
	}

	return jsMap
}

// Display diff of generated directory and its reference version
func (j *J) showDiffs() {

	var refDir = j.registry().referenceDir()
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

var TextFileExtensions = NewSet[string]()

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

func (j *J) Seed(seed int) *J {
	j.randSeed = seed
	j.rand = nil
	return j
}

func (j *J) Rand() *rand.Rand {
	if j.rand == nil {
		if j.randSeed == 0 {
			j.randSeed = 1965
		}
		j.rand = rand.New(rand.NewSource(int64(j.randSeed)))
	}
	return j.rand
}

// Generate a directory structure based upon a JSMap script.  The target argument, if not an absolute directory,
// is assumed to be relative to the test's results directory.
// The jsmap has keys representing files or directories.  If the value is a string, it generates a random text file;
// and if it is a jsmap, it generates a directory recursively.
func (j *J) GenerateSubdirs(target Path, jsmap JSMap) {
	var dir Path
	if target.IsAbs() {
		dir = target
	} else {
		dir = j.GetTestResultsDir().JoinM(target.String())
	}
	j.auxGenDir(dir, jsmap)
}

func (j *J) auxGenDir(dir Path, jsmap JSMap) {
	dir.MkDirsM()
	for key, val := range jsmap.WrappedMap() {
		s, ok := val.(JSMap)
		if ok {
			j.auxGenDir(dir.JoinM(key), s)
		} else {
			targ := dir.JoinM(key)
			text := data.RandomText(j.Rand(), 80, false) + "\n"
			targ.WriteStringM(text)
		}
	}
}
