package app

import (
	. "github.com/jpsember/golang-base/base"
)

var _ = Pr

type Oper struct {
	UserCommand string
	Perform     func()
	App         *App
}

func NewOper() *Oper {
	var w = new(Oper)
	return w
}

func (a *Oper) Logger() Logger {
	return a.App.Logger()
}
