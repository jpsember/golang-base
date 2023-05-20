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
