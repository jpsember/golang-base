package app

import (
	. "github.com/jpsember/golang-base/base"
)

type Oper interface {
	// Get the name of the operation, to select it on the command line from others;
	// if there is only one operation, it can return ""
	UserCommand() string
	// Add help information for this operation
	GetHelp(printer *BasePrinter)
	// Run the operation
	Perform(app *App)
}

// A subtype of Oper that supports json arguments
type OperWithJsonArgs interface {
	Oper
	// Get the default arguments
	GetArguments() DataClass
	// Does an explicit arguments file have to exist, vs using the defaults?
	ArgsFileMustExist() bool
}

// A subtype of Oper that supports additional arguments on the command line,
// other than flags ('-x', '--yyyy')
type OperWithCmdLineArgs interface {
	Oper
	// Handle remaining arguments.  See main.go for an example
	ProcessArgs(c *CmdLineArgs)
}
