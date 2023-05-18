package files

import (
	. "github.com/jpsember/golang-base/base"
	"os"
	"path/filepath"
	"strings"
)

type Path string

var EmptyPath = Path("")

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

func NewPathM(s string) Path {
	p, err := NewPath(s)
	CheckOkWithSkip(1, err)
	return p
}

// Deprecated: we can't use Must for anything else
func Must(path Path, e error) Path {
	if e != nil {
		BadArgWithSkip(2, e)
	}
	return path
}

func (p Path) Join(s string) (Path, error) {
	j := filepath.Join(string(p), s)
	return NewPath(j)
}

func (p Path) JoinM(s string) Path {
	j, e := p.Join(s)
	CheckOkWithSkip(1, e)
	return j
}

func (p Path) String() string {
	return string(p)
}

func (p Path) CheckNonEmpty() {
	p.CheckNonEmptyWithSkip(1)
}

func (p Path) CheckNonEmptyWithSkip(skip int) {
	if p.Empty() {
		BadArgWithSkip(1+skip, "Path is empty")
	}
}

func (p Path) Parent() Path {
	p.CheckNonEmptyWithSkip(1)
	var s = filepath.Dir(string(p))
	if s == "." {
		return EmptyPath
	}
	return Path(s)
}

func (p Path) Exists() bool {
	p.CheckNonEmptyWithSkip(1)
	_, err := os.Stat(string(p))
	return err == nil
}

func (p Path) Empty() bool {
	return string(p) == ""
}

// Write string to file
func (p Path) WriteString(content string) error {
	p.CheckNonEmptyWithSkip(1)
	return os.WriteFile(string(p), []byte(content), 0644)
}

func (p Path) WriteStringM(content string) {
	must(p.WriteString(content))
}

func (p Path) Base() string {
	p.CheckNonEmptyWithSkip(1)
	return filepath.Base(string(p))
}

func must(e error) {
	CheckOkWithSkip(1, e)
}
