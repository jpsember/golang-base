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
		var dir, err = os.Getwd()
		CheckOk(err)
		startDir = dir
	}
	var path = startDir

	var prevPath = path
	for {
		var cand = filepath.Join(path, seekFile)
		_, e := os.Stat(cand)
		if e == nil {
			return path, nil
		}
		path = filepath.Dir(path)
		if path == prevPath {
			return "", errors.New(ToString("Cannot find", seekFile, "in tree containing", startDir))
		}
		prevPath = path
	}

}
