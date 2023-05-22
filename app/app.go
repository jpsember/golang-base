package app

import (
	"fmt"
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/files"
	. "github.com/jpsember/golang-base/json"
	"os"
	"reflect"
	"strconv"
	"strings"
)

type App struct {

	// Client app should supply these fields:
	Name                  string
	Version               string
	ProcessAdditionalArgs func(c *CmdLineArgs, oper Oper)
	ArgsFileMustExist     bool

	logger          Logger
	operMap         map[string]Oper
	orderedCommands Array[string]
	cmdLineArgs     *CmdLineArgs

	dryRun        bool
	testArgs      []string
	genArgsFlag   bool
	argsFile      Path
	operArguments DataClass
	errorMessage  []any
}

func NewApp() *App {
	var w = new(App)
	w.logger = NewLogger(w)
	w.operMap = make(map[string]Oper)
	return w
}

func (a *App) Logger() Logger {
	return a.logger
}

const (
	ClArgVerbose  = "verbose"
	ClArgVersion  = "version"
	ClArgDryrun   = "dryrun"
	ClArgGenArgs  = "gen-args"
	ClArgArgsFile = "args"
)

func (a *App) CmdLineArgs() *CmdLineArgs {
	if a.cmdLineArgs != nil {
		return a.cmdLineArgs
	}

	ca := NewCmdLineArgs()
	a.cmdLineArgs = ca

	if false {
		Todo("have optional func pointer to addAppCommandLineArgs")
	}
	// a.addAppCommandLineArgs(ca)
	//for _, oper = range a.operMap {
	//
	//}

	ca.WithBanner("!!! please add a banner !!!")
	ca.Add(ClArgDryrun).Desc("Dry run")
	ca.Add(ClArgVerbose).Desc("Verbose messages").ShortName("v")
	ca.Add(ClArgVersion).Desc("Display version number").ShortName("n")
	ca.Add(ClArgGenArgs).Desc("Generate args for operation")
	ca.Add(ClArgArgsFile).SetString().Desc("Specify arguments file (json)")

	sb := strings.Builder{}
	sb.WriteString(strings.ToLower(a.GetName()))
	sb.WriteString(" version: ")
	sb.WriteString(a.Version)
	sb.WriteString("\n")

	if a.hasMultipleOperations() {
		sb.WriteString("\nUsage: [--<app arg>]* [<operation> <operation arg>*]*\n\n")
		sb.WriteString("Operations:\n")
	}
	for _, key := range a.orderedCommands.Array() {
		oper := a.operMap[key]
		bp := NewBasePrinter()
		oper.GetHelp(bp)
		if !a.hasMultipleOperations() {
			sb.WriteString("\nUsage: " + a.GetName() + " ")
		}
		sb.WriteString(bp.String())
		sb.WriteString("\n")
	}

	if a.hasMultipleOperations() {
		sb.WriteString("\nApp arguments:")
	}
	ca.WithBanner(sb.String())

	return ca
}

func (a *App) SetTestArgs(args string) {
	a.testArgs = strings.Split(args, " ")
}

func (a *App) RegisterOper(oper Oper) {
	key := oper.UserCommand()
	_, ok := a.operMap[key]
	CheckState(!ok, "duplicate oper key:", key)
	if ok {

	}
	a.orderedCommands.Add(key)
	a.operMap[key] = oper
}

func (a *App) hasMultipleOperations() bool {
	return len(a.operMap) > 1
}

func (a *App) GetName() string {
	if a.Name == "" {
		if t := reflect.TypeOf(a); t.Kind() == reflect.Ptr {
			a.Name = t.Elem().Name()
		} else {
			a.Name = "***unknown app name***"
		}
	}
	return a.Name
}

func (a *App) Start() {
	a.auxStart()
	if a.error() {
		fmt.Fprintln(os.Stderr, "*** "+ToString(a.errorMessage...))
		os.Exit(1)
	}
}

func (a *App) auxStart() {
	args := os.Args[1:]

	if a.testArgs != nil {
		args = a.testArgs
		Pr("Using test args:", args)
	}

	var ordered = NewArray[string]()
	for k := range a.operMap {
		ordered.Add(k)
	}

	err := ordered.Sort()
	CheckOk(err)

	var c = a.CmdLineArgs()
	c.Parse(args)
	if c.GetProblem() != nil {
		a.SetError(c.GetProblem()...)
		return
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

	a.Logger().SetVerbose(c.Get(ClArgVerbose))
	a.dryRun = c.Get(ClArgDryrun)
	var pr = Printer(a)

	var oper = a.determineOper()
	a.operArguments = oper.GetArguments()
	if a.operArguments != nil {
		a.genArgsFlag = c.Get(ClArgGenArgs)
		var path = NewPathOrEmptyM(c.GetString(ClArgArgsFile))
		if path.NonEmpty() {
			path.EnsureExists("args file")
		}
		a.argsFile = path
		pr("args file:", path)
	} else {
		pr("no oper arguments were supplied")
	}

	pr("calling processArgs")
	a.processArgs(oper)
	if a.error() {
		return
	}
	var unusedArgs = c.UnusedExtraArgs()
	if len(unusedArgs) != 0 {
		a.SetError("Extraneous arguments:", strings.Join(unusedArgs, ", "))
		return
	}
	oper.Perform(a)
}

// Determine which operation is to be run
func (a *App) determineOper() Oper {
	var pr = Printer(a)

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
			CheckState(oper != nil, "no such operation:", operName)
		} else {
			Pr("*** Please specify an operation ***")
		}
	}
	return oper
}

func (a *App) processArgs(oper Oper) {
	pr := Printer(a)

	var c = a.CmdLineArgs()

	Todo("We probably want to only call ProcessAdditionalArgs if *not* using DataClass arguments")
	for c.HandlingArgs() {
		if a.ProcessAdditionalArgs != nil {
			a.ProcessAdditionalArgs(c, oper)
		}
	}

	if a.genArgsFlag {
		var data = a.operArguments
		// Get default arguments by parsing an empty map
		defaultArgs := data.Parse(NewJSMap())
		Pr(defaultArgs)
		return
	}

	if a.operArguments != nil {
		pr("calling compileDataArgs")
		a.compileDataArgs()
		if a.error() {
			return
		}
	}
}

func (a *App) compileDataArgs() {
	pr := Printer(a)

	// Start with default arguments
	var operArgs = a.operArguments
	pr("...default arguments:", INDENT, operArgs)

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
			if a.ArgsFileMustExist {
				a.SetError("No args file specified, and no default found at:", argsFile)
				return
			}
		}

		operArgs = operArgs.Parse(argsFile.ReadStringIfExistsM("{}"))
		pr("...read from", argsFile, ":", INDENT, operArgs)
	}

	var js = operArgs.ToJson().(*JSMap)

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

		if !c.HasNextArg() {
			BadArg("Missing value for key", Quoted(key))
		}

		var userArg = c.PeekNextArg()
		var consume = true

		// Don't consume the argument if the type of the field is boolean and this
		// doesn't look like a true/false
		if _, ok := value.(JBool); ok {
			if !(userArg == "true" || userArg == "false") {
				consume = false
				userArg = "true"
			}
		}
		if consume {
			c.NextArg()
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

	// Re-parse the arguments from the (possibly modified) jsmap

	operArgs = operArgs.Parse(js)
	Pr("new oper args:", INDENT, operArgs)

	a.operArguments = operArgs
}

func (a *App) SetError(message ...any) {
	if !a.error() {
		a.errorMessage = message
	}
}

func (a *App) error() bool {
	return a.errorMessage != nil
}
