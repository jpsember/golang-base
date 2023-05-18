package files

import (
	. "github.com/jpsember/golang-base/base"
	"regexp"
)

// var pr = Pr

type DirWalk struct {
	withRecurse    bool
	patternsToOmit map[*regexp.Regexp]bool
}

func NewDirWalk() *DirWalk {
	var w = new(DirWalk)
	w.patternsToOmit = make(map[*regexp.Regexp]bool)
	return w
}

func (w *DirWalk) WithRecurse(flag bool) *DirWalk {
	w.withRecurse = flag
	return w
}

func (w *DirWalk) OmitNames(nameExprs ...string) *DirWalk {
	for _, expr := range nameExprs {
		r, err := regexp.Compile(expr)
		CheckOk(err, "failed to compile omit expr:", expr)
		w.patternsToOmit[r] = true
	}
	return w
}
