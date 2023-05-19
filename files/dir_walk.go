package files

import (
	. "github.com/jpsember/golang-base/base"
	"os"
	"regexp"
	"strings"
)

type DirWalk struct {
	startDirectory       Path
	withRecurse          bool
	patternsSet          *Set[string]
	patternsToOmit       *Array[*regexp.Regexp]
	patternsToIncludeSet *Set[string]
	patternsToInclude    *Array[*regexp.Regexp]
	absFilesList         []Path
	relFilesList         []Path
	logger               Logger
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
	w.patternsToInclude = NewArray[*regexp.Regexp]()
	w.patternsToIncludeSet = NewSet[string]()

	w.OmitNames(defaultOmitPrefixes...)
	return w
}

var defaultOmitPrefixes = []string{
	"_SKIP_", "_OLD_",
}

func (w *DirWalk) WithRecurse(flag bool) *DirWalk {
	w.assertMutable()
	w.withRecurse = flag
	return w
}

func (w *DirWalk) WithExtensions(ext ...string) *DirWalk {
	w.assertMutable()
	for _, exp := range ext {
		var exp2 = `\.` + exp + `$`
		if w.patternsToIncludeSet.Add(exp2) {
			r, err := regexp.Compile(exp2)
			CheckOk(err, "failed to compile omit expr:", exp2)
			w.patternsToInclude.Add(r)
		}
	}
	return w
}

// Add a list of regular expressions describing filenames that should be omitted.
// Wraps each expression in '^' ... '$' so that the expression must match the entire string.
// See: https://pkg.go.dev/regexp/syntax
func (w *DirWalk) OmitNames(nameExprs ...string) *DirWalk {
	w.assertMutable()
	var strlist = NewArray[string]()
	for _, expr := range nameExprs {
		if !strings.HasPrefix(expr, "^") {
			expr = "^" + expr + "$"
		}
		strlist.Add(expr)
	}
	return w.OmitNamesWithSubstrings(strlist.wrappedArray...)
}

// Add a list of regular expressions describing filenames that should be omitted.
// Any name that contains text matching one of the regular expressions will be omitted
// See: https://pkg.go.dev/regexp/syntax
func (w *DirWalk) OmitNamesWithSubstrings(nameExprs ...string) *DirWalk {
	w.assertMutable()
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

	var pr = Printer(w)
	pr("Files()")

	if w.absFilesList == nil {

		pr("start dir:", w.startDirectory)

		var lst []Path
		var stack = NewArray[Path]()
		stack.Add(w.startDirectory)

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

				var child = dir.JoinM(nm)
				var childIsDir = child.DirExists()

				// If no explicit patterns to *include* were given, then we apply the omit filter
				// to everything.  Otherwise, we apply the omit filter only to directories.

				Todo("Maybe the thing to do is to (formally) have separate filters for directories vs files")

				if childIsDir || w.patternsToInclude.IsEmpty() {

					for _, pat := range w.patternsToOmit.Array() {
						if pat.MatchString(nm) {
							omit = true
							pr("...omitting:", nm)
							break
						}
					}

					if omit {
						continue
					}
				}

				if !childIsDir && w.patternsToInclude.NonEmpty() {
					omit = true
					for _, pat := range w.patternsToInclude.Array() {
						if pat.MatchString(nm) {
							omit = false
							pr("...including file:", nm)
							break
						}
					}
				}
				if omit {
					continue
				}

				if childIsDir {
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

func (w *DirWalk) FilesRelative() []Path {
	if w.relFilesList == nil {
		var prefixLength = len(w.startDirectory.String()) + 1 // include the separator
		var x = NewArray[Path]()

		for _, path := range w.Files() {
			var str = path.String()
			if len(str) <= prefixLength {
				continue
			}
			x.Add(NewPathM(str[prefixLength:]))
		}
		w.relFilesList = x.Slice()
	}
	return w.relFilesList
}

func (w *DirWalk) assertMutable() {
	CheckState(w.absFilesList == nil, "results already generated")
}
