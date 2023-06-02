package webserv

import (
	. "github.com/jpsember/golang-base/base"
	"html"
)

type HtmlStringStruct struct {
	Source           string
	escaped          string
	escapedGenerated bool
}

type HtmlString = *HtmlStringStruct

func EscapedHtml(markup string) HtmlString {
	h := HtmlStringStruct{
		Source: markup,
	}

	return &h
}

func (h HtmlString) String() string {
	if !h.escapedGenerated {
		h.escaped = html.EscapeString(h.Source)
		Pr("orig:", h.Source)
		Pr("escaped:", h.escaped)
		h.escapedGenerated = true
	}
	return h.escaped
}
