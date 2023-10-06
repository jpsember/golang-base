package webserv

import (
	. "github.com/jpsember/golang-base/base"
)

type PageArgsStruct struct {
	args    []string
	cursor  int
	problem bool
}

type PageArgs = *PageArgsStruct

func PageArgsWith(args ...any) PageArgs {
	var strargs []string
	for _, a := range args {
		var s string
		switch t := a.(type) {
		case string:
			s = t
		case int:
			s = IntToString(t)
		}
		strargs = append(strargs, s)
	}
	return NewPageArgs(strargs)
}

func NewPageArgs(args []string) PageArgs {
	if args == nil {
		args = []string{}
	}
	t := &PageArgsStruct{
		args: args,
	}
	return t
}

func (p PageArgs) CheckDone() bool {
	if !p.Done() {
		p.SetProblem()
	}
	return !p.Problem()
}

func (p PageArgs) Done() bool {
	return p.cursor == len(p.args)
}

func (p PageArgs) Next() string {
	value := p.Peek()
	if value != "" {
		p.cursor++
	}
	return value
}

func (p PageArgs) Peek() string {
	var result string
	if !p.Problem() && !p.Done() {
		return p.args[p.cursor]
	}
	return result
}

func (p PageArgs) Problem() bool {
	return p.problem
}

func (p PageArgs) SetProblem() {
	p.problem = true
}

func (p PageArgs) Int() int {
	result := -1
	if p.Done() {
		p.SetProblem()
	} else {
		a := p.Next()
		val, err := ParseInt(a)
		if err != nil {
			p.SetProblem()
		} else {
			result = val
		}
	}
	return result
}

func (p PageArgs) PositiveInt() int {
	result := p.Int()
	if result <= 0 {
		p.SetProblem()
	}
	return result
}

func (p PageArgs) ReadIf(value string) bool {
	if p.Peek() == value {
		p.Next()
		return true
	}
	return false
}
