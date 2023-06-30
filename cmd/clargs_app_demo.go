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

	Alert("this is a standard alert", "jello", 42)
	Alert("0:this is a suppressed alert")

	for y := 0; y < 10; y++ {
		Pr(VERT_SP, "y:", y)
		for x := 0; x < 10; x++ {
			j := IntToString(x)
			Alert(j + ":alert priority " + j)
		}
		Alert("#10:repeat 10 times")
		Alert("#2:repeat twice")
		SleepMs(500)
	}
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

func main() {
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
	//app.AddTestArgs("--verbose --dryrun height compact compact zebra height compact")
	app.Start()
}
