package files_test

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/files"
	. "github.com/jpsember/golang-base/json"
	"github.com/jpsember/golang-base/jt"
	"testing"
)

var _ = Pr
var _ = JSFalse

func sampleDir(j *jt.J) Path {
	return j.GetUnitTestDir().JoinM("sample_dir")
}

func TestDirWalk(t *testing.T) {
	j := jt.New(t) // Use Newz to regenerate hash
	j.SetVerbose()

	var w = NewDirWalk(sampleDir(j))
	w.Logger().SetVerbose(true)
	w.WithRecurse()

	// Skip specific full names
	//
	w.OmitNames("unit_test")

	// Skip names containing substrings (prefixes in this case)
	//
	w.OmitNamesWithSubstrings("^_SKIP_", `^\.`)

	var m = NewJSMap()
	for _, x := range w.FilesRelative() {
		m.PutNumbered(x.String())
	}
	j.AssertMessage(m)

}

func TestIncludePrefixes(t *testing.T) {
	j := jt.New(t)
	var w = NewDirWalk(sampleDir(j))
	w.Logger().SetVerbose(j.Verbose())
	w.WithRecurse()
	// Omit any files (or directories) starting with _SKIP_ or a dot
	w.OmitNamesWithSubstrings("^_SKIP_", `^\.`)
	// Include files with particular extensions
	w.ForFiles().IncludeExtensions("go", "json")
	// Omit subdirectories with substring "harl"
	w.ForDirs().OmitNamesWithSubstrings("harl")

	var m = NewJSMap()
	for _, x := range w.FilesRelative() {
		m.PutNumbered(x.String())
	}
	j.AssertMessage(m)

}
