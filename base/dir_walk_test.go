package base_test

import (
	. "github.com/jpsember/golang-base/base"
	"github.com/jpsember/golang-base/jt"
	"strings"
	"testing"
)

var _ = Pr
var _ = JSFalse

func sampleDir(j jt.JTest) Path {
	return j.GetUnitTestDir().JoinM("sample_dir")
}

func TestDirWalk(t *testing.T) {
	j := jt.New(t)
	var w = sampleWalker(j)
	w.WithRecurse()

	// Skip specific full names
	//
	w.OmitNames("unit_test")

	// Skip names containing substrings (prefixes in this case)
	//
	w.OmitNamesWithSubstrings("^_SKIP_", `^\.`)
	assertWalk(j, w)
}

func sampleWalker(j jt.JTest) *DirWalk {
	var w = NewDirWalk(sampleDir(j))
	w.SetVerbose(j.Verbose())
	return w
}

func assertWalk(j jt.JTest, w *DirWalk) {
	var m = NewJSMap()
	for _, x := range w.FilesRelative() {
		m.PutNumbered(x.String())
	}
	j.AssertMessage(m)
}

func TestSampleDir(t *testing.T) {
	j := jt.New(t)
	var w = sampleWalker(j)
	w.WithRecurse()
	assertWalk(j, w)
}

func TestSampleDirWithDirNames(t *testing.T) {
	j := jt.New(t)
	var w = sampleWalker(j)
	w.WithRecurse()
	w.WithDirNames()
	assertWalk(j, w)
}

func TestIncludePrefixes(t *testing.T) {
	j := jt.New(t)
	var w = sampleWalker(j)
	w.WithRecurse()
	// Omit any files (or directories) starting with _SKIP_ or a dot
	w.OmitNamesWithSubstrings("^_SKIP_", `^\.`)
	// Include files with particular extensions
	w.ForFiles().IncludeExtensions("go", "json")
	// Omit subdirectories with substring "harl"
	w.ForDirs().OmitNamesWithSubstrings("harl")
	assertWalk(j, w)
}

func TestAscendToDirectoryContainingFile(t *testing.T) {
	j := jt.Newz(t)
	_, err := AscendToDirectoryContainingFile(EmptyPath, "hello")
	j.Log("Ascend result:", err)
	j.AssertTrue(strings.Contains(err.Error(), "Cannot find hello"))
}
