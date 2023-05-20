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

func TestDirWalk(t *testing.T) {
	j := jt.New(t) // Use Newz to regenerate hash
	j.SetVerbose()

	var dir = j.GetModuleDir()
	Pr("module dir:", dir)

	var w = NewDirWalk(dir)
	w.Logger().SetVerbose(true)
	w.WithRecurse(true)

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

	j.SetVerbose()
	var dir = j.GetModuleDir()
	var w = NewDirWalk(dir)
	w.Logger().SetVerbose(true)
	w.WithRecurse(true)
	// Omit any files (or directories) starting with _SKIP_ or a dot
	w.OmitNamesWithSubstrings("^_SKIP_", `^\.`)
	// Include files with particular extensions
	w.ForFiles().IncludeExtensions("go", "json")

	var m = NewJSMap()
	for _, x := range w.FilesRelative() {
		m.PutNumbered(x.String())
	}
	j.AssertMessage(m)

}
