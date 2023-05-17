package files

import (
	. "github.com/jpsember/golang-base/base"
	"path/filepath"
	"strings"
)

type Path string

func NewPath(s string) (Path, error) {
	var cleaned = filepath.Clean(s)
	if cleaned != s {
		return "", Error("Path isn't clean:", Quoted(s), "; should be:", Quoted(cleaned))
	}
	if strings.HasPrefix(s, "..") {
		return "", Error("Illegal path:", Quoted(s))
	}
	return Path(s), nil
}

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

func (p Path) String() string {
	return string(p)
}
