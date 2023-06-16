package main

import (
	. "github.com/jpsember/golang-base/app"
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/gen/sample"
)

var _ = Pr

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

func main() {
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
	app.AddTestArgs("-d --speed 42 --verbose --dryrun target 18 simulate -g")
	app.Start()
}
