package main

import (
	. "github.com/jpsember/golang-base/app"
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/files"
	. "github.com/jpsember/golang-base/gen/sample"
	"os"
	"regexp"
	"strings"
)

type FilenamesOper struct {
	BaseObject
	errLog     ErrLog
	errPath    Path
	sourcePath Path
	destPath   Path
	errCount   int
	config     NamesConfig
	namesCount int
	pattern    *regexp.Regexp
	deleteFlag bool
}

func (oper *FilenamesOper) GetArguments() DataClass {
	return DefaultNamesConfig
}

func (oper *FilenamesOper) ArgsFileMustExist() bool {
	return false
}

func (oper *FilenamesOper) AcceptArguments(a DataClass) {
	oper.config = a.(NamesConfig)
}

func (oper *FilenamesOper) UserCommand() string {
	return "names"
}

func (oper *FilenamesOper) Perform(app *App) {
	oper.SetVerbose(app.Verbose())
	oper.pattern = Regexp(oper.config.Pattern())
	{
		var operSourceDir Path
		problem := ""
		for {
			operSourceDir, problem = procPath("Source directory", oper.config.Source())
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

	oper.errLog = NewErrLog(oper.config.Log())
	oper.errLog.Clean = oper.config.CleanLog()

	dirStack := NewArray[Path]()
	depthStack := NewArray[int]()
	dirStack.Add(oper.sourcePath)
	depthStack.Add(0)

	sourcePrefixLen := len(oper.sourcePath.String())

	for dirStack.NonEmpty() {
		maxIssues := oper.config.MaxProblems()
		if maxIssues > 0 && oper.errLog.IssueCount() >= int(maxIssues) {
			oper.errLog.Add(Warning, "Stopping since max issue count has been reached")
			break
		}
		dir := dirStack.Pop()
		depth := depthStack.Pop()

		oper.examineFilename(dir)
		if oper.processDeleteFlag(dir) {
			continue
		}
		dirEntries, err := os.ReadDir(dir.String())
		if err != nil {
			oper.errLog.Add(err, "unable to ReadDir", dir)
			continue
		}

		for _, dirEntry := range dirEntries {
			nm := dirEntry.Name()

			sourceFile := dir.JoinM(nm)
			oper.examineFilename(sourceFile)
			if oper.processDeleteFlag(sourceFile) {
				continue
			}

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

func (oper *FilenamesOper) processDeleteFlag(path Path) bool {
	result := oper.deleteFlag
	if result {
		if path.IsDir() {
			path.DeleteDirectoryM("~$")
		} else {
			path.DeleteFileM()
		}
	}
	return result
}

func (oper *FilenamesOper) GetHelp(bp *BasePrinter) {
	bp.Pr("Examine filenames; source <source dir> [clean_log]")
}

var windowsTempPattern = Regexp(`^~\$`)

func (oper *FilenamesOper) examineFilename(p Path) {
	oper.deleteFlag = false
	oper.namesCount++
	base := p.Base()

	// See https://en.wikipedia.org/wiki/Tilde
	if windowsTempPattern.MatchString(base) {
		switch oper.config.Microsoft() {
		default:
			Die("unsupported option:", oper.config.Microsoft())
		case Ignore:
			break
		case Warn:
			if oper.config.VerboseProblems() {
				oper.errLog.Add(Warning, "temporary Word file:", Quoted(base), "in", p)
			} else {
				oper.errLog.Add(Warning, "Word:", Quoted(base))
			}
		case Delete:
			oper.errLog.Add(Warning, "Deleting Word:", Quoted(base))
			oper.deleteFlag = true
		}
		return
	}

	if !oper.pattern.MatchString(base) {
		summary := oper.highlightStrangeCharacters(base)
		if oper.config.VerboseProblems() {
			oper.errLog.Add(Warning, "strange characters:", summary, "in", p)
		} else {
			oper.errLog.Add(Warning, "Chars:", summary)
		}
	}
}

func (oper *FilenamesOper) highlightStrangeCharacters(str string) string {
	// I was doing a binary search, but I found out that due to utf-8, some chars (runes)
	// are different lengths; so just build up the substring from the left until we find the problem
	sb := strings.Builder{}
	sbPost := strings.Builder{}

	problemFound := false
	prob := ""
	for _, ch := range str {
		if !problemFound {
			sb.WriteRune(ch)
			prob = sb.String()
			if !oper.pattern.MatchString(prob) {
				problemFound = true
			}
		} else {
			sbPost.WriteRune(ch)
		}
	}
	return Quoted(sb.String() + "<<<" + sbPost.String())
}

func addExamineFilenamesOper(app *App) {
	var oper = &FilenamesOper{}
	oper.ProvideName("names")
	app.RegisterOper(AssertJsonOper(oper))
	//app.SetTestArgs("names  source osxutils/sample clean_log  ")
}
