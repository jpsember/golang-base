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
	absFilesList   []Path
	relFilesList   []Path
	logger         Logger
	patternFlags   int
	regexpSet      map[string]*regexp.Regexp
	includeDirs    bool
}

// BaseObject implementation
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

// Include directory names in the returned file list.  Normally,
// these are omitted.
func (w *DirWalk) WithDirNames() *DirWalk {
	w.includeDirs = true
	return w
}

// Have subsequent patterns affect only files
func (w *DirWalk) ForFiles() *DirWalk {
	w.assertMutable()
	w.patternFlags = patflag_file
	return w
}

// Have subsequent patterns affect only directories
func (w *DirWalk) ForDirs() *DirWalk {
	w.assertMutable()
	w.patternFlags = patflag_dir
	return w
}

// Recurse into subdirectories
func (w *DirWalk) WithRecurse() *DirWalk {
	w.assertMutable()
	w.withRecurse = true
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

// Include filenames that have particular extensions
// (This includes directories, unless ForFiles() is in effect; which
// might be a bit strange)
func (w *DirWalk) IncludeExtensions(ext ...string) *DirWalk {
	for _, exp := range ext {
		w.includePattern(`\.` + exp + `$`)
	}
	return w
}

func (w *DirWalk) includePattern(exp string) {
	var flags = w.patternFlags
	r := w.addPattern(exp)
	if (flags & patflag_file) != 0 {
		w.filePatterns.Include.Add(r)
	}
	if (flags & patflag_dir) != 0 {
		w.dirPatterns.Include.Add(r)
	}
}

func (w *DirWalk) omitPattern(exp string) {
	var flags = w.patternFlags
	r := w.addPattern(exp)
	if (flags & patflag_file) != 0 {
		w.filePatterns.Omit.Add(r)
	}
	if (flags & patflag_dir) != 0 {
		w.dirPatterns.Omit.Add(r)
	}
}

// Add a list of regular expressions describing filenames that should be omitted.
// Wraps each expression in '^' ... '$' so that the expression must match the entire string.
// See: https://pkg.go.dev/regexp/syntax
func (w *DirWalk) OmitNames(nameExprs ...string) *DirWalk {
	for _, exp := range nameExprs {
		w.omitPattern(`^` + exp + `$`)
	}
	return w
}

// Add a list of regular expressions describing filenames that should be omitted if they
// contain a substring.
// Any name that contains text matching one of the regular expressions will be omitted
// See: https://pkg.go.dev/regexp/syntax
func (w *DirWalk) OmitNamesWithSubstrings(nameExprs ...string) *DirWalk {
	for _, exp := range nameExprs {
		w.omitPattern(exp)
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

			if w.includeDirs {
				lst = append(lst, dir)
			}

			files, err := os.ReadDir(dir.String())
			CheckOk(err, "failed to read dir:", dir)

			for _, file := range files {
				var nm = file.Name()
				var omit = false

				var child = dir.JoinM(nm)
				var childIsDir = child.IsDir()

				// Determine which pattern set to apply
				var pats *patternCollection
				if childIsDir {
					pats = w.dirPatterns
				} else {
					pats = w.filePatterns
				}

				for _, pat := range pats.Omit.wrappedArray {
					if pat.MatchString(nm) {
						omit = true
						break
					}
				}

				if omit {
					continue
				}

				if pats.Include.NonEmpty() {
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
					stack.Add(child)
					pr("stacking dir", child)
				} else {
					pr("adding  file", child)
					lst = append(lst, child)
				}
			}
		}
		w.absFilesList = lst
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
