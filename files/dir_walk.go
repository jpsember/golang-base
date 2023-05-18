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
	logger         Logger
}

func (w *DirWalk) Logger() Logger {
	return w.logger
}

func NewDirWalk(directory Path) *DirWalk {
	var w = new(DirWalk)
	w.logger = NewLogger(w)
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

	var pr = Printer(w)
	pr("Files()")

	if w.absFilesList == nil {
		var lst []Path
		var stack = NewArray[Path]()
		stack.Add(w.startDirectory)

		pr("start dir:", w.startDirectory)
		var firstDir = true
		for !stack.IsEmpty() {

			var dir = stack.Pop()
			if !firstDir && !w.withRecurse {
				continue
			}
			firstDir = false

			files, err := os.ReadDir(dir.String())
			CheckOk(err, "failed to read dir:", dir)

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
					pr("...omitting:", nm)
					continue
				}
				var child = dir.JoinM(nm)

				if child.DirExists() {
					Todo("Have option to include directories in returned list")
					stack.Add(child)
					pr("stacking dir", child)
				} else {
					pr("adding  file", child)
					lst = append(lst, child)
				}
			}
		}
		w.absFilesList = lst
		pr("constructed list:", lst)
	}
	return w.absFilesList
}
