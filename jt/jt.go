package jt

import (
	"hash/fnv"
	"os"
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
		dir.RemakeDir("/generated/")
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
		j.showDiffs()
		j.Fail()
	}
	registry.SaveTestResults()
}

func dirSummary(dir Path) *JSMap {
	return auxDirSummary(dir, true)
}

func auxDirSummary(dir Path, calculateFileHashes bool) *JSMap {
	var jsMap = NewJSMap()

	var entries, err = os.ReadDir(dir.String())
	CheckOk(err)

	for _, ent := range entries {

		var filename = ent.Name()
		var value any

		value = "?"
		if ent.IsDir() {
			var subdirSummary = auxDirSummary(dir.JoinM(filename), calculateFileHashes)
			Todo("have JSMap.size or empty method")
			value = subdirSummary
		} else if calculateFileHashes {
			var bytes []byte
			bytes, err = dir.JoinM(filename).ReadBytes()
			CheckOk(err)
			value = HashOfBytes(bytes)
		}
		jsMap.Put(filename, value)
	}

	return jsMap
}

// Display diff of generated directory and its reference version
func (j *J) showDiffs() {

	//    private void showDiffs() {

	//     File refDir = referenceDir();
	//     if (!refDir.exists())
	//       return;

	//     Set<File> relFiles = hashSet();

	//     DirWalk dirWalk = new DirWalk(refDir).withRecurse(true).omitNames(".DS_Store");
	//     relFiles.addAll(dirWalk.filesRelative());
	//     dirWalk = new DirWalk(generatedDir()).withRecurse(true).omitNames(".DS_Store");
	//     relFiles.addAll(dirWalk.filesRelative());

	//     for (File fileReceived : relFiles) {
	//       File fileRecAbs = dirWalk.abs(fileReceived);
	//       File fileRefAbs = new File(refDir, fileReceived.getPath());

	//       if (fileRefAbs.exists() && fileRecAbs.exists()
	//           && Arrays.equals(Files.toByteArray(fileRecAbs, null), Files.toByteArray(fileRefAbs, null)))
	//         continue;

	//       if (!mUnitTest.verbose())
	//         continue;

	//       pr(CR,
	//           "------------------------------------------------------------------------------------------------");
	//       pr(fileReceived);

	//       if (!fileRefAbs.exists()) {
	//         pr("...unexpected file");
	//         continue;
	//       }
	//       if (!fileRecAbs.exists()) {
	//         pr("...file has disappeared");
	//         continue;
	//       }

	//       // If it looks like a text file, call the 'diff' utility to display differences.
	//       // Otherwise, only do this (using binary mode) if in verbose mode
	//       //
	//       String ext = Files.getExtension(fileReceived);

	//       boolean isTextFile = sTextFileExtensions.contains(ext);

	//       SystemCall sc = new SystemCall().arg("diff");
	//       if (isTextFile)
	//         sc.arg("--text"); // "Treat all files as text."

	//       if (true) {
	//         sc.arg("-C", "2");
	//       } else {
	//         sc.arg("--side-by-side"); //  "Output in two columns."
	//       }
	//       sc.arg(fileRefAbs, fileRecAbs);
	//       pr();
	//       pr(sc.systemOut());
	//       // It is returning 2 if it encounters binary files (e.g. xxx.zip), which is problematic
	//       if (sc.exitCode() > 2)
	//         badState("System call failed:", INDENT, sc);
	//     }
	//   }

}
