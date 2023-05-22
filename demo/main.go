package main

import (
	. "github.com/jpsember/golang-base/app"
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/gen/sample"
)

var _ = Pr

func main() {

	jsonExample()

	cmdLineExample()
}

// -------------------------------------------------------------------------

type SpeakOper struct {
	compactMode bool
	config      DemoConfig
}

func (oper *SpeakOper) UserCommand() string {
	return "speak"
}

func (oper *SpeakOper) Perform(app *App) {
	app.Logger().SetVerbose(true)
	pr := app.Logger().Pr
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
	var oper = &SpeakOper{}
	var app = NewApp()
	app.Version = "2.0.3"
	app.CmdLineArgs(). //
				Add("debugging").Desc("perform extra tests"). //
				Add("speed").SetInt().Add("jumping")
	app.RegisterOper(oper)
	app.SetTestArgs("-d --speed 42 --verbose --dryrun target 18 simulate --help")
	Todo("help message isn't being shown")
	app.Start()
}

// -------------------------------------------------------------------------

type JumpOper struct {
	compactMode bool
}

func (oper *JumpOper) UserCommand() string {
	return "jump"
}

func (oper *JumpOper) Perform(app *App) {
	app.Logger().SetVerbose(true)
	pr := app.Logger().Pr
	pr("this is JumpOper.perform")
	Pr("goodbye")
}

func (oper *JumpOper) GetHelp(bp *BasePrinter) {
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
			//BadArg("extraneous argument:", arg)
		}
	}
}

func cmdLineExample() {
	var oper = &JumpOper{}
	var app = NewApp()
	app.Version = "2.1.3"
	app.CmdLineArgs(). //
				Add("debugging").Desc("perform extra tests"). //
				Add("speed").SetInt().Add("jumping")
	app.RegisterOper(oper)
	app.SetTestArgs("--verbose --dryrun height compact compact zebra height compact")
	app.Start()
}
