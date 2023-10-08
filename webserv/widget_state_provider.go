package webserv

import (
	. "github.com/jpsember/golang-base/base"
)

type WidgetStateProviderStruct struct {
	Prefix string // A prefix to remove from keys before accessing the state
	State  JSMap  // The map containing the state
}

type WidgetStateProvider = *WidgetStateProviderStruct

func (pv WidgetStateProvider) String() string {
	return "{SP " + Quoted(pv.Prefix) + " state:" + Truncated(pv.State.CompactString()) + " }"
}

func NewStateProvider(prefix string, state JSEntity) WidgetStateProvider {
	return &WidgetStateProviderStruct{Prefix: prefix, State: state.AsJSMap()}
}

// Read widget value; assumed to be an int.
func readStateIntValue(p WidgetStateProvider, id string) int {
	return p.State.OptInt(compileId(p.Prefix, id), 0)
}

// Read widget value; assumed to be a bool.
func readStateBoolValue(p WidgetStateProvider, id string) bool {
	return p.State.OptBool(compileId(p.Prefix, id), false)
}

// Read widget value; assumed to be a string.
func readStateStringValue(p WidgetStateProvider, id string) string {
	key := compileId(p.Prefix, id)
	if false && Alert("checking for non-existent key") {
		if !p.State.HasKey(key) {
			Pr("State has no key", QUO, key, " (id ", QUO, id, "), state:", INDENT, p.State)
			Pr("prefix:", p.Prefix)
		}
	}
	return p.State.OptString(key, "")
}
