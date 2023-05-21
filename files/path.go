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

// Construct a Path from a string, or the empty path if string is empty
func NewPathOrEmpty(s string) (Path, error) {
	if s == "" {
		return EmptyPath, nil
	}
	return NewPath(s)
}

// Construct a Path from a string, or the empty path if string is empty
func NewPathOrEmptyM(s string) Path {
	p, err := NewPathOrEmpty(s)
	CheckOkWithSkip(1, err)
	return p
}

// Construct a Path from a string; panic if there is a problem
func NewPathM(s string) Path {
	p, err := NewPath(s)
	CheckOkWithSkip(1, err)
	return p
}

// Join path to a relative path (string)
func (path Path) Join(s string) (Path, error) {
	j := filepath.Join(string(path), s)
	return NewPath(j)
}

// Join path to a relative path (string); panic if error
func (path Path) JoinM(s string) Path {
	j, e := path.Join(s)
	CheckOkWithSkip(1, e)
	return j
}

// Get string representation of path
func (path Path) String() string {
	if path.Empty() {
		return "<EMPTY>"
	}
	return string(path)
}

// Panic if path is empty
func (path Path) CheckNonEmpty() Path {
	return path.CheckNonEmptyWithSkip(1)
}

func (path Path) CheckNonEmptyWithSkip(skip int) Path {
	if path.Empty() {
		BadArgWithSkip(1+skip, "Path is empty")
	}
	return path
}

// Get parent of (nonempty) path; returns empty path if it has no parent
func (path Path) Parent() Path {
	path.CheckNonEmptyWithSkip(1)
	var s = filepath.Dir(string(path))
	if s == "." {
		return EmptyPath
	}
	return Path(s)
}

// Determine if path refers to a file (or directory)
func (path Path) Exists() bool {
	path.CheckNonEmptyWithSkip(1)
	_, err := os.Stat(string(path))
	return err == nil
}

func (path Path) IsDir() bool {
	fileInfo, err := os.Stat(string(path))
	return err == nil && fileInfo.IsDir()
}

// Determine if path is empty
func (path Path) Empty() bool {
	return string(path) == ""
}

// Write string to file
func (path Path) WriteString(content string) error {
	path.CheckNonEmptyWithSkip(1)
	return os.WriteFile(string(path), []byte(content), 0644)
}

// Write string to file; panic if error
func (path Path) WriteStringM(content string) {
	CheckOkWithSkip(1, path.WriteString(content))
}

// Get the filename denoted by (nonempty) path
func (path Path) Base() string {
	path.CheckNonEmptyWithSkip(1)
	return filepath.Base(string(path))
}

func (path Path) MkDirs() error {
	return os.MkdirAll(string(path), os.ModePerm)
}

func (path Path) MkDirsM() {
	CheckOkWithSkip(1, path.MkDirs())
}

func (path Path) RemakeDir(substring string) error {
	err := path.DeleteDirectory(substring)
	if err == nil {
		err = path.MkDirs()
	}
	return err
}

func (path Path) DeleteDirectory(substring string) error {
	CheckArg(!path.Empty())
	if len(substring) < 5 || !strings.Contains(string(path), substring) {
		BadArg("DeleteDirectory, path doesn't contain suitably long substring:", path, Quoted(substring))
	}
	return os.RemoveAll(string(path))
}

func (path Path) MoveTo(target Path) error {
	CheckArg(!path.Empty())
	CheckArg(!target.Empty())
	if target.Exists() && !target.IsDir() {
		return Error("Can't move to existing file:", target)
	}
	return os.Rename(string(path), string(target))
}

func (path Path) Extension() string {
	return strings.TrimPrefix(filepath.Ext(path.String()), ".")
}

func (path Path) NonEmpty() bool {
	return !path.Empty()
}

func (path Path) EnsureExists(message ...any) {
	if !path.Exists() {
		BadArg(JoinLists([]any{"File doesn't exist:", path, ";"}, message))
	}
}
