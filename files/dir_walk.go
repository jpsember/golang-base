package files

import (
	. "github.com/jpsember/golang-base/base"
	"regexp"
)

// var pr = Pr

type DirWalk struct {
	withRecurse    bool
	patternsSet    *Set[string]
	patternsToOmit []*regexp.Regexp
}

func NewDirWalk() *DirWalk {
	var w = new(DirWalk)
	w.patternsToOmit = []*regexp.Regexp{}
	w.patternsSet = NewSet[string]()
	return w
}

func (w *DirWalk) WithRecurse(flag bool) *DirWalk {
	w.withRecurse = flag
	return w
}

func (w *DirWalk) OmitNames(nameExprs ...string) *DirWalk {
	for _, expr := range nameExprs {
		if !w.patternsSet.Add(expr) {
			continue
		}
		r, err := regexp.Compile(expr)
		CheckOk(err, "failed to compile omit expr:", expr)
		Todo("Add an ArrayList class")
		w.patternsToOmit = append(w.patternsToOmit, r)
	}
	return w
}
