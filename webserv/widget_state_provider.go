package webserv

import (
	. "github.com/jpsember/golang-base/base"
)

type WidgetStateProviderStruct struct {
	State JSMap // The map containing the state
}

type WidgetStateProvider = *WidgetStateProviderStruct

func (pv WidgetStateProvider) String() string {
	if pv == nil {
		return "<nil>"
	}
	return "{state:" + Truncated(pv.State.CompactString()) + " }"
}

func (pv WidgetStateProvider) ToJson() JSMap {
	m := NewJSMap()
	if pv.State != nil {
		m.Put("state", pv.State)
	}
	return m
}

func NewStateProvider(prefix string, state JSMap) WidgetStateProvider {
	Alert("!WidgetStateProvider is now just a JSMap; we will try to use the state stack if necessary to deal with prefixes")
	CheckArg(prefix == "")
	return &WidgetStateProviderStruct{State: state}
}
