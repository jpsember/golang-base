package files

import (
	. "github.com/jpsember/golang-base/base"
	"os"
	"regexp"
)

type DirWalk struct {
	startDirectory Path
	withRecurse    bool
	filePatterns   *patternCollection
	dirPatterns    *patternCollection
	absFilesList []Path
	relFilesList []Path
	logger       Logger
	patternFlags int
	regexpSet    map[string]*regexp.Regexp
}

func (w *DirWalk) Logger() Logger {
	return w.logger
}

const patflag_file = 1 << 0
const patflag_dir = 1 << 1

func NewDirWalk(directory Path) *DirWalk {
	var w = new(DirWalk)
	w.patternFlags = patflag_file | patflag_dir
	w.regexpSet = make(map[string]*regexp.Regexp)
	w.logger = NewLogger(w)
	w.startDirectory = directory.CheckNonEmptyWithSkip(1)
	w.filePatterns = newPatternCollection()
	w.dirPatterns = newPatternCollection()
	w.OmitNames(defaultOmitPrefixes...)
	return w
}

var defaultOmitPrefixes = []string{
	"_SKIP_", "_OLD_",
}

// Have subsequent patterns affect only files
func (w *DirWalk) ForFiles() *DirWalk {
	w.patternFlags = patflag_file
	return w
}

// Have subsequent patterns affect only directories
func (w *DirWalk) ForDirs() *DirWalk {
	w.patternFlags = patflag_dir
	return w
}

func (w *DirWalk) WithRecurse(flag bool) *DirWalk {
	w.assertMutable()
	w.withRecurse = flag
	return w
}

func (w *DirWalk) addPatterns(pat ...string) {
	for _, p := range pat {
		w.addPattern(p)
	}
}

func (w *DirWalk) addPattern(pat string) *regexp.Regexp {
	w.assertMutable()
	r, hasKey := w.regexpSet[pat]
	if !hasKey {
		r2, err := regexp.Compile(pat)
		CheckOk(err, "failed to compile reg exp:", pat)
		w.regexpSet[pat] = r2
		r = r2
	}
	return r
}

func (w *DirWalk) IncludeExtensions(ext ...string) *DirWalk {
	Todo("only files should have extensions?")
	var flags = w.patternFlags
	Pr("include extensions:", ext)
	for _, exp := range ext {
		var exp2 = `\.` + exp + `$`
		Pr("adding include exp:", exp2)
		r := w.addPattern(exp2)

		if (flags & patflag_file) != 0 {
			w.filePatterns.Include.Add(r)
		}
		if (flags & patflag_dir) != 0 {
			w.dirPatterns.Include.Add(r)
		}
	}
	return w
}

// Add a list of regular expressions describing filenames that should be omitted.
// Wraps each expression in '^' ... '$' so that the expression must match the entire string.
// See: https://pkg.go.dev/regexp/syntax
func (w *DirWalk) OmitNames(nameExprs ...string) *DirWalk {
	var flags = w.patternFlags
	for _, exp := range nameExprs {
		var exp2 = `^` + exp + `$`
		r := w.addPattern(exp2)

		if (flags & patflag_file) != 0 {
			w.filePatterns.Omit.Add(r)
		}
		if (flags & patflag_dir) != 0 {
			w.dirPatterns.Omit.Add(r)
		}
	}
	return w
}

// Add a list of regular expressions describing filenames that should be omitted.
// Any name that contains text matching one of the regular expressions will be omitted
// See: https://pkg.go.dev/regexp/syntax
func (w *DirWalk) OmitNamesWithSubstrings(nameExprs ...string) *DirWalk {
	var flags = w.patternFlags
	for _, exp := range nameExprs {
		r := w.addPattern(exp)
		if (flags & patflag_file) != 0 {
			w.filePatterns.Omit.Add(r)
		}
		if (flags & patflag_dir) != 0 {
			w.dirPatterns.Omit.Add(r)
		}
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
				var childIsDir = child.IsDir()
				pr("candidate:", nm, "dir:", childIsDir)

				// Determine which pattern set to apply
				var pats *patternCollection
				if childIsDir {
					pats = w.dirPatterns
					pr("looking at dirPatterns")
				} else {
					pr("looking at filePatterns")
					pats = w.filePatterns
				}

				for _, pat := range pats.Omit.wrappedArray {
					if pat.MatchString(nm) {
						omit = true
						pr("...omitting:", nm)
						break
					}
				}

				if omit {
					continue
				}

			if pats.Include.NonEmpty() {
					pr("include is nonempty, checking...")
					omit = true
					for _, pat := range pats.Include.wrappedArray {
						if pat.MatchString(nm) {
							omit = false
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

//type patternSet struct {
//	PatternSet  *Set[string]
//	PatternList *Array[*regexp.Regexp]
//}

//func newPatternSet() *patternSet {
//	var p = new(patternSet)
//	p.PatternSet = NewSet[string]()
//	p.PatternList = NewArray[*regexp.Regexp]()
//	return p
//}

type patternCollection struct {
	Include *Array[*regexp.Regexp]
	Omit    *Array[*regexp.Regexp]
}

func newPatternCollection() *patternCollection {
	var p = new(patternCollection)
	p.Include = NewArray[*regexp.Regexp]()
	p.Omit = NewArray[*regexp.Regexp]()
	return p
}

//type patterns struct {
//	patternsSet          *Set[string]
//	patternsToOmit       *Array[*regexp.Regexp]
//	patternsToIncludeSet *Set[string]
//	patternsToInclude    *Array[*regexp.Regexp]
//
//}
