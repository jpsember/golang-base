package webserv

import (
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
		h.escapedGenerated = true
	}
	return h.escaped
}
