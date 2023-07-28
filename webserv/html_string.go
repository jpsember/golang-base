package webserv

import (
	. "github.com/jpsember/golang-base/base"
	"html"
)

type HtmlStringStruct struct {
	rawString        string
	escaped          string
	paragraphs       []string
	escapedGenerated bool
}

type HtmlString = *HtmlStringStruct

func NewHtmlString(rawString string) HtmlString {
	Todo("!What about text that might occur within quotes, e.g. in inputs?")
	h := HtmlStringStruct{
		rawString: rawString,
	}
	return &h
}

func NewHtmlStringEscaped(escapedString string) HtmlString {
	return &HtmlStringStruct{
		rawString:        escapedString,
		escaped:          escapedString,
		escapedGenerated: true,
	}
}

func (h HtmlString) String() string {
	return h.Escaped()
}

func (h HtmlString) parse() {
	if !h.escapedGenerated {
		h.escaped =
			html.EscapeString(h.rawString)
		h.escapedGenerated = true
	}
}

func (h HtmlString) Paragraphs() []string {
	if h.paragraphs == nil {
		h.paragraphs = stringToEscapedParagraphs(h.rawString)
	}
	return h.paragraphs
}

// Get the escaped form of the string.
func (h HtmlString) Escaped() string {
	h.parse()
	return h.escaped
}

func stringToEscapedParagraphs(markup string) []string {
	c := NewArray[string]()

	var currentPar []byte
	for i := 0; i < len(markup); i++ {
		ch := markup[i]
		if ch == '\n' {
			if currentPar != nil {
				s := string(currentPar)
				c.Add(html.EscapeString(s))
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
		c.Add(html.EscapeString(string(currentPar)))
	}
	if c.Size() == 0 {
		c.Add("")
	}
	return c.Array()
}
