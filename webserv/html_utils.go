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

func EscapedHtmlIntoParagraphs(markup string) []HtmlString {
	c := NewArray[HtmlString]()

	var currentPar []byte
	for i := 0; i < len(markup); i++ {
		ch := markup[i]
		if ch == '\n' {
			if currentPar != nil {
				s := string(currentPar)
				c.Add(EscapedHtml(s))
				currentPar = nil
			}
		} else {
			if currentPar == nil {
				currentPar = make([]byte, 0)
			}
			currentPar = append(currentPar, ch)
		}
	}
	if currentPar != nil {
		c.Add(EscapedHtml(string(currentPar)))
	}
	return c.Array()
}

func (h HtmlString) String() string {
	if !h.escapedGenerated {
		h.escaped = html.EscapeString(h.Source)
		h.escapedGenerated = true
	}
	return h.escaped
}
