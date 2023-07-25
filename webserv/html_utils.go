package webserv

import (
	. "github.com/jpsember/golang-base/base"
	"html"
)

type HtmlStringStruct struct {
	rawString        string
	escaped          []string
	escapedGenerated bool
}

type HtmlString = *HtmlStringStruct

func NewHtmlString(rawString string) HtmlString {
	h := HtmlStringStruct{
		rawString: rawString,
	}
	return &h
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
	return c.Array()
}

func (h HtmlString) String() string {
	return "HtmlString, source:" + Quoted(h.rawString)
}

func (h HtmlString) parse() {
	if !h.escapedGenerated {
		h.escaped = stringToEscapedParagraphs(h.rawString)
		h.escapedGenerated = true
	}
}

func (h HtmlString) ParagraphCount() int {
	h.parse()
	return len(h.escaped)
}

func (h HtmlString) Paragraph(index int) string {
	h.parse()
	return h.escaped[index]
}

func (h HtmlString) Paragraphs() []string {
	h.parse()
	return h.escaped
}
