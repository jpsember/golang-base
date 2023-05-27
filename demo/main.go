package main

import (
	"fmt"
	. "github.com/jpsember/golang-base/app"
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/gen/sample"
	. "github.com/jpsember/golang-base/gen/webservgen"
	"github.com/jpsember/golang-base/webserv"
)

var _ = Pr

func main() {

	if false {

		// Nil not always nil... sounds like a huge language code smell
		//
		// https://stackoverflow.com/questions/60733102

		var p *int        // (type=*int,value=nil)
		var i interface{} // (type=nil,value=nil)

		if i != nil { // (type=nil,value=nil) != (type=nil,value=nil)
			fmt.Println("a not a nil")
		}

		i = p // assign p to i

		// a hardcoded nil is always nil,nil (type,value)
		if i != nil { // (type=*int,value=nil) != (type=nil,value=nil)
			fmt.Println("b not a nil")
		}

		return
	}
	if false {
		s := DefaultSession
		Pr("session:", s)

		b := s.ToBuilder()
		Pr("builder:", b)

		jm := s.ToJson()
		Pr("json:", INDENT, jm)

		c := DefaultSession.Parse(jm).(Session)
		Pr("c:", INDENT, c)

		b2 := c.ToBuilder()
		Pr("b2:", b2)

		b3 := b2.ToBuilder()
		Pr("b3:", b3)

		return

	}
	if true {
		webserv.Demo()
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
