package files

import (
	. "github.com/jpsember/golang-base/base"
	"os"
	"regexp"
)

type DirWalk struct {
	startDirectory Path
	withRecurse    bool
	patternsSet    *Set[string]
	patternsToOmit *Array[*regexp.Regexp]
	absFilesList   []Path
}

func NewDirWalk(directory Path) *DirWalk {
	var w = new(DirWalk)
	w.startDirectory = directory.CheckNonEmptyWithSkip(1)
	w.patternsToOmit = NewArray[*regexp.Regexp]()
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
		w.patternsToOmit.Add(r)
	}
	return w
}

func (w *DirWalk) Files() []Path {
	if w.absFilesList == nil {

		var stack = NewArray[Path]()
		stack.Add(w.startDirectory)
		var firstDir = true
		for !stack.IsEmpty() {
			var dir = stack.Pop()

			if !firstDir && !w.withRecurse {
				continue
			}
			firstDir = false

			files, err := os.ReadDir(w.startDirectory.String())
			CheckOk(err)

			var lst []Path

			for _, file := range files {
				var nm = file.Name()
				for _, pat := range w.patternsToOmit.Array() {
					if pat.MatchString(nm) {
						continue
					}
				}

				var child = dir.JoinM(nm)

				if child.DirExists() {
					Todo("Have option to include directories in returned list")
					stack.Add(child)
				} else {
					lst = append(lst, child)
				}
			}
		}
		w.absFilesList = lst
	}
	return w.absFilesList
}
