package webserv

import (
	. "github.com/jpsember/golang-base/base"
)

type ClientValueObj struct {
	values  []string
	problem string
}

type ClientValue = *ClientValueObj

func MakeClientValue(values []string) ClientValue {
	c := ClientValueObj{
		values: values,
	}
	return &c
}

func (c ClientValue) SetProblem(message ...any) ClientValue {
	if c.problem == "" {
		c.problem = "Problem with ajax request: " + ToString(message...)
		Pr("Setting problem with ClientValue:", INDENT, c.problem)
	}
	return c
}

func (c ClientValue) GetString() string {
	if c.problem == "" {
		if len(c.values) == 1 {
			return c.values[0]
		}
		c.SetProblem("Expected single string")
	}
	return ""
}

func (c ClientValue) Ok() bool {
	return c.problem == ""
}
