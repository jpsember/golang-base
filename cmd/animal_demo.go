package main

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/app"
	. "github.com/jpsember/golang-base/webapp"
)

func main() {
	WrapMain(auxMain)
}

func auxMain() {
	//ClearAlertHistory()
	//SetWidgetDebugRendering()

	var app = NewApp()
	app.SetName("Animal Demo")
	app.Version = "1.0"
	app.CmdLineArgs().Add("insecure").Desc("insecure (http) mode")

	app.RegisterOper(&AnimalOperStruct{
		//FullWidth: true,
		TopPadding: 5,
	})
	app.Start()
}
