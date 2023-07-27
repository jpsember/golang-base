package webserv

import (
	. "github.com/jpsember/golang-base/base"
	"html"
)

// Escaper interface performs html escaping on its argument
type Escaper interface {
	Escaped() string
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
