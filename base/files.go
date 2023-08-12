package base

import (
	"io"
	"os"
)

func AscendToDirectoryContainingFile(startDir Path, seekFile string) (Path, error) {
	CheckArg(NonEmpty(seekFile))

	if startDir.Empty() {
		startDir = CurrentDirectory()
	}

	pr := PrIf(false)

	pr("AscendToDirectoryContainingFile, startDir:", startDir, "seekFile:", seekFile)
	var path = startDir
	for {
		pr("path:", path)
		var cand, _ = path.Join(seekFile)
		pr("candidate:", cand)
		if cand.Exists() {
			return path, nil
		}
		if path.Empty() {
			return path, Error("Cannot find", seekFile, "in tree containing", startDir)
		}
		pr("path:", path, "parent:", path.Parent())
		path = path.Parent()
		pr("path now:", path, "isEmpty:", path.Empty(), "empty str:", EmptyPath.String())
		if path.Empty() {
			return path, Error("Cannot find", seekFile, "in tree containing", startDir)
		}
	}
}

func AscendToDirectoryContainingFileM(startDir Path, seekFile string) Path {
	return CheckOkWith(AscendToDirectoryContainingFile(startDir, seekFile))
}

func (path Path) ReadString() (content string, err error) {
	var bytes []byte
	bytes, err = path.ReadBytes()
	if err == nil {
		content = string(bytes)
	}
	return content, err
}

func (path Path) ReadStringM() string {
	return CheckOkWith(path.ReadString())
}

func (path Path) ReadStringIfExists(defaultContent string) (content string, err error) {
	if path.Exists() {
		var bytes []byte
		bytes, err = path.ReadBytes()
		if err == nil {
			content = string(bytes)
		}
	} else {
		content = defaultContent
	}
	return content, err
}

func (path Path) ReadStringIfExistsM(defaultContent string) string {
	return CheckOkWith(path.ReadStringIfExists(defaultContent))
}

// Deprecated: use Path type
func ReadBytes(path string) (content []byte, err error) {
	return os.ReadFile(path)
}

func (path Path) ReadBytes() (content []byte, err error) {
	return os.ReadFile(string(path))
}

func (path Path) ReadBytesM() (content []byte) {
	return CheckOkWith(os.ReadFile(string(path)))
}

func (path Path) Chmod(mode os.FileMode) error {
	return os.Chmod(path.String(), mode)
}

func (path Path) ChmodM(mode os.FileMode) {
	CheckOk(path.Chmod(mode))
}

func CurrentDirectory() Path {
	path := CheckOkWith(os.Getwd())
	return NewPathM(path)
}

func JSMapFromFile(file Path) (JSMap, error) {
	var result JSMap
	content, err := file.ReadString()
	if err == nil {
		result, err = JSMapFromString(content)
	}
	return result, err
}

func JSMapFromFileM(file Path) JSMap {
	return CheckOkWith(JSMapFromFile(file))
}

func JSMapFromFileIfExists(file Path) (JSMap, error) {
	var content, _ = file.ReadStringIfExists("{}")
	return JSMapFromString(content)
}

func JSMapFromFileIfExistsM(file Path) JSMap {
	return CheckOkWith(JSMapFromFileIfExists(file))
}

// Copies file.  If destination exists, its contents will be replaced.
func CopyFile(sourcePath Path, destPath Path) (err error) {
	src := sourcePath.String()
	dst := destPath.String()

	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return
	}
	err = out.Sync()
	return
}

func FindProjectDirM() Path {
	var path, err = FindProjectDir()
	return CheckOkWith(path, err, "<1 Can't find project directory")
}

func FindProjectDir() (Path, error) {
	if !cachedProjectDirFlag {
		cachedProjectDir, cachedProjectDirErr = AscendToDirectoryContainingFile("", "project_config")
		cachedProjectDirFlag = true
	}
	return cachedProjectDir, cachedProjectDirErr
}

func FindRepoDir() (Path, error) {
	if !cachedRepoDirFlag {
		cachedRepoDir, cachedRepoDirErr = AscendToDirectoryContainingFile("", ".git")
		cachedRepoDirFlag = true
	}
	return cachedRepoDir, cachedRepoDirErr
}

var cachedProjectDirFlag bool
var cachedProjectDir Path
var cachedProjectDirErr error
var cachedRepoDirFlag bool
var cachedRepoDir Path
var cachedRepoDirErr error

type Closeable interface {
	Close() error
}

func ClosePeacefully[T Closeable](c T) T {
	{
		err := c.Close()
		if err != nil {
			Pr(CallerLocation(1), "*** Problem closing;", err)
		}
		// See https://stackoverflow.com/questions/70585852
		var result T
		c = result
	}
	return c
}
