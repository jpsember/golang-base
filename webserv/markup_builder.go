package webserv

import (
	. "github.com/jpsember/golang-base/base"
	"strings"
)

// A builder for constructing htlm markup

type MarkupBuilderObj struct {
	strings.Builder
	indent int
}

type MarkupBuilder = *MarkupBuilderObj

func NewMarkupBuilder() MarkupBuilder {
	v := MarkupBuilderObj{}
	return &v
}

func (m MarkupBuilder) A(text string) MarkupBuilder {
	m.WriteString(text)
	return m
}

func (m MarkupBuilder) Pr(message ...any) MarkupBuilder {
	m.WriteString(ToString(message...))
	return m
}

func (b MarkupBuilder) HtmlComment(messages ...any) MarkupBuilder {
	b.A(`<!-- `)
	b.Pr(messages...)
	b.A("-->\n")
	return b
}

func (b MarkupBuilder) OpenHtml(tag string, comment string) MarkupBuilder {
	CheckState(b.indent < 100, "too many indents")
	if comment != "" {
		b.A("<!--")
		b.A(comment)
		b.A(" --> ")
	}
	b.A("<")
	b.A(tag)
	b.A(">")
	return b
}

func (b MarkupBuilder) CloseHtml(tag string, comment string) MarkupBuilder {
	b.A("</")
	b.A(tag)
	if comment != "" {
		b.A("> <!--")
		b.A(comment)
		b.A(" -->\n")
	} else {
		b.A(">\n")
	}
	return b
}

func (b MarkupBuilder) Cr() MarkupBuilder {
	return b.A("\n")
}

func WrapWithinComment(text string) string {
	CheckArg(!strings.HasPrefix(text, "<"))
	return "<!-- " + text + " -->"
}
