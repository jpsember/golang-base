package main

import (
	. "github.com/jpsember/golang-base/app"
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/files"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var _ = Pr

//
//func main() {
//	app := prepareApp()
//	addCopyDirOper(app)
//	addExamineFilenamesOper(app)
//	app.Start()
//}

type CopyDirOper struct {
	BaseObject
	errLog     ErrLog
	sourcePath Path
	destPath   Path
	errCount   int
}

func (oper *CopyDirOper) UserCommand() string {
	return "copydir"
}

func procPath(desc string, expr string) (Path, string) {
	var err error
	problem := ""
	result := EmptyPath
	for {
		if expr == "" {
			problem = "path is empty"
			break
		}
		absPath, err := filepath.Abs(expr)
		if err != nil {
			break
		}
		result, err = NewPath(absPath)
		if err != nil {
			break
		}
		break
	}
	if err != nil {
		problem = err.Error()
	}
	if problem != "" {
		problem = desc + "; problem: " + problem
	}
	return result, problem
}

func (oper *CopyDirOper) Perform(app *App) {
	oper.SetVerbose(app.Verbose())

	c := app.CmdLineArgs()
	{
		var operSourceDir, operDestDir Path
		problem := ""
		for {
			operSourceDir, problem = procPath("Source directory", c.GetString("source"))
			if problem == "" {
				operDestDir, problem = procPath("Target directory", c.GetString("dest"))
			}
			if problem != "" {
				break
			}
			if !operSourceDir.IsDir() {
				problem = "source is not a directory: " + operSourceDir.String()
				break
			}
			break
		}
		if problem != "" {
			Pr("Problem:", problem)
			os.Exit(1)
		}
		oper.sourcePath = operSourceDir
		oper.destPath = operDestDir
	}

	oper.errLog = NewErrLog(oper.destPath)

	dirStack := NewArray[Path]()
	depthStack := NewArray[int]()
	dirStack.Add(oper.sourcePath)
	depthStack.Add(0)

	sourcePrefixLen := len(oper.sourcePath.String())
	targetPrefix := oper.destPath.String()

	for dirStack.NonEmpty() {
		dir := dirStack.Pop()
		depth := depthStack.Pop()
		//oper.Log("Popped dir:", dir)

		// Make target directory if it doesn't already exist
		targetDir := NewPathM(targetPrefix + dir.String()[sourcePrefixLen:])
		err := targetDir.MkDirs()
		if err != nil {
			oper.errLog.Add(err, "unable to make directory", dir)
			continue
		}
		dirEntries, err := os.ReadDir(dir.String())
		if err != nil {
			oper.errLog.Add(err, "unable to read directory contents", dir)
			continue
		}

		for _, dirEntry := range dirEntries {
			nm := dirEntry.Name()

			sourceFile := dir.JoinM(nm)

			// Check if source is a symlink.  If so, skip it.
			srcFileInfo, err := os.Lstat(sourceFile.String())
			if err != nil {
				oper.errLog.Add(err, "unable to get Lstat for", sourceFile)
				continue
			}
			if srcFileInfo.Mode()&os.ModeSymlink == os.ModeSymlink {
				continue
			}

			sfn := sourceFile.String()
			CheckArg(strings.HasPrefix(sfn, sourceFile.String()))

			sourceFileSuffix := sourceFile.String()[sourcePrefixLen:]
			targetFile := NewPathM(targetPrefix + sourceFileSuffix)

			// If target file already exists, verify it is the same type (dir or file) as source
			if targetFile.Exists() {
				if sourceFile.IsDir() != targetFile.IsDir() {
					oper.errLog.Add(err, "source is not same file/dir type as target:", sourceFile, INDENT,
						"vs", targetFile)
					continue
				}
				// If it is a file, do nothing else
				if !sourceFile.IsDir() {
					continue
				}
			}

			if sourceFile.IsDir() {
				dirStack.Add(sourceFile)
				depthStack.Add(depth + 1)
				continue
			}
			oper.Log(DepthDots(depth, sourceFileSuffix))

			sourceFileStat, err := os.Stat(sourceFile.String())
			if err != nil {
				oper.errLog.Add(err, "getting Stat", sourceFile)
				continue
			}
			if !sourceFileStat.Mode().IsRegular() {
				oper.errLog.Add(err, "source file is not a regular file", sourceFile)
				continue
			}
			err = copyFileContents(sourceFile, targetFile)
			if err != nil {
				oper.errLog.Add(err, "copying file contents", sourceFile, targetFile)
				continue
			}

			modifiedTime := sourceFileStat.ModTime()
			err = os.Chtimes(targetFile.String(), modifiedTime, modifiedTime)
			if err != nil {
				oper.errLog.Add(err, "unable to set modified time", targetFile)
				continue
			}
		}
	}
	oper.errLog.PrintSummary()
}

// Copies file.  If destination exists, its contents will be replaced.
func copyFileContents(srcp, dstp Path) (err error) {
	src := srcp.String()
	dst := dstp.String()

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

func (oper *CopyDirOper) GetHelp(bp *BasePrinter) {
	bp.Pr("Copy a directory  -s <source dir> -d <dest dir>")
}

func addCopyDirOper(app *App) {
	var oper = &CopyDirOper{}
	oper.ProvideName(oper)
	app.RegisterOper(oper)
	Todo("assume if string is empty that none was given")
	app.CmdLineArgs(). //
				Add("source").SetString().Desc("source directory").   //
				Add("dest").SetString().Desc("destination directory") //
}
