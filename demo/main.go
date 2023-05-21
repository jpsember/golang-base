package main

import (
	. "github.com/jpsember/golang-base/app"
	. "github.com/jpsember/golang-base/base"
)

var _ = Pr

type SpeakOper struct {
}

func (oper *SpeakOper) UserCommand() string {
	return "speak"
}

func (oper *SpeakOper) Perform(app *App) {
	app.Logger().SetVerbose(true)
	pr := app.Logger().Pr
	pr("this is SpeakOper.perform")

	Pr("hello")
}

func (oper *SpeakOper) GetHelp(bp *BasePrinter) {
}

func main() {
	var oper = &SpeakOper{}
	var app = NewApp()
	app.Version = "2.0.3"
	app.CmdLineArgs(). //
				Add("debugging").Desc("perform extra tests"). //
				Add("speed").SetInt().Add("jumping")

	app.RegisterOper(oper)
	app.SetTestArgs("-d --speed 42 --verbose --dryrun --version")
	app.Start()
}
