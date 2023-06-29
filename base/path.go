package base

import (
	"os"
	"path/filepath"
	"strings"
)

type Path string

var EmptyPath = Path("")

var homeDir = EmptyPath

func HomeDirM() Path {
	if homeDir.Empty() {
		p, err := os.UserHomeDir()
		CheckOkWithSkip(1, err)
		homeDir = NewPathM(p)
	}
	return homeDir
}

func TempFile(prefix string) (Path, error) {
	if prefix == "" {
		prefix = "jefftemp_*"
	}
	path := EmptyPath
	f, err := os.CreateTemp("", prefix)
	if err != nil {
		path, err = NewPath(f.Name())
	}
	return path, err
}

func TempFileM(prefix string) Path {
	result, err := TempFile(prefix)
	CheckOkWithSkip(2, err)
	return result
}

// Construct a Path from a string; return error if there is a problem
func NewPath(s string) (Path, error) {
	if s == "" {
		return "", Error("Path is empty")
	}
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

// Join path to a relative path (Path)
func (path Path) JoinPath(other Path) (Path, error) {
	return path.Join(string(other))
}

// Join path to a relative path (Path)
func (path Path) JoinPathM(other Path) Path {
	return path.JoinM(string(other))
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
	if s == "." || s == "/" {
		Todo("need to distinguish between root path and empty path")
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
	return path.WriteBytes([]byte(content))
}

// Write string to file; panic if error
func (path Path) WriteStringM(content string) {
	CheckOkWithSkip(1, path.WriteString(content))
}

// Write bytes to file
func (path Path) WriteBytes(content []byte) error {
	path.CheckNonEmptyWithSkip(1)
	return os.WriteFile(string(path), content, 0644)
}

// Write string to file; panic if error
func (path Path) WriteBytesM(content []byte) {
	CheckOkWithSkip(1, path.WriteBytes(content))
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

func (path Path) RemakeDirM(substring string) {
	CheckOkWithSkip(1, path.RemakeDir(substring))
}

func (path Path) DeleteDirectory(substring string) error {
	CheckArg(!path.Empty())
	if len(substring) < 5 || !strings.Contains(string(path), substring) {
		BadArg("DeleteDirectory, path doesn't contain suitably long substring:", path, Quoted(substring))
	}
	return os.RemoveAll(string(path))
}

func (path Path) DeleteDirectoryM(substring string) {
	CheckOkWithSkip(1, path.DeleteDirectory(substring))
}

func (path Path) DeleteFile() error {
	CheckArg(!path.Empty())
	return os.Remove(string(path))
}

func (path Path) DeleteFileM() {
	CheckOkWithSkip(1, path.DeleteFile())
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

func (path Path) IsAbs() bool {
	path.CheckNonEmptyWithSkip(1)
	return filepath.IsAbs(path.String())
}

func (path Path) GetAbs() (Path, error) {
	path.CheckNonEmptyWithSkip(1)
	pth, err := filepath.Abs(path.String())
	result := EmptyPath
	if err == nil {
		result, err = NewPath(pth)
	}
	return result, err
}

func (path Path) GetAbsM() Path {
	result, err := path.GetAbs()
	CheckOkWithSkip(1, err)
	return result
}

func (path Path) GetAbsFrom(defaultParentDir Path) (Path, error) {
	var err error
	result := path
	if !path.IsAbs() {
		result, err = defaultParentDir.JoinPath(path)
	}
	return result, err
}

func (path Path) GetAbsFromM(defaultParentDir Path) Path {
	result, err := path.GetAbsFrom(defaultParentDir)
	CheckOkWithSkip(1, err)
	return result
}

func (path Path) Info(message ...any) JSMap {
	m := NewJSMap()
	m.Put("", ToString(JoinElementToList("File info;", message)...))
	var absPath Path
	if path.NonEmpty() {
		content := "MISSING"
		if path.Exists() {
			if path.IsDir() {
				content = "DIRECTORY"
			} else {
				content = "FILE"
			}
		}
		if path.IsAbs() {
			m.Put("2 name", path.Base())
			m.Put("3 parent", path.Parent().String())
			absPath = path
		} else {
			relName := path.String()
			curr := CurrentDirectory()
			m.Put("2 rel", relName)
			m.Put("3 cdir", curr.String())
			absPath = curr.JoinPathM(path)
		}
		m.Put("4 abs", absPath.String())
		m.Put("1 status", content)
	}
	return m
}
