package webserv

import (
	. "github.com/jpsember/golang-base/base"
)

type HtmlStringStruct struct {
	rawString        string
	escaped          []string
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
	h := HtmlStringStruct{
		rawString:        escapedString,
		escaped:          []string{escapedString},
		escapedGenerated: true,
	}
	return &h
}

func (h HtmlString) String() string {
	return "HtmlString, source:" + Quoted(h.rawString)
}

func (h HtmlString) parse() {
	if !h.escapedGenerated {
		Todo("this paragraph stuff is going to trouble")
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

// Get the expected single escaped html paragraph.
// If there are no paragraphs, return an empty string.
func (h HtmlString) Escaped() string {
	p := h.Paragraphs()
	var result string
	switch len(p) {
	default:
		BadArg("<1 Expected a single escaped paragraph from:", Quoted(h.rawString), "but got:", JSListWith(h.escaped))
	case 1:
		result = p[0]
	case 0:
	}
	return result
}
