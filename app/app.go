package app

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/files"
	. "github.com/jpsember/golang-base/json"
	"os"
	"reflect"
	"strconv"
	"strings"
)

var _ = Pr

type App struct {
	logger                Logger
	operMap               map[string]Oper
	orderedCommands       Array[string]
	cmdLineArgs           *CmdLineArgs
	Name                  string
	Version               string
	DryRun                bool
	ProcessAdditionalArgs func(c *CmdLineArgs)
	testArgs              []string
	genArgsFlag           bool
	argsFile              Path
	oper                  Oper
	operArguments         DataClass
	gotOperArguments      bool
	ArgsFileMustExist     bool
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
	CLARG_VERBOSE   = "verbose"
	CLARG_VERSION   = "version"
	CLARG_DRYRUN    = "dryrun"
	CLARG_GEN_ARGS  = "gen-args"
	CLARG_ARGS_FILE = "args"
	//CLARG_SHOW_EXCEPTIONS = "exceptions"
	//CLARG_VALIDATE_KEYS = "check-keys"
)

func (a *App) CmdLineArgs() *CmdLineArgs {
	if a.cmdLineArgs != nil {
		return a.cmdLineArgs
	}

	ca := NewCmdLineArgs()
	a.cmdLineArgs = ca

	Todo("add support for args file")
	//if (supportArgsFile()) {
	//   ca.add(CLARG_ARGS_FILE).def("").desc("Specify file containing arguments").shortName("a");
	//   ca.add(CLARG_VALIDATE_KEYS).desc("Check for extraneous keys").shortName("K");
	//   ca.add(CLARG_GEN_ARGS).desc("Generate default operation arguments").shortName("g");
	// }

	if false {
		Todo("have optional func pointer to addAppCommandLineArgs")
	}
	// a.addAppCommandLineArgs(ca)
	//for _, oper = range a.operMap {
	//
	//}

	ca.WithBanner("!!! please add a banner !!!")
	ca.Add(CLARG_DRYRUN).Desc("Dry run")
	ca.Add(CLARG_VERBOSE).Desc("Verbose messages").ShortName("v")
	ca.Add(CLARG_VERSION).Desc("Display version number").ShortName("n")
	ca.Add(CLARG_GEN_ARGS).Desc("Generate args for operation")
	ca.Add(CLARG_ARGS_FILE).SetString().Desc("Specify arguments file (json)")

	sb := strings.Builder{}
	sb.WriteString(strings.ToLower(a.GetName()))
	sb.WriteString(" version: ")
	sb.WriteString(a.Version)
	sb.WriteString("\n")

	if a.HasMultipleOperations() {
		sb.WriteString("\nUsage: [--<app arg>]* [<operation> <operation arg>*]*\n\n")
		sb.WriteString("Operations:\n")
	}
	for _, key := range a.orderedCommands.Array() {
		oper := a.operMap[key]
		bp := NewBasePrinter()
		oper.GetHelp(bp)
		if !a.HasMultipleOperations() {
			sb.WriteString("\nUsage: " + a.GetName() + " ")
		}
		sb.WriteString(bp.String())
		sb.WriteString("\n")
	}

	if a.HasMultipleOperations() {
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

func (a *App) HasMultipleOperations() bool {
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

	// If we showed the help, exit
	if c.HelpShown() {
		return
	}

	// If user wants the version number, print it and exit
	if c.Get(CLARG_VERSION) {
		var vers = a.Version
		if vers == "" {
			Pr("*** no version specified ***")
		} else {
			Pr(vers)
		}
		return
	}

	a.Logger().SetVerbose(c.Get(CLARG_VERBOSE))
	a.DryRun = c.Get(CLARG_DRYRUN)

	// Determine which operation is to be run

	var oper Oper
	if !a.HasMultipleOperations() {
		CheckState(a.orderedCommands.NonEmpty(), "no operations defined")
		oper = a.operMap[a.orderedCommands.Get(0)]
	} else {
		if c.HasNextArg() {
			var operation = c.NextArg()
			oper = a.operMap[operation]
			CheckState(oper != nil, "no such operation:", operation)
		} else {
			Pr("*** Please specify an operation ***")
			return
		}
	}
	a.oper = oper

	if a.getOperArguments() != nil {
		a.genArgsFlag = c.Get(CLARG_GEN_ARGS)
		var path = NewPathOrEmptyM(c.GetString(CLARG_ARGS_FILE))
		if path.NonEmpty() {
			path.EnsureExists("args file")
		}
		a.argsFile = path
	}

	a.processArgs()

	/**
		<pre>


	    if (supportArgsFile()) {
	      mGenArgs = cmdLineArgs().get(CLARG_GEN_ARGS);
	      mArgsFile = new File(cmdLineArgs().getString(CLARG_ARGS_FILE));
	      if (Files.nonEmpty(mArgsFile))
	        Files.assertExists(mArgsFile, "args file");
	    } else
	      mArgsFile = Files.DEFAULT;

	    try {
	      runApplication(cmdLineArgs());
	    } catch (AppErrorException e) {
	    }


		</pre>
	*/

	//if !a.HasMultipleOperations() {
	//	CheckState(a.orderedCommands.NonEmpty(), "no operations defined")
	//	var oper = a.operMap[a.orderedCommands.Get(0)]
	//	a.auxRunOper(oper)
	//} else {
	//	for c.HasNextArg() {
	//		var operation = c.NextArg()
	//		var oper = a.operMap[operation]
	//		CheckState(oper != nil, "no such operation:", operation)
	//		a.auxRunOper(oper)
	//	}
	//}

	if c.HasNextArg() {
		Pr("*** Ignoring remaining arguments:", c.ExtraArgs())
	}

}

func (a *App) processArgs() {
	var c = a.CmdLineArgs()
	for c.HandlingArgs() {
		if a.ProcessAdditionalArgs != nil {
			a.ProcessAdditionalArgs(c)
		}
	}

	if a.genArgsFlag {
		var data = a.getOperArguments()
		// Get default arguments by parsing an empty map
		defaultArgs := data.Parse(NewJSMap())
		Pr(defaultArgs)
		return
	}

	if a.getOperArguments() != nil {
		a.compileDataArgs()
	}
}
func (a *App) compileDataArgs() {
	pr := Printer(a)

	var operArgs = a.getOperArguments()

	// Start with the args file that the user supplied as the command line argument (if any)
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
					BadArg("Could not find arguments file:", argsFile, "but did find one with different spelling:", fixedFile, "(assuming this is a mistake)")
				}
			}
			//
			if a.ArgsFileMustExist {
				BadArg("No args file specified, and no default found at:", argsFile)
			}
		}

		operArgs = operArgs.Parse(argsFile.ReadStringIfExistsM("{}"))

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
			var userArg = c.NextArg()

			var newVal JSEntity

			// Determine the type of the field
			switch t := value.(type) {
			case JInteger:
				val, err := strconv.Atoi(userArg)
				CheckOk(err, "failed to convert to integer:", userArg)
				newVal = MakeJInteger(int64(val))
			case JFloat:
				val, err := strconv.ParseFloat(userArg, 64)
				CheckOk(err, "failed to convert to float64:", userArg)
				newVal = MakeJFloat(float64(val))
			case JBool:
				switch userArg {
				case "t", "true":
					newVal = JBoolTrue
				case "f", "false":
					newVal = JBoolFalse
				default:
					BadArg("Bad bool value for key", Quoted(key), ":", Quoted(userArg))
				}
			case JString:
				newVal = MakeJString(userArg)
			default:
				BadState("Unsupported value for key", Quoted(key), ":", t)
			}

			// Replace the value within the json map
			js.Put(key, newVal)
		}

		// Re-parse the arguments from the (possibly modified) jsmap

		Pr("about to re-parse:", INDENT, js)
		operArgs = operArgs.Parse(js)
		Pr("new oper args:", INDENT, operArgs)

		a.operArguments = operArgs
	}
}

func (a *App) auxRunOper(oper Oper) {
	a.processArgs()
	oper.Perform(a)
}

func (a *App) getOperArguments() DataClass {
	if !a.gotOperArguments {
		a.operArguments = a.oper.GetArguments()
		a.gotOperArguments = true
	}
	return a.operArguments
}
