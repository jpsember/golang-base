package files

import (
	. "github.com/jpsember/golang-base/base"
	"os"
	"path/filepath"
	"strings"
)

type Path string

var EmptyPath = Path("")

// Construct a Path from a string; return error if there is a problem
func NewPath(s string) (Path, error) {
	var cleaned = filepath.Clean(s)
	if cleaned != s {
		return "", Error("Path isn't clean:", Quoted(s), "; should be:", Quoted(cleaned))
	}
	if strings.HasPrefix(s, "..") {
		return "", Error("Illegal path:", Quoted(s))
	}
	if s == "." {
		return "", Error("Attempt to construct empty path:", Quoted(s))
	}
	return Path(s), nil
}

// Construct a Path from a string; panic if there is a problem
func NewPathM(s string) Path {
	p, err := NewPath(s)
	CheckOkWithSkip(1, err)
	return p
}

// Join path to a relative path (string)
func (p Path) Join(s string) (Path, error) {
	j := filepath.Join(string(p), s)
	return NewPath(j)
}

// Join path to a relative path (string); panic if error
func (p Path) JoinM(s string) Path {
	j, e := p.Join(s)
	CheckOkWithSkip(1, e)
	return j
}

// Get string representation of path
func (p Path) String() string {
	return string(p)
}

// Panic if path is empty
func (p Path) CheckNonEmpty() {
	p.CheckNonEmptyWithSkip(1)
}

func (p Path) CheckNonEmptyWithSkip(skip int) {
	if p.Empty() {
		BadArgWithSkip(1+skip, "Path is empty")
	}
}

// Get parent of (nonempty) path; returns empty path if it has no parent
func (p Path) Parent() Path {
	p.CheckNonEmptyWithSkip(1)
	var s = filepath.Dir(string(p))
	if s == "." {
		return EmptyPath
	}
	return Path(s)
}

// Determine if path refers to a file (or directory)
func (p Path) Exists() bool {
	p.CheckNonEmptyWithSkip(1)
	_, err := os.Stat(string(p))
	return err == nil
}

// Determine if path refers to directory
func (p Path) DirExists() bool {
	fileInfo, err := os.Stat(string(p))
	return err == nil && fileInfo.IsDir()
}

// Determine if path is empty
func (p Path) Empty() bool {
	return string(p) == ""
}

// Write string to file
func (p Path) WriteString(content string) error {
	p.CheckNonEmptyWithSkip(1)
	return os.WriteFile(string(p), []byte(content), 0644)
}

// Write string to file; panic if error
func (p Path) WriteStringM(content string) {
	CheckOkWithSkip(1, p.WriteString(content))
}

// Get the filename denoted by (nonempty) path
func (p Path) Base() string {
	p.CheckNonEmptyWithSkip(1)
	return filepath.Base(string(p))
}
