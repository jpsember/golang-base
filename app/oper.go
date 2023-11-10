package app

import (
	. "github.com/jpsember/golang-base/base"
)

type Oper interface {
	// Get the name of the operation, to select it on the command line from others;
	// if there is only one operation, it can return ""
	UserCommand() string
	// Get a summary of the operation, and a summary of the arguments
	GetHelp() (summary, usage string)
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
	// Accept the possibly modified arguments for later processing
	AcceptArguments(args DataClass)
}

// A subtype of Oper that supports additional arguments on the command line,
// other than flags ('-x', '--yyyy')
type OperWithCmdLineArgs interface {
	Oper
	// Handle remaining arguments.  See web_server.go for an example
	ProcessArgs(c *CmdLineArgs)
}
