package base

import (
	"reflect"
	"strings"
)

// Interface that supports logging, naming

type Logger interface {
	Name() string
	SetName(name string)
	Pr(messages ...any)
	SetVerbose(verbose bool)
	Verbose() bool
}

type BaseObject interface {
	Logger() Logger
}

type concreteLogger struct {
	name    string
	owner   any
	verbose bool
}

func (c *concreteLogger) Name() string {
	if c.name == "" {
		t := reflect.TypeOf(c.owner)
		var s = t.String()
		s = strings.TrimPrefix(s, "*")
		i := strings.LastIndex(s, ".")
		if i >= 0 {
			s = s[i+1:]
		}
		c.name = s
	}
	return c.name
}

func (c *concreteLogger) SetName(name string) {
	c.name = name
}

func (c *concreteLogger) SetVerbose(flag bool) {
	c.verbose = flag
}

func (c *concreteLogger) Verbose() bool {
	return c.verbose
}

func (c *concreteLogger) Pr(messages ...any) {
	if c.verbose {
		Pr(append([]any{"[", c.Name(), "]:"}, messages...)...)
	}
}

func NewLogger(owner any) Logger {
	x := new(concreteLogger)
	x.owner = owner
	return x
}

// Get a function that prints if verbosity is set
func Printer(obj BaseObject) func(...any) {
	logger := obj.Logger()
	if logger.Verbose() {
		return logger.Pr
	}
	return nullPrinter
}

var nullPrinter = func(...any) {
}
