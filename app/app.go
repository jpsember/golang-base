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
	logger                Logger
	operMap               map[string]Oper
	orderedCommands       Array[string]
	cmdLineArgs           *CmdLineArgs
	Name                  string
	Version               string
	DryRun                bool
	ProcessAdditionalArgs func(c *CmdLineArgs)
	testArgs              []string
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

	Todo("sort the Array of oper names")

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

	a.processArgs()
}

func (a *App) processArgs() {
	var c = a.CmdLineArgs()
	var inf = 0
	for c.HandlingArgs() {
		inf++
		if inf > 100 {
			Die("infinite loop")
		}
		if a.ProcessAdditionalArgs != nil {
			a.ProcessAdditionalArgs(c)
		}
	}

	Todo("Check for no operation selected (only if multiple opers)")
	//  if (app().hasMultipleOperations() && nullOrEmpty(userCommand())) {
	//    throw badArg("No userCommand defined");
	//  }
	//

	Todo("Perform generate args if requested")
	//  if (app().genArgsFlag()) {
	//    AbstractData data = defaultArgs();
	//    if (data == null) {
	//      pr("*** json arguments aren't supported for:", userCommand());
	//    } else {
	//      pr(config());
	//    }
	//    throw new ExitOperImmediately();
	//  }
	//
	Todo("if json arguments are supported, process them")
	//  if (argsSupported()) {
	//    mJsonArgs = defaultArgs();
	//
	//    // Start with the args file that the user supplied as the command line argument (if any)
	//    File argsFile = app().argsFile();
	//    argsFile = Files.subprojectVariant(Files.ifEmpty(argsFile, defaultArgsFilename()));
	//    log("...looking for arguments in:", argsFile);
	//    if (!argsFile.exists()) {
	//      // If there is a version of the args file with underscores instead, raise hell
	//      {
	//        String name = argsFile.getName();
	//        String fixed = name.replace('_', '-');
	//        if (!fixed.equals(name)) {
	//          File fixedFile = new File(Files.parent(argsFile), fixed);
	//          if (fixedFile.exists())
	//            setError("Could not find arguments file:", argsFile,
	//                "but did find one with different spelling:", fixedFile, "(assuming this is a mistake)");
	//        }
	//      }
	//
	//      if (argsFileMustExist())
	//        throw setError("No args file specified, and no default found at:", argsFile);
	//    } else {
	//      mJsonArgs = Files.parseAbstractData(mJsonArgs, argsFile);
	//      if (a.get(App.CLARG_VALIDATE_KEYS)) {
	//        //
	//        // Generate a JSMap A from the parsed arguments
	//        // Re-parse the args file to a JSMap B.
	//        // See if B.keys - A.keys is nonempty... if so, that's a problem.
	//        //
	//        // NOTE: this will only check the top-level JSMap, not any nested maps.
	//        //
	//        Set<String> keysA = mJsonArgs.toJson().asMap().keySet();
	//        JSMap json = JSMap.fromFileIfExists(argsFile);
	//        Set<String> keysB = json.keySet();
	//        keysB.removeAll(keysA);
	//        if (!keysB.isEmpty())
	//          throw setError("Unexpected keys in", argsFile, INDENT, keysB);
	//      }
	//    }
	//
	//    AbstractData argsBuilder = mJsonArgs.toBuilder();
	//
	//    // While a next arg exists, and matches one of the keys in the args map,
	//    // parse a key/value pair as an override
	//    //
	//    while (a.hasNextArg()) {
	//      String key = a.peekNextArg();
	//      Object parsedValue = null;
	//      Accessor accessor = null;
	//      try {
	//        // Attempt to construct a data accessor for a field with this name
	//        accessor = Accessor.dataAccessor(argsBuilder, key);
	//        a.nextArg();
	//      } catch (IllegalArgumentException e) {
	//        log("no accessor built for arg:", key, e.getMessage());
	//        break;
	//      }
	//      Object value = accessor.get();
	//      if (value == null)
	//        throw badArg("Accessor for", quote(key), "returned null; is it optional? They aren't supported");
	//
	//      Class valueClass = accessor.get().getClass();
	//      String valueAsString = null;
	//      if (a.hasNextArg())
	//        valueAsString = a.peekNextArg();
	//
	//      // Special handling for boolean args: if no value given or
	//      // fails parsing (i.e. next arg is some other key/value pair), assume 'true'
	//      //
	//      if (valueClass == Boolean.class) {
	//        if (valueAsString == null)
	//          parsedValue = true;
	//        else {
	//          parsedValue = tryParseAsBoolean(valueAsString);
	//          if (parsedValue != null)
	//            a.nextArg();
	//          else
	//            parsedValue = true;
	//        }
	//      }
	//
	//      if (parsedValue == null) {
	//        if (!a.hasNextArg())
	//          throw badArg("Missing value for command line argument:", key);
	//        valueAsString = a.nextArg();
	//        try {
	//          parsedValue = DataUtil.parseValueFromString(valueAsString, valueClass);
	//        } catch (Throwable t) {
	//          throw badArgWithCause(t, "Failed to parse", quote(key), ":", valueAsString);
	//        }
	//      }
	//      accessor.set(parsedValue);
	//    }
	//    mJsonArgs = argsBuilder.build();
	//  }
	//}
	//

}
func (a *App) auxRunOper(oper Oper) {
	a.processArgs()
	oper.Perform(a)
}
