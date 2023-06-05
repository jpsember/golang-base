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

	//s := "abc?hij"
	//sz := len(s)
	//flaw := 3
	//for i := 0; i < flaw; i++ {
	//	for j := flaw; j < sz; j++ {
	//		t := s[i : j+1]
	//		k := oper.highlightStrangeCharacters(t)
	//		Pr("i:", i, Quoted(t), len(t), "k:", k)
	//		Pr(t[:k] + ">>>" + t[k:])
	//	}
	//}
	//Halt()

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

		dirEntries, err := os.ReadDir(dir.String())
		if err != nil {
			oper.errLog.Add(err, "unable to ReadDir", dir)
			continue
		}

		for _, dirEntry := range dirEntries {
			nm := dirEntry.Name()

			sourceFile := dir.JoinM(nm)
			oper.examineFilename(sourceFile)

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
	bp.Pr("Examine filenames; source <source dir> [clean_log]")
}

var windowsTempPattern = Regexp(`^~\$`)

func (oper *FilenamesOper) examineFilename(p Path) {
	oper.namesCount++
	base := p.Base()

	// See https://en.wikipedia.org/wiki/Tilde
	if windowsTempPattern.MatchString(base) {
		switch oper.config.Microsoft() {
		default:
			Die("unsupported option:", oper.config.Microsoft())
		case Ignore:
			return
		case Warn:
			if oper.config.VerboseProblems() {
				oper.errLog.Add(Warning, "temporary Word file:", Quoted(base), "in", p)
			} else {
				oper.errLog.Add(Warning, "Word:", Quoted(base))
			}
		}
		return
	}

	if oper.pattern.MatchString(base) {
		return
	}
	//var summary string
	summary := oper.highlightStrangeCharacters(base) + " -- " + Quoted(base)
	//if hlPos < 0 || hlPos >= len(base) {
	//	summary = ToString("Unknown! hlpos:", hlPos, "base:", base)
	//} else {
	//	txt := base[0:hlPos] + "![" + base[hlPos:hlPos+1] + "]"
	//	if hlPos+1 < len(base) {
	//		txt += base[hlPos+1:]
	//	}
	//	summary = Quoted(txt)
	//}
	if oper.config.VerboseProblems() {
		oper.errLog.Add(Warning, "strange characters:", summary, "in", p)
	} else {
		oper.errLog.Add(Warning, "Chars:", summary)
	}
}

func (oper *FilenamesOper) highlightStrangeCharacters(str string) string {
	//str = "...unknown?.."
	pr := PrIf(false)

	x := len(str)
	pr("find flaw in:", Quoted(str), "Length:", x)
	low := 0
	high := x
	inf := 10
	for low != high {
		i := (low + high) / 2
		pref := str[0:i]
		pr("low", low, "high", high, "i", i, "substring:", Quoted(pref))
		if i == 0 || oper.pattern.MatchString(pref) {
			low = i + 1
			pr("match, low now", low)
		} else {
			high = i
			pr("no match, high now", i)
		}
		inf--
		CheckState(inf != 0)
	}
	pref := str[0:low]
	pr("substring:", Quoted(pref))

	j := low - 1
	return Quoted(str[0:j] + ">>>" + str[:j])
}

func addExamineFilenamesOper(app *App) {
	var oper = &FilenamesOper{}
	oper.ProvideName("names")
	app.RegisterOper(AssertJsonOper(oper))
	//app.SetTestArgs("names  source osxutils/sample clean_log  ")
}
