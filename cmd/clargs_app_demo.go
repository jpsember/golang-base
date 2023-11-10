package main

import (
	. "github.com/jpsember/golang-base/app"
	. "github.com/jpsember/golang-base/base"
)

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

func (oper *JumpOper) GetHelp() (summary, usage string) {
	summary = "An example of an app that uses conventional command line arguments only."
	return
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

func main() {
	Pr(VERT_SP, DASHES, "cmdLineExample", CR, DASHES)

	var oper = &JumpOper{}
	oper.ProvideName(oper)
	var app = NewApp()
	app.SetName("cmd_line_example")
	app.Version = "2.1.3"
	app.RegisterOper(oper)

	if false {
		oper2 := &FooOper{}
		app.RegisterOper(oper2)
	}
	app.CmdLineArgs(). //
				Add("debugging").Desc("perform extra tests"). //
				Add("speed").SetInt().Add("jumping")          //

	//app.AddTestArgs("--verbose --dryrun height compact compact zebra height compact")
	app.AddTestArgs("--help")
	app.Start()
}

type FooOper struct {
	BaseObject
}

func (oper *FooOper) UserCommand() string {
	return "moo"
}

func (oper *FooOper) Perform(app *App) {
	oper.SetVerbose(true)
	pr := oper.Log
	pr("this is FooOper.perform")
	Pr("goodbye")
}

func (oper *FooOper) GetHelp() (summary, usage string) {
	summary = "This is help for FooOper."
	usage = "blargh <directory> count <int>"
	return
}

func (oper *FooOper) ProcessArgs(c *CmdLineArgs) {
}
