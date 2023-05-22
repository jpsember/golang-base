package app

import (
	. "github.com/jpsember/golang-base/base"
)

var _ = Pr

type Oper interface {
	UserCommand() string
	Perform(app *App)
	GetHelp(printer *BasePrinter)
}

type OperWithArguments interface {
	Oper
	GetArguments() DataClass
	ArgsFileMustExist() bool
}

type OperWithCmdLineArgs interface {
	Oper
	ProcessAdditionalArgs(c *CmdLineArgs)
}
