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

	var w = NewDirWalk(dir).WithRecurse(true).OmitNames(`\.DS_Store`, `\.git`)

	Pr(w.Files())

}
