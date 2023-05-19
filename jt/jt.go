package jt

import (
	"hash/fnv"
	"os/exec"
	"reflect"
	"strings"
	"testing"

	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/files"
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
//
//goland:noinspection GoUnusedExportedFunction
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
	testResultsDir    Path
	unitTestDir       Path
	moduleDir         Path
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

func (j *J) GetUnitTestDir() Path {
	if j.unitTestDir.Empty() {
		var dir = j.GetModuleDir().JoinM("unit_test")
		dir.MkDirsM()
		j.unitTestDir = dir
	}
	return j.unitTestDir
}

func (j *J) GetTestResultsDir() Path {
	if j.testResultsDir.Empty() {
		var genDir = j.GetUnitTestDir().JoinM("generated")
		genDir.MkDirsM()
		var dir = genDir.JoinM(j.Filename + "/" + j.BaseName())
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

func HashOfJSMap(jsonMap *JSMap) int32 {
	return HashOfString(PrintJSEntity(jsonMap, false))
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

	var jsonMap = dirSummary(j.GetTestResultsDir())
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

func dirSummary(dir Path) *JSMap {
	var jsMap = NewJSMap()

	var w = NewDirWalk(dir).WithRecurse(true)
	for _, ent := range w.Files() {

		var filename = ent.Base()
		var value any

		value = "?"
		if ent.IsDir() {
			var subdirSummary = dirSummary(dir.JoinM(filename))
			Todo("have JSMap.size or empty method")
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

	Todo("Always skip .DS_Store?")
	var dirWalk = NewDirWalk(refDir).WithRecurse(true).OmitNames(`\.DS_Store`)
	relFiles.AddAll(dirWalk.FilesRelative())

	dirWalk = NewDirWalk(genDir).WithRecurse(true).OmitNames(`\.DS_Store`)
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

		output, err := makeSysCall(args.Slice())
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
