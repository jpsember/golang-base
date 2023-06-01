package files

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/json"
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

	var path = startDir
	for {
		var cand, _ = path.Join(seekFile)
		if cand.Exists() {
			return path, nil
		}
		if path.Empty() {
			return path, Error("Cannot find", seekFile, "in tree containing", startDir)
		}
		path = path.Parent()
		if path.Empty() {
			return path, Error("Cannot find", seekFile, "in tree containing", startDir)
		}
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

//// Deprecated: use Path type
//func ReadStringIfExists(file string, defaultContent string) (content string, err error) {
//	Todo("have a special type for File, to avoid confusion with strings, and to support paths etc")
//	if Exists(file) {
//		var bytes []byte
//		bytes, err = ReadBytes(file)
//		if err == nil {
//			content = string(bytes)
//		}
//	} else {
//		content = defaultContent
//	}
//	return content, err
//}

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

func CurrentDirectory() Path {
	path, err := os.Getwd()
	CheckOk(err)
	return NewPathM(path)
}

func JSMapFromFileIfExists(file Path) (*JSMapStruct, error) {
	var content, _ = file.ReadStringIfExists("{}")
	return JSMapFromString(content)
}

func JSMapFromFileIfExistsM(file Path) *JSMapStruct {
	var result, err = JSMapFromFileIfExists(file)
	CheckOkWithSkip(1, err)
	return result
}
