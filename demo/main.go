package main

import (
	. "github.com/jpsember/golang-base/app"
	. "github.com/jpsember/golang-base/base"
	gen "github.com/jpsember/golang-base/gen/sample"
)

var _ = Pr

type SpeakOper struct {
	compactMode bool
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

func (oper *SpeakOper) GetArguments() DataClass {
	return gen.DefaultDemoConfig
}

func OurProcAdditionalArgs(args *CmdLineArgs, oper Oper) {
	var sop = oper.(*SpeakOper)

	for args.HasNextArg() {
		var arg = args.NextArg()
		switch arg {
		case "compact":
			Pr("compact mode")
			sop.compactMode = true
		case "height":
			Pr("jump")
		default:
			BadArg("extraneous argument:", arg)
		}
	}
}

func main() {
	var oper = &SpeakOper{}
	var app = NewApp()
	app.Version = "2.0.3"
	app.CmdLineArgs(). //
				Add("debugging").Desc("perform extra tests"). //
				Add("speed").SetInt().Add("jumping")
	app.RegisterOper(oper)
	app.SetTestArgs("-d --speed 42 --verbose --dryrun")
	app.SetTestArgs("--verbose --dryrun height compact compact")
	app.ProcessAdditionalArgs = OurProcAdditionalArgs
	app.Start()
}
