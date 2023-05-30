package main

import (
	. "github.com/jpsember/golang-base/app"
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/gen/sample"
	"github.com/jpsember/golang-base/webserv"
)

var _ = Pr

func main() {

	if true {
		Todo("Todo()", "foo")
		Alert("Alert()", "Investigating caller location and skip counts")
		AlertWithSkip(0, "AlertWithSkip()", "Investigating caller location and skip counts")
		Pr(CallerLocation(0), "CallerLocation")
		return
		webserv.WebServerDemo()
		return
	}

	jsonExample()

	cmdLineExample()
}

// -------------------------------------------------------------------------

type SpeakOper struct {
	BaseObject
	compactMode bool
	config      DemoConfig
}

func (oper *SpeakOper) UserCommand() string {
	return "speak"
}

func (oper *SpeakOper) Perform(app *App) {
	oper.SetVerbose(true)
	pr := oper.Log
	pr("this is SpeakOper.perform")
	pr("Arguments:", INDENT, oper.config)
}

func (oper *SpeakOper) GetHelp(bp *BasePrinter) {
	bp.Pr("An example of an app that takes json (data class) arguments.")
}

func (oper *SpeakOper) GetArguments() DataClass {
	return DefaultDemoConfig
}

func (oper *SpeakOper) ArgsFileMustExist() bool { return false }
func (oper *SpeakOper) AcceptArguments(a DataClass) {
	oper.config = a.(DemoConfig)
}

func jsonExample() {
	Pr(VERT_SP, DASHES, "jsonExample", CR, DASHES)
	var oper = &SpeakOper{}
	oper.ProvideName(oper)
	var app = NewApp()
	app.SetName("jsonExample")
	app.Version = "2.0.3"
	app.CmdLineArgs(). //
				Add("debugging").Desc("perform extra tests"). //
				Add("speed").SetInt().Add("jumping")
	app.RegisterOper(oper)
	app.SetTestArgs("-d --speed 42 --verbose --dryrun target 18 simulate")
	app.Start()
}

// -------------------------------------------------------------------------

type JumpOper struct {
	BaseObject
	compactMode bool
}

func (oper *JumpOper) UserCommand() string {
	return "jump"
}

func (oper *JumpOper) Perform(app *App) {
	oper.SetVerbose(true)
	pr := oper.Log
	pr("this is JumpOper.perform")
	Pr("goodbye")
}

func (oper *JumpOper) GetHelp(bp *BasePrinter) {
	bp.Pr("An example of an app that uses conventional command line arguments only.")
}

func (oper *JumpOper) ProcessArgs(c *CmdLineArgs) {
	for c.HasNextArg() {
		var arg = c.NextArg()
		switch arg {
		case "compact":
			Pr("compact mode")
			oper.compactMode = true
		case "height":
			Pr("jump")
		default:
			c.SetError("extraneous argument:", arg)
		}
	}
}

func cmdLineExample() {
	Pr(VERT_SP, DASHES, "cmdLineExample", CR, DASHES)
	var oper = &JumpOper{}
	oper.ProvideName(oper)
	var app = NewApp()
	app.SetName("cmd_line_example")
	app.Version = "2.1.3"
	app.RegisterOper(oper)
	app.CmdLineArgs(). //
				Add("debugging").Desc("perform extra tests"). //
				Add("speed").SetInt().Add("jumping")          //
	app.SetTestArgs("--verbose --dryrun height compact compact zebra height compact --help")
	app.Start()
}
