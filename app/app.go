package app

import (
	"fmt"
	. "github.com/jpsember/golang-base/base"
	"os"
	"strconv"
	"strings"
)

type App struct {
	BaseObject

	// Client app should supply these fields:
	Version string

	operMap         map[string]Oper
	orderedCommands Array[string]
	cmdLineArgs     *CmdLineArgs

	dryRun              bool
	testArgs            *Array[string]
	genArgsFlag         bool
	argsFile            Path
	operDataClassArgs   DataClass
	errorMessage        []any
	operWithJsonArgs    OperWithJsonArgs
	operWithCmdLineArgs OperWithCmdLineArgs
	oper                Oper
	startDir            Path
}

func NewApp() *App {
	var w = new(App)
	w.operMap = make(map[string]Oper)
	return w
}

const (
	ClArgVerbose  = "verbose"
	ClArgVersion  = "version"
	ClArgDryrun   = "dryrun"
	ClArgGenArgs  = "gen-args"
	ClArgArgsFile = "args"
	ClIDE         = "ide"
	ClStartDir    = "startdir"
)

func (a *App) CmdLineArgs() *CmdLineArgs {
	if a.cmdLineArgs != nil {
		return a.cmdLineArgs
	}

	ca := NewCmdLineArgs()
	a.cmdLineArgs = ca

	ca.WithBanner("!!! please add a banner !!!")
	ca.Add(ClIDE).Desc("Running within IDE").ShortName("I")
	ca.Add(ClArgDryrun).Desc("Dry run")
	ca.Add(ClArgVerbose).Desc("Verbose messages").ShortName("v")
	ca.Add(ClArgVersion).Desc("Display version number").ShortName("n")
	ca.Add(ClArgGenArgs).Desc("Generate args for operation").ShortName("g")
	ca.Add(ClArgArgsFile).SetString().Desc("Specify arguments file (json)")
	ca.Add(ClStartDir).SetString().Desc("Directory to start within").ShortName("S")

	sb := strings.Builder{}
	sb.WriteString(a.Name())
	sb.WriteString(" version: ")
	sb.WriteString(a.Version)
	sb.WriteString("\n\n")

	if a.hasMultipleOperations() {
		sb.WriteString("Usage: [--<app arg>]* [<operation> <operation arg>*]*\n\n")
		sb.WriteString("Operations:\n\n")
	}
	for _, key := range a.orderedCommands.Array() {
		oper := a.operMap[key]
		bp := NewBasePrinter()
		oper.GetHelp(bp)
		sb.WriteString(bp.String())
		if !a.hasMultipleOperations() {
			sb.WriteString("\n\nUsage: " + a.Name())
		}
		sb.WriteString("\n")
	}

	if a.hasMultipleOperations() {
		sb.WriteString("\nApp arguments:")
	}
	ca.WithBanner(sb.String())
	return ca
}

// Parse a string of arguments separated by whitespace, and add to a list of test arguments
func (a *App) AddTestArgs(args string) *App {
	if a.testArgs == nil {
		a.testArgs = NewArray[string]()
	}
	items := strings.Split(args, " ")
	for _, k := range items {
		k := strings.TrimSpace(k)
		if k == "" {
			continue
		}
		a.testArgs.Add(k)
	}
	return a
}

func (a *App) HasTestArgs() bool {
	return a.testArgs != nil
}

func AssertJsonOper(oper Oper) OperWithJsonArgs {
	result, ok := oper.(OperWithJsonArgs)
	CheckArg(ok, "oper does not support OperWithJsonArgs interface:", oper)
	return result
}

func AssertCmdLineOper(oper Oper) OperWithCmdLineArgs {
	result, ok := oper.(OperWithCmdLineArgs)
	CheckArg(ok, "oper does not support OperWithCmdLineArgs interface:", oper)
	return result
}

func (a *App) RegisterOper(oper Oper) {
	key := oper.UserCommand()
	_, ok := a.operMap[key]
	CheckState(!ok, "duplicate oper key:", key)
	a.orderedCommands.Add(key)
	a.operMap[key] = oper
}

func (a *App) hasMultipleOperations() bool {
	return len(a.operMap) > 1
}

func (a *App) Start() {
	a.auxStart()
	if a.error() {
		fmt.Fprintln(os.Stderr, "*** "+ToString(a.errorMessage...))
		os.Exit(1)
	}
}

func (a *App) auxStart() {
	defer SharedBackgroundTaskManager().Stop()
	args := os.Args[1:]

	if a.testArgs != nil {
		args = a.testArgs.Array()
	}

	var ordered = NewArray[string]()
	for k := range a.operMap {
		ordered.Add(k)
	}

	CheckOk(ordered.Sort())

	var c = a.CmdLineArgs()
	c.Parse(args)
	if a.handleCmdLineArgsError() {
		return
	}

	if c.Get(ClIDE) {
		// Clear the console
		fmt.Print("\033c")
	}

	// If we showed the help, exit
	if c.HelpShown() {
		return
	}

	// If user wants the version number, print it and exit
	if c.Get(ClArgVersion) {
		var vers = a.Version
		if vers == "" {
			Pr("*** no version specified ***")
		} else {
			Pr(vers)
		}
		return
	}

	a.SetVerbose(c.Get(ClArgVerbose))
	a.dryRun = c.Get(ClArgDryrun)
	var pr = a.Log

	a.determineOper()
	if a.oper == nil {
		return
	}

	if a.operWithJsonArgs != nil {
		a.operDataClassArgs = a.operWithJsonArgs.GetArguments()
		CheckArg(a.operDataClassArgs != nil, "No arguments returned by oper")
		a.genArgsFlag = c.Get(ClArgGenArgs)
		var path = NewPathOrEmptyM(c.GetString(ClArgArgsFile))

		if path.Empty() {
			// Look for a default args file, <opername>-args.json
			defaultArgsPath := NewPathM(a.oper.UserCommand() + "-args.json")
			if defaultArgsPath.Exists() {
				path = defaultArgsPath
			}
		}

		if path.NonEmpty() {
			path.EnsureExists("args file")
			// If no explicit start directory was given, use the directory containing the arguments
			// Convert to an absolute path before, to ensure a parent directory is known
			a.SpecifyStartDir(path.GetAbsM().Parent())
		}
		a.argsFile = path
		pr("args file:", path)
	}

	pr("calling processArgs")
	a.processArgs()
	if a.error() {
		return
	}
	if a.genArgsFlag {
		return
	}
	var unusedArgs = c.UnusedExtraArgs()
	if len(unusedArgs) != 0 {
		a.SetError("Extraneous arguments:", strings.Join(unusedArgs, ", "))
		return
	}
	pr("calling oper.Perform")
	a.oper.Perform(a)
}

// TODO: this can probably be private
func (a *App) SpecifyStartDir(path Path) {
	if a.startDir.NonEmpty() {
		return
	}
	if path.String() == "/" {
		BadArg("start dir is system root:", path)
	}
	a.startDir = path.AssertNonEmpty()
}

func (a *App) StartDir() Path {
	if a.startDir.Empty() {
		var pth Path
		startDir := a.CmdLineArgs().GetString(ClStartDir)
		if startDir != "" {
			pth = NewPathM(startDir)
		} else {
			pth = CurrentDirectory()
		}
		a.SpecifyStartDir(pth)
	}
	return a.startDir
}

func (a *App) handleCmdLineArgsError() bool {
	if !a.error() {
		var c = a.CmdLineArgs()
		if c.HasError() {
			a.SetError(c.GetError()...)
		}
	}
	return a.error()
}

// Determine which operation is to be run
func (a *App) determineOper() {
	var pr = a.Log

	var c = a.CmdLineArgs()

	var oper Oper
	if !a.hasMultipleOperations() {
		CheckState(a.orderedCommands.NonEmpty(), "no operations defined")
		oper = a.operMap[a.orderedCommands.Get(0)]
		pr("single operation")
	} else {
		if c.HasNextArg() {
			var operName = c.NextArg()
			pr("looking for operation named:", operName)
			oper = a.operMap[operName]
			if oper == nil {
				a.SetError("no such operation:", Quoted(operName))
			}
		} else {
			Pr("*** Please specify an operation ***")
		}
	}
	if oper != nil {
		a.oper = oper
		if x, ok := oper.(OperWithJsonArgs); ok {
			a.operWithJsonArgs = x
		}
		if x, ok := oper.(OperWithCmdLineArgs); ok {
			a.operWithCmdLineArgs = x
		}
	}
}

func (a *App) processArgs() {
	pr := a.Log

	var c = a.CmdLineArgs()

	operc := a.operWithCmdLineArgs
	operj := a.operWithJsonArgs

	if operc != nil {
		for c.handlingArgs() {
			operc.ProcessArgs(c)
		}
		if a.handleCmdLineArgsError() {
			return
		}
	}

	if a.genArgsFlag {
		if operj != nil {
			var data = a.operDataClassArgs
			// Get default arguments by parsing an empty map
			defaultArgs := data.Parse(NewJSMap())
			Pr(defaultArgs)
		} else {
			Pr("Unavailable for this operation")
		}
		return
	}

	if a.operDataClassArgs != nil {
		pr("calling compileDataArgs")
		a.compileDataArgs()
		if a.error() {
			return
		}
	}
}

func (a *App) compileDataArgs() {
	pr := a.Log

	var oper = a.operWithJsonArgs

	// Start with default arguments
	var operArgs = a.operDataClassArgs
	//pr("...default arguments:", INDENT, operArgs)

	// Replace with args file, if there was one
	if a.argsFile.NonEmpty() {
		argsFile := a.argsFile

		// Todo: add support for subprojects
		//    argsFile = Files.subprojectVariant(Files.ifEmpty(argsFile, defaultArgsFilename()));

		pr("...looking for arguments in:", argsFile)
		if !argsFile.Exists() {
			// If there is a version of the args file with underscores instead, raise hell
			name := argsFile.Base()
			fixed := strings.ReplaceAll(name, "_", "-")
			if fixed != name {
				fixedFile := argsFile.Parent().JoinM(fixed)
				if fixedFile.Exists() {
					a.SetError("Could not find arguments file:", argsFile,
						"but did find one with different spelling:", fixedFile, "(assuming this is a mistake)")
					return
				}
			}
			//
			if oper.ArgsFileMustExist() {
				a.SetError("No args file specified, and no default found at:", argsFile)
				return
			}
		}
		argsJSMap := JSMapFromFileIfExistsM(argsFile)
		operArgs = operArgs.Parse(argsJSMap)
	}

	var js = operArgs.ToJson().(*JSMapStruct)

	// While a next arg exists, and matches one of the keys in the args map,
	// parse a key/value pair as an override

	var c = a.CmdLineArgs()
	for c.HasNextArg() {
		var key = c.PeekNextArg()
		value := js.OptAny(key)
		if value == nil {
			Pr("...can't find key:", Quoted(key), "in operation arguments")
			break
		}
		c.NextArg()

		var userArg string

		// If it's a boolean argument, the value is optional.
		// Don't consume the next argument if the type of the field is boolean and the
		// argument doesn't look like a true/false
		if _, ok := value.(JBool); ok {
			var nextArg = c.PeekNextArgOr("")
			if !(nextArg == "true" || nextArg == "false") {
				userArg = "true"
			}
		}

		if userArg == "" {
			if !c.HasNextArg() {
				c.SetError("Missing value for key", Quoted(key))
				break
			}
			userArg = c.NextArg()
		}

		var newVal JSEntity

		// Determine the type of the field
		switch t := value.(type) {
		case JInteger:
			val, err := strconv.Atoi(userArg)
			if err != nil {
				a.SetError("Problem with command line arguments; unable to convert", userArg, "to integer")
				return
			}
			newVal = MakeJInteger(int64(val))
		case JFloat:
			val, err := strconv.ParseFloat(userArg, 64)
			if err != nil {
				a.SetError("Problem with command line arguments; unable to convert", userArg, "to float")
				return
			}
			newVal = MakeJFloat(val)
		case JBool:
			switch userArg {
			case "true":
				newVal = JBoolTrue
			case "false":
				newVal = JBoolFalse
			default:
				BadArg("should not have happened")
			}
		case JString:
			newVal = MakeJString(userArg)
		default:
			a.SetError("Problem with command line arguments; unsupported value for key", Quoted(key), ":", t)
			return
		}

		// Replace the value within the json map
		js.Put(key, newVal)
	}
	if a.handleCmdLineArgsError() {
		return
	}

	// Re-parse the arguments from the (possibly modified) jsmap

	operArgs = operArgs.Parse(js)
	//pr("...modified arguments:", INDENT, operArgs)

	a.operDataClassArgs = operArgs // Replace the previous version, though I don't think this field is used past this point
	oper.AcceptArguments(operArgs)
}

func (a *App) SetError(message ...any) {
	if !a.error() {
		a.errorMessage = message
	}
}

func (a *App) error() bool {
	return a.errorMessage != nil
}
