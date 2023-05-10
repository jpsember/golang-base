package jt

import (
	. "js/base"
	"js/files"
	. "js/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var pr = Pr

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

type J struct {
	testing.TB
	Filename       string
	verbose        bool
	testResultsDir string
	unitTestDir    string
	moduleDir      string
	baseNameCached string
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
	var content = ToString(message...)

	if j.Verbose() {
		Pr(content)
	}

	Todo("If it is JSON, we want to pretty print it; I guess ToString should handle that")
	var text = ToString(message)
	if j.Verbose() {
		var q strings.Builder
		for _, s := range strings.Split(text, "\n") {
			q.WriteString("\u21e8")
			q.WriteString(s)
			q.WriteString("\u21e6\n")
		}
		Todo("Add CR, INDENT support to Pr, etc")
		//j.Log(CR, "Content:", INDENT, q);
		j.Log("Content:")
		j.Log(q.String())
	}
	j.generateMessageTo("message.txt", text)
}

func (j *J) generateMessageTo(filename string, content string) {
	var path = filepath.Join(j.GetTestResultsDir(), filename)
	files.WriteString(path, content)
}

func (j *J) AssertMessage(message ...any) {
	j.generateMessageTo("message.txt", ToString(message...))
	j.AssertGenerated()
}

// Construct hash of generated directory, and verify it has the expected value.
func (j *J) AssertGenerated() {
	if j.Verbose() {
		j.createInspectionDir()
	}

	var jsonMap = dirSummary(j.GetTestResultsDir())

	// /**
	//   - Construct hash of generated directory, and verify it has the expected value
	//     */
	//     public void assertGeneratedDirectoryHash() {
	//     if (mUnitTest.verbose())
	//     createInspectionDir();
	//     try {
	//     JSMap jsonMap = MyTestUtils.dirSummary(generatedDir());
	//     // Convert hash code to one using exactly four digits
	//     int currentHash = (jsonMap.hashCode() & 0xffff) % 9000 + 1000;
	//     HashCodeRegistry registry = HashCodeRegistry.registryFor(mUnitTest);
	//     if (!registry.verifyHash(mUnitTest.name(), currentHash, mInvalidateOldHash)) {
	//     fail(BasePrinter.toString("\nUnexpected hash value for directory contents:", CR, DASHES, CR, //
	//     jsonMap, CR, DASHES, CR));
	//     }
	//     } catch (Throwable t) {
	//     showDiffs();
	//     throw t;
	//     }
	//     saveTestResults();
	//     }
	Halt("not finished", jsonMap)
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
			//         value = Files.tryHash(f);
			//         if (value == null)
			//           value = DataUtil.checksum(f);
			Todo("calculate hash of file")
			value = 123
		}
		jsMap.Put(filename, value)
	}
	//if Empty(subdirSummary)
	//       Object value = "?";
	//       if (f.isDirectory()) {
	//         JSMap subdirSummary = auxDirSummary(f, ignored, calculateFileHashes);
	//         if (subdirSummary.isEmpty())
	//           continue;
	//         value = subdirSummary;
	//       } else if (calculateFileHashes) {
	//       }

	return jsMap
	// private static JSMap auxDirSummary(File dir, Set<String> ignored, boolean calculateFileHashes) {
	//     List<File> files = files(dir);
	//     JSMap m = map();
	//     for (File f : files) {
	//       String s = f.getName();
	//       if (ignored.contains(s)) {
	//         continue;
	//       }
	//       Object value = "?";
	//       if (f.isDirectory()) {
	//         JSMap subdirSummary = auxDirSummary(f, ignored, calculateFileHashes);
	//         if (subdirSummary.isEmpty())
	//           continue;
	//         value = subdirSummary;
	//       } else if (calculateFileHashes) {
	//         value = Files.tryHash(f);
	//         if (value == null)
	//           value = DataUtil.checksum(f);
	//       }
	//       m.putUnsafe(s, value);
	//     }
	//     return m;
	//   }

}

func (j *J) createInspectionDir() {
	Todo("unimplemented: createInspectionDir")
}
