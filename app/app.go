package app

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/files"
	"os"
)

var _ = Pr

type App struct {
	logger             Logger
	UserCommand        func() string
	Perform            func()
	RegisterOperations func()
	operMap            map[string]*Oper
	orderedCommands    Array[string]
}

func NewApp() *App {
	var w = new(App)
	w.logger = NewLogger(w)
	w.operMap = make(map[string]*Oper)
	return w
}

func (a *App) Logger() Logger {
	return a.logger
}

func (a *App) RegisterOper(oper *Oper) {
	key := oper.UserCommand()
	_, ok := a.operMap[key]
	CheckState(!ok, "duplicate oper key:", key)
	if ok {

	}
	a.orderedCommands.Add(key)
	a.operMap[key] = oper
	oper.App = a
}

func (a *App) HasMultipleOperations() bool {
	return len(a.operMap) > 1
}

func (a *App) Start() {
	args := os.Args[1:]
	_ = args
	Pr("args:", args)

	var ordered = NewArray[string]()
	for k := range a.operMap {
		ordered.Add(k)
	}

	Todo("sort the Array of oper names")

	Todo("construct and parse command line arguments")
	/**
		<pre>

		    mOperMap = hashMap();
	    mOrderedOperCommands = arrayList();
	    registerOperations();
	    mOrderedOperCommands.sort(null);

	    cmdLineArgs().parse(cmdLineArguments);
	    if (cmdLineArgs().helpShown())
	      return;

	    if (cmdLineArgs().get(CLARG_VERSION)) {
	      pr(getVersion());
	      return;
	    }

	    setVerbose(cmdLineArgs().get(CLARG_VERBOSE));
	    mDryRun = cmdLineArgs().get(CLARG_DRYRUN);
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
}
