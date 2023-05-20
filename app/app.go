package app

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/files"
	"os"
	"reflect"
	"strings"
)

var _ = Pr

type App struct {
	logger          Logger
	operMap         map[string]Oper
	orderedCommands Array[string]
	cmdLineArgs     *CmdLineArgs
	Name            string
	Version         string
	DryRun          bool
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
	CLARG_VERBOSE = "verbose"
	CLARG_VERSION = "version"
	CLARG_DRYRUN  = "dryrun"
	//CLARG_GEN_ARGS = "gen-args"
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

	Todo("have optional func pointer to addAppCommandLineArgs")
	// a.addAppCommandLineArgs(ca)
	//for _, oper = range a.operMap {
	//
	//}

	ca.WithBanner("!!! please add a banner !!!")
	ca.Add(CLARG_DRYRUN).Desc("Dry run")
	ca.Add(CLARG_VERBOSE).Desc("Verbose messages").ShortName("v")
	ca.Add(CLARG_VERSION).Desc("Display version number").ShortName("n")

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
	//_ = args
	//Pr("args:", args)

	var ordered = NewArray[string]()
	for k := range a.operMap {
		ordered.Add(k)
	}

	a.operMap[ordered.Get(0)].Perform(a)
	Todo("sort the Array of oper names")

	Todo("construct and parse command line arguments")

	var c = a.CmdLineArgs()
	c.Parse(args)
	if c.HelpShown() {
		return
	}
	if c.Get(CLARG_VERSION) {
		Pr(a.Version)
		return
	}
	a.Logger().SetVerbose(c.Get(CLARG_VERBOSE))
	a.DryRun = c.Get(CLARG_DRYRUN)

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

	if !a.HasMultipleOperations() {
		CheckState(a.orderedCommands.NonEmpty(), "no operations defined")
		var oper = a.operMap[a.orderedCommands.Get(0)]
		a.auxRunOper(oper)
	} else {
		for c.HasNextArg() {
			var operation = c.NextArg()
			var oper = a.operMap[operation]
			CheckState(oper != nil, "no such operation:", operation)
			a.auxRunOper(oper)
		}
	}

	if c.HasNextArg() {
		Pr("*** Ignoring remaining arguments:", c.ExtraArgs())
	}
}

func (a *App) auxRunOper(oper Oper) {
	oper.ProcessArgs(a.cmdLineArgs)
	oper.Perform(a)
}
