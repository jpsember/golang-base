package files

import (
	"errors"
	. "github.com/jpsember/golang-base/base"
	"os"
	"path/filepath"
	"strings"
)

// var pr = Pr

// Delete a directory.  For safety, the path must contain a particular substring.
func DeleteDir(path string, substring string) error {
	CheckArg(len(substring) >= 5, "substring is too short:", Quoted(substring))
	CheckArg(strings.Contains(path, substring), "path", Quoted(path), "doesn't contain substring", Quoted(substring))
	return os.RemoveAll(path)
}

// Write string to file
// Panics if error occurs
func WriteString(path string, content string) {
	var err = os.WriteFile(path, []byte(content), 0644)
	CheckOk(err, "Failed to write string to path:", path)
}

func AscendToDirectoryContainingFile(startDir string, seekFile string) (string, error) {
	CheckArg(NonEmpty(seekFile))
	if Empty(startDir) {
		startDir = CurrentDirectory()
	}
	ValidatePath(startDir)
	var path = startDir
	for {
		var cand = filepath.Join(path, seekFile)
		if Exists(cand) {
			return path, nil
		}
		path = Parent(path)
		if path == "" {
			return "", errors.New(ToString("Cannot find", seekFile, "in tree containing", startDir))
		}
	}
}

func ReadStringIfExists(file string, defaultContent string) (content string, err error) {
	Todo("have a special type for File, to avoid confusion with strings, and to support paths etc")
	if Exists(file) {
		var bytes []byte
		bytes, err = ReadBytes(file)
		if err == nil {
			content = string(bytes)
		}
	} else {
		content = defaultContent
	}
	return content, err
}

func ReadBytes(path string) (content []byte, err error) {
	return os.ReadFile(path)
}

func MkDirs(file string) (string, error) {
	//Pr("attempt to MkDirs:", file)
	err := os.MkdirAll(file, os.ModePerm)
	CheckOk(err, "failed MkDirs:", file)
	return file, err
}

func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func DirExists(path string) bool {
	fileInfo, err := os.Stat(path)
	return err == nil && fileInfo.IsDir()
}

func PathJoin(parent, child string) string {
	if strings.HasSuffix(parent, "/") || len(parent) == 0 || strings.HasPrefix(child, "/") || len(child) == 0 {
		Die("illegal args for PathJoin:", parent, child)
	}
	return parent + "/" + child
}

func Parent(path string) string {
	CheckArg(!strings.HasSuffix(path, "/"))
	i := strings.LastIndex(path, "/")
	if i < 0 {
		return ""
	}
	return path[0:i]
}

func ValidatePath(path string) string {
	if path == "" || strings.HasSuffix(path, "/") {
		BadArgWithSkip(1, "invalid path:", path)
	}
	return path
}

func GetName(path string) string {
	ValidatePath(path)
	i := strings.LastIndex(path, "/")
	if i < 0 {
		return path
	}
	return path[i+1:]
}

func CurrentDirectory() string {
	path, err := os.Getwd()
	CheckOk(err)
	return path
}
