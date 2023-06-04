package main

import (
	. "github.com/jpsember/golang-base/app"
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/files"
	"os"
	"strings"
)

type FilenamesOper struct {
	BaseObject
	errLog     ErrLog
	errPath    Path
	sourcePath Path
	destPath   Path
	errCount   int
}

func (oper *FilenamesOper) UserCommand() string {
	return "names"
}

func (oper *FilenamesOper) Perform(app *App) {
	oper.SetVerbose(app.Verbose())

	c := app.CmdLineArgs()
	{
		var operSourceDir Path
		problem := ""
		for {
			operSourceDir, problem = procPath("Source directory", c.GetString("source"))
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
	}

	oper.errLog = NewErrLog(oper.sourcePath)

	dirStack := NewArray[Path]()
	depthStack := NewArray[int]()
	dirStack.Add(oper.sourcePath)
	depthStack.Add(0)

	sourcePrefixLen := len(oper.sourcePath.String())

	for dirStack.NonEmpty() {
		dir := dirStack.Pop()
		depth := depthStack.Pop()
		//oper.Log("Popped dir:", dir)

		dirEntries, err := os.ReadDir(dir.String())
		if err != nil {
			oper.errLog.Add(err, "unable to ReadDir", dir)
			continue
		}

		for _, dirEntry := range dirEntries {
			nm := dirEntry.Name()

			sourceFile := dir.JoinM(nm)

			// Check if source is a symlink.  If so, skip it.
			srcFileInfo, err := os.Lstat(sourceFile.String())
			if err != nil {
				oper.errLog.Add(err, "unable to Lstat", sourceFile)
				continue
			}
			if srcFileInfo.Mode()&os.ModeSymlink == os.ModeSymlink {
				oper.errLog.Add(Warning, "Found symlink:", sourceFile)
				continue
			}

			sfn := sourceFile.String()
			CheckArg(strings.HasPrefix(sfn, sourceFile.String()))
			sourceFileSuffix := sourceFile.String()[sourcePrefixLen:]

			if sourceFile.IsDir() {
				dirStack.Add(sourceFile)
				depthStack.Add(depth + 1)
				continue
			}

			oper.Log(DepthDots(depth, sourceFileSuffix))

			sourceFileStat, err := os.Stat(sourceFile.String())
			if err != nil {
				oper.errLog.Add(err, "unable to Stat", sourceFile)
				continue
			}
			if !sourceFileStat.Mode().IsRegular() {
				oper.errLog.Add(err, "file is not a regular file", sourceFile)
				continue
			}
		}
	}
	oper.errLog.PrintSummary()
}

func (oper *FilenamesOper) GetHelp(bp *BasePrinter) {
	bp.Pr("Examine filenames -s <source dir>")
}

func addExamineFilenamesOper(app *App) {
	var oper = &FilenamesOper{}
	oper.ProvideName("names")
	app.RegisterOper(oper)
}
