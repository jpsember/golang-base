package base

import (
	"reflect"
	"strings"
)

type BaseObject struct {
	verbose bool
	name    string
}

func (obj *BaseObject) Name() string {
	if obj.name == "" {
		obj.SetName("<no name defined>")
	}
	return obj.name
}

func (obj *BaseObject) SetName(name string) {
	obj.name = name
}

func (obj *BaseObject) ProvideName(owner any) {
	name := obj.name
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
	obj.SetName(name)
}

func (obj *BaseObject) Log(messages ...any) {
	if obj.verbose {
		Pr(JoinElementToList("["+obj.Name()+":]", messages)...)
	}
}

func (obj *BaseObject) SetVerbose(verbose bool) {
	obj.verbose = verbose
}

func (obj *BaseObject) AlertVerbose() {
	AlertWithSkip(1, "Setting verbosity for:", obj.Name())
	obj.SetVerbose(true)
}

func (obj *BaseObject) Verbose() bool {
	return obj.verbose
}
