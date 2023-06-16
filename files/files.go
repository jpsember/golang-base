package files

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/json"
	"io"
	"os"
	"strings"
)

// Deprecated: use Path type
// Delete a directory.  For safety, the path must contain a particular substring.
func DeleteDir(path string, substring string) error {
	CheckArg(len(substring) >= 5, "substring is too short:", Quoted(substring))
	CheckArg(strings.Contains(path, substring), "path", Quoted(path), "doesn't contain substring", Quoted(substring))
	return os.RemoveAll(path)
}

// Deprecated: use Path type
// Write string to file
// Panics if error occurs
func WriteString(path string, content string) {
	var err = os.WriteFile(path, []byte(content), 0644)
	CheckOk(err, "Failed to write string to path:", path)
}

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
		pr("path now:", path, "isEmpty:", path.Empty(), "emptry str:", EmptyPath.String())
		if path.Empty() {
			return path, Error("Cannot find", seekFile, "in tree containing", startDir)
		}
		CheckState(path.String() != "/")
	}
}

func AscendToDirectoryContainingFileM(startDir Path, seekFile string) Path {
	var pth, err = AscendToDirectoryContainingFile(startDir, seekFile)
	CheckOkWithSkip(1, err)
	return pth
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
	content, err := path.ReadString()
	CheckOkWithSkip(1, err)
	return content
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
	content, err := path.ReadStringIfExists(defaultContent)
	CheckOkWithSkip(1, err)
	return content
}

// Deprecated: use Path type
func ReadBytes(path string) (content []byte, err error) {
	return os.ReadFile(path)
}

func (path Path) ReadBytes() (content []byte, err error) {
	return os.ReadFile(string(path))
}

func (path Path) ReadBytesM() (content []byte) {
	bytes, err := os.ReadFile(string(path))
	CheckOkWithSkip(2, err)
	return bytes
}

func (path Path) Chmod(mode os.FileMode) error {
	return os.Chmod(path.String(), mode)
}

func (path Path) ChmodM(mode os.FileMode) {
	CheckOk(path.Chmod(mode))
}

func CurrentDirectory() Path {
	path, err := os.Getwd()
	CheckOk(err)
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
	var result, err = JSMapFromFile(file)
	CheckOkWithSkip(1, err)
	return result
}

func JSMapFromFileIfExists(file Path) (JSMap, error) {
	var content, _ = file.ReadStringIfExists("{}")
	return JSMapFromString(content)
}

func JSMapFromFileIfExistsM(file Path) JSMap {
	var result, err = JSMapFromFileIfExists(file)
	CheckOkWithSkip(1, err)
	return result
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
	CheckOkWithSkip(1, err, "can't find project directory")
	return path
}

func FindProjectDir() (Path, error) {
	return AscendToDirectoryContainingFile("", "project_config")
}
