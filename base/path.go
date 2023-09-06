package base

import (
	"os"
	"path/filepath"
	"strings"
)

type Path string

var EmptyPath = Path("")

func homeDirFunc() (any, error) {
	var pth Path
	p, err := os.UserHomeDir()
	if err == nil {
		pth, err = NewPath(p)
	}
	return pth, err
}

var homeDir = newCachedResult(homeDirFunc)
var projectDir = newCachedResult(func() (any, error) {
	return AscendToDirectoryContainingFile(EmptyPath, ".git")
})

var FileNotFoundError = Error("file not found")

// ------------------------------------------------------------------------------------
// An experimental structure for delaying finding a path, and recording err when it does
// ------------------------------------------------------------------------------------

type evalFunc func() (any, error)

type cachedResultStruct struct {
	result any
	err    error
	fn     evalFunc
}
type cachedResult = *cachedResultStruct

func newCachedResult(eval evalFunc) cachedResult {
	t := &cachedResultStruct{
		fn: eval,
	}
	return t
}

func (cd cachedResult) Result() (any, error) {
	if cd.fn != nil {
		cd.result, cd.err = cd.fn()
		cd.fn = nil
	}
	return cd.result, cd.err
}

func (cd cachedResult) ResultM() any {
	return CheckOkWith(cd.Result())
}

// ------------------------------------------------------------------------------------

// ------------------------------------------------------------------------------------
// An experimental structure for delaying finding a path, and recording err when it does
// ------------------------------------------------------------------------------------

type cachedDirEvalFunc func() (Path, error)

type cachedDirStruct struct {
	path Path
	err  error
	f    cachedDirEvalFunc
}
type cachedDir = *cachedDirStruct

func newCachedDir(eval cachedDirEvalFunc) cachedDir {
	t := &cachedDirStruct{
		f: eval,
	}
	return t
}

func (cd cachedDir) PathM() Path {
	cd.Path()
	if cd.err != nil {
		BadState("<1Trouble determining directory:", cd.err)
	}
	return cd.path
}

func (cd cachedDir) Path() (Path, error) {
	if cd.f != nil {
		pth, err := cd.f()
		cd.path = pth
		cd.err = err
	}
	return cd.path, cd.err
}

// ------------------------------------------------------------------------------------

// Deprecated.
func FindFileUpward(name string, startDir Path) (Path, error) {
	CheckArg(startDir.IsDir(), "not a directory")

	currentDir := startDir
	for currentDir.NonEmpty() {
		candidate := currentDir.JoinM(name)
		Pr("candidate:", candidate)
		if candidate.Exists() {
			return candidate, nil
		}
		currentDir = currentDir.Parent()
	}
	return EmptyPath, FileNotFoundError
}

func ProjectDirM() Path {
	return projectDir.ResultM().(Path)
}

func HomeDirM() Path {
	return homeDir.ResultM().(Path)
}

func HomeDir() (Path, error) {
	result, err := homeDir.Result()
	return result.(Path), err
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
	return CheckOkWith(TempFile(prefix))
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
	return CheckOkWith(NewPathOrEmpty(s))
}

// Construct a Path from a string; panic if there is a problem
func NewPathM(s string) Path {
	return CheckOkWith(NewPath(s))
}

// Join path to a relative path (string)
func (path Path) Join(s string) (Path, error) {
	j := filepath.Join(string(path), s)
	return NewPath(j)
}

// Join path to a relative path (string); panic if error
func (path Path) JoinM(s string) Path {
	return CheckOkWith(path.Join(s))
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
	if path.Empty() {
		BadArg("Path is empty")
	}
	return path
}

// Get parent of (nonempty) path; returns empty path if it has no parent
func (path Path) Parent() Path {
	path.CheckNonEmpty()
	input := string(path)
	var s = filepath.Dir(input)
	if s == input {
		return EmptyPath
	}
	return Path(s)
}

// Determine if path refers to a file (or directory)
func (path Path) Exists() bool {
	path.CheckNonEmpty()
	_, err := os.Stat(string(path))
	return err == nil
}

func (path Path) IsDir() bool {
	fileInfo, err := os.Stat(string(path))
	return err == nil && fileInfo.IsDir()
}

func (path Path) IsRoot() bool {
	path.CheckNonEmpty()
	return path.String() == "/"
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
	CheckOk(path.WriteString(content))
}

// Write bytes to file
func (path Path) WriteBytes(content []byte) error {
	path.CheckNonEmpty()
	return os.WriteFile(string(path), content, 0644)
}

// Write string to file; panic if error
func (path Path) WriteBytesM(content []byte) {
	CheckOk(path.WriteBytes(content))
}

// Get the filename denoted by (nonempty) path
func (path Path) Base() string {
	path.CheckNonEmpty()
	return filepath.Base(string(path))
}

func (path Path) MkDirs() error {
	return os.MkdirAll(string(path), os.ModePerm)
}

func (path Path) MkDirsM() {
	CheckOk(path.MkDirs())
}

func (path Path) RemakeDir(substring string) error {
	err := path.DeleteDirectory(substring)
	if err == nil {
		err = path.MkDirs()
	}
	return err
}

func (path Path) RemakeDirM(substring string) {
	CheckOk(path.RemakeDir(substring))
}

func (path Path) DeleteDirectory(substring string) error {
	CheckArg(!path.Empty())
	if len(substring) < 5 || !strings.Contains(string(path), substring) {
		BadArg("DeleteDirectory, path doesn't contain suitably long substring:", path, Quoted(substring))
	}
	return os.RemoveAll(string(path))
}

func (path Path) DeleteDirectoryM(substring string) {
	CheckOk(path.DeleteDirectory(substring))
}

func (path Path) DeleteFile() error {
	CheckArg(!path.Empty())
	if !path.Exists() {
		return nil
	}
	return os.Remove(string(path))
}

func (path Path) DeleteFileM() {
	CheckOk(path.DeleteFile())
}

func (path Path) MoveTo(target Path) error {
	CheckArg(!path.Empty())
	CheckArg(!target.Empty())
	if target.Exists() && !target.IsDir() {
		return Error("Can't move to existing file:", target)
	}
	return os.Rename(string(path), string(target))
}

func ExtensionFrom(path string) string {
	return strings.TrimPrefix(filepath.Ext(path), ".")
}

func (path Path) Extension() string {
	return ExtensionFrom(path.String())
}

func (path Path) TrimExtension() Path {
	p := path.CheckNonEmpty().String()
	ext := filepath.Ext(p)
	if ext != "" {
		i := len(p)
		return NewPathM(p[:i-len(ext)])
	}
	return path
}

func (path Path) SetExtension(ext string) Path {
	CheckNonEmpty(ext)
	return NewPathM(path.TrimExtension().String() + "." + ext)
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
	path.CheckNonEmpty()
	return filepath.IsAbs(path.String())
}

func (path Path) GetAbs() (Path, error) {
	path.CheckNonEmpty()
	pth, err := filepath.Abs(path.String())
	result := EmptyPath
	if err == nil {
		result, err = NewPath(pth)
	}
	return result, err
}

func (path Path) GetAbsM() Path {
	return CheckOkWith(path.GetAbs())
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
	return CheckOkWith(path.GetAbsFrom(defaultParentDir))
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
