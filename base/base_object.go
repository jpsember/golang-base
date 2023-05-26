package base

import (
	"reflect"
	"strings"
)

type Logger struct {
	verbose bool
	name    string
}

func (logger *Logger) Name() string {
	if logger.name == "" {
		logger.SetName("<no name defined>")
	}
	return logger.name
}

func (logger *Logger) SetName(name string) {
	logger.name = name
}

func (logger *Logger) ProvideName(owner any) {
	name := logger.name
	if name != "" {
		return
	}

	if n, ok := owner.(string); ok {
		name = n
	} else {
		name = reflect.TypeOf(owner).String()
	}
	// Remove any pointers or packages
	i := strings.LastIndex(name, "*")
	i = MaxInt(i, strings.LastIndex(name, "."))
	name = name[i+1:]
	CheckArg(name != "")
	logger.SetName(name)
}

func (logger *Logger) Pr(messages ...any) {
	if logger.verbose {
		Pr(JoinLists([]any{"[" + logger.Name() + ":]"}, messages)...)
	}
}

func (logger *Logger) SetVerbose(verbose bool) {
	logger.verbose = verbose
}

func (logger *Logger) Verbose() bool {
	return logger.verbose
}
