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
	if pv == nil {
		return "<nil>"
	}
	return "{Prefix " + Quoted(pv.Prefix) + " state:" + Truncated(pv.State.CompactString()) + " }"
}

func NewStateProvider(prefix string, state JSMap) WidgetStateProvider {
	return &WidgetStateProviderStruct{Prefix: prefix, State: state}
}
