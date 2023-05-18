package files

import (
	. "github.com/jpsember/golang-base/base"
	"os"
	"regexp"
	"strings"
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

// Add a list of regular expressions describing filenames that should be omitted.
// Wraps each expression in '^' ... '$' so that the expression must match the entire string.
// See: https://pkg.go.dev/regexp/syntax
func (w *DirWalk) OmitNames(nameExprs ...string) *DirWalk {
	for _, expr := range nameExprs {
		if !w.patternsSet.Add(expr) {
			continue
		}
		if strings.HasPrefix(expr, "^") {
			BadArgWithSkip(1, "Unexpected regex expression:", Quoted(expr))
		}
		var expr2 = "^" + expr + "$"
		r, err := regexp.Compile(expr2)
		CheckOk(err, "failed to compile omit expr:", expr)
		w.patternsToOmit.Add(r)
	}
	return w
}

func (w *DirWalk) Files() []Path {
	var inf = 300

	Pr("walk")
	if w.absFilesList == nil {
		var lst []Path
		var stack = NewArray[Path]()
		stack.Add(w.startDirectory)
		Pr("start dir:", w.startDirectory)
		var firstDir = true
		for !stack.IsEmpty() {
			inf--
			CheckState(inf != 0)

			var dir = stack.Pop()
			if !firstDir && !w.withRecurse {
				continue
			}
			firstDir = false

			files, err := os.ReadDir(dir.String())
			CheckOk(err, "failed to read dir:", dir)

			Todo("patterns should have start, stop limits")
			for _, file := range files {
				var nm = file.Name()
				var omit = false
				for _, pat := range w.patternsToOmit.Array() {
					if pat.MatchString(nm) {
						omit = true
						break
					}
				}
				if omit {
					Pr("...omitting:", nm)
					continue
				}
				var child = dir.JoinM(nm)

				if child.DirExists() {
					Todo("Have option to include directories in returned list")
					stack.Add(child)
					Pr("stacking dir", child)
				} else {
					Pr("adding  file", child)
					lst = append(lst, child)
				}
			}
		}
		w.absFilesList = lst
		Pr("constructed list:", lst)
	}
	return w.absFilesList
}
