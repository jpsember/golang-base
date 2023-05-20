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
	app.RegisterOper(oper)
	app.Start()
}
