package main

import (
	. "github.com/jpsember/golang-base/app"
	. "github.com/jpsember/golang-base/base"
)

func main() {
	var oper = &ImgOper{}
	oper.ProvideName(oper)
	var app = NewApp()
	app.SetName("encrypt_demo")
	app.Version = "2.1.3"
	app.RegisterOper(oper)
	app.CmdLineArgs()
	app.AddTestArgs("--verbose")
	app.Start()
}

type ImgOper struct {
	BaseObject
}

func (oper *ImgOper) UserCommand() string {
	return "img"
}

func (oper *ImgOper) GetHelp(bp *BasePrinter) {
	bp.Pr("Image manipulation.")
}

func (oper *ImgOper) ProcessArgs(c *CmdLineArgs) {
	for c.HasNextArg() {
		var arg = c.NextArg()
		switch arg {
		default:
			c.SetError("extraneous argument:", arg)
		}
	}
}

func (oper *ImgOper) Perform(app *App) {
	pth := NewPathM("img/resources/0.jpg")
	pth.ReadBytesM()
}
