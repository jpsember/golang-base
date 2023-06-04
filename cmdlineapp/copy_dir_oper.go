package main

import (
	"errors"
	. "github.com/jpsember/golang-base/app"
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/files"
	"io"
	"os"
	"strings"
)

var _ = Pr

func main() {
	cmdLineExample()
}

type CopyDirOper struct {
	BaseObject
	errPath    Path
	sourcePath Path
	destPath   Path
}

func (oper *CopyDirOper) UserCommand() string {
	return "copydir"
}

const dots = "............................................................................................................................................................................."

func (oper *CopyDirOper) Perform(app *App) {
	oper.SetVerbose(app.Verbose())

	c := app.CmdLineArgs()
	{
		var operSourceDir, operDestDir Path
		problem := ""
		for {
			if c.GetString("source") == "" || c.GetString("dest") == "" {
				problem = "Source and dest must both be nonempty"
				break
			}
			operSourceDir = NewPathM(c.GetString("source"))
			operDestDir = NewPathM(c.GetString("dest"))

			if !operSourceDir.IsDir() {
				problem = "source is not a directory: " + operSourceDir.String()
				break
			}
			if false && operDestDir.Exists() {
				problem = "dest path already exists: " + operDestDir.String()
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

	dirStack := NewArray[Path]()
	depthStack := NewArray[int]()
	dirStack.Add(oper.sourcePath)
	depthStack.Add(0)

	sourcePrefixLen := len(oper.sourcePath.String())
	targetPrefix := oper.destPath.String()

	for dirStack.NonEmpty() {
		dir := dirStack.Pop()
		depth := depthStack.Pop()
		oper.Log("Popped dir:", dir)

		// Make target directory if it doesn't already exist
		targetDir := NewPathM(targetPrefix + dir.String()[sourcePrefixLen:])
		err := targetDir.MkDirs()
		if err != nil {
			oper.outputError(err, dir)
			continue
		}
		dirEntries, err := os.ReadDir(dir.String())
		if err != nil {
			oper.outputError(err, dir)
		}

		for _, dirEntry := range dirEntries {
			nm := dirEntry.Name()

			sourceFile := dir.JoinM(nm)

			sfn := sourceFile.String()
			CheckArg(strings.HasPrefix(sfn, sourceFile.String()))

			sourceFileSuffix := sourceFile.String()[sourcePrefixLen:]
			targetFile := NewPathM(targetPrefix + sourceFileSuffix)

			// If target file already exists, verify it is the same type (dir or file) as source
			if targetFile.Exists() {
				if sourceFile.IsDir() != targetFile.IsDir() {
					oper.outputError(errors.New(ToString("source is not same file/dir type as target:", sourceFile.String(), INDENT,
						"vs", targetFile.String())), dir)
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
			prefLen := 2 * depth
			CheckState(prefLen < len(dots))
			oper.Log(dots[0:prefLen] + " " + sourceFileSuffix)

			sourceFileStat, err := os.Stat(sourceFile.String())
			if err != nil {
				oper.outputError(err, sourceFile)
				continue
			}
			if !sourceFileStat.Mode().IsRegular() {
				oper.outputError(errors.New("source file "+sourceFile.String()+" is not a regular file"), sourceFile)
				continue
			}
			err = copyFileContents(sourceFile, targetFile)
			if err != nil {
				oper.outputError(err, sourceFile)
				continue
			}
		}
	}
}

// copyFileContents copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file.
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

func (oper *CopyDirOper) outputError(err error, dir Path) {
	errMsg := ToString("*** error copying subdirectory:", dir)
	Pr(errMsg)

	if oper.errPath.Empty() {
		oper.errPath = NewPathM(oper.destPath.String() + "_errors")
	}
	f, err := os.OpenFile(oper.errPath.String(),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	CheckOk(err, "Failed opening error file:", oper.errPath)
	defer f.Close()
	_, err = f.WriteString(errMsg + "\n")
	CheckOk(err, "Failed appending to error file:", oper.errPath)
}

func (oper *CopyDirOper) GetHelp(bp *BasePrinter) {
	bp.Pr("Copy a directory  -s <source dir> -d <dest dir>")
}

func cmdLineExample() {
	Pr(VERT_SP, DASHES, "copydir", CR, DASHES)
	var oper = &CopyDirOper{}
	oper.ProvideName(oper)
	var app = NewApp()
	app.SetName("copydir")
	app.Version = "2.1.3"
	app.RegisterOper(oper)
	app.CmdLineArgs(). //
				Add("source").SetString().Desc("source directory").   //
				Add("dest").SetString().Desc("destination directory") //
	//app.SetTestArgs("--verbose --dryrun --source cmdlineapp/sample --dest cmdlineapp/output")
	app.Start()
}
