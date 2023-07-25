package webserv

import (
	. "github.com/jpsember/golang-base/base"
	"strings"
)

// A builder for constructing html markup

type MarkupBuilderObj struct {
	strings.Builder
	indent       int
	indented     bool
	crRequest    int
	omitComments bool
}

type MarkupBuilder = *MarkupBuilderObj

func NewMarkupBuilder() MarkupBuilder {
	v := MarkupBuilderObj{}
	return &v
}

func (m MarkupBuilder) DoIndent() MarkupBuilder {
	m.Cr()
	m.indent += 2
	return m
}

func (m MarkupBuilder) DoOutdent() MarkupBuilder {
	m.indent -= 2
	CheckState(m.indent >= 0, "indent underflow")
	m.Cr()
	return m
}

func (m MarkupBuilder) DebugOpen(widget Widget) MarkupBuilder {
	m.Cr()
	m.A(`<div class="card border border-primary shadow-0 mb-3"><div class="card-body">`)
	m.DoIndent()
	return m
}

func (m MarkupBuilder) DebugClose() MarkupBuilder {
	m.DoOutdent()
	m.A(`</div></div>`)
	m.Cr()
	return m
}

func (m MarkupBuilder) DebugOpenSpan(widget Widget) MarkupBuilder {
	m.A(`<span class="border">`)
	m.A("span border open")
	return m
}

func (m MarkupBuilder) DebugCloseSpan() MarkupBuilder {
	m.A("span border close")
	m.A(`</span>`)
	return m
}

func (m MarkupBuilder) RenderInvisible(w Widget, tag string) MarkupBuilder {
	m.A(`<`)
	m.A(tag)
	m.A(` id='`)
	m.A(w.GetId())
	m.A(`'></`)
	m.A(tag)
	m.A(`>`)
	m.Cr()
	return m
}

func (m MarkupBuilder) Quoted(text string) MarkupBuilder {
	return m.A(Quoted(text))
}

func (m MarkupBuilder) A(text string) MarkupBuilder {
	if m.crRequest != 0 {
		if m.crRequest == 1 {
			m.WriteString("\n")
		} else {
			m.WriteString("\n\n")
		}
		m.crRequest = 0
		m.doIndent()
	}
	m.WriteString(text)
	return m
}

func (m MarkupBuilder) Pr(message ...any) MarkupBuilder {
	m.A(ToString(message...))
	return m
}

func (b MarkupBuilder) HtmlComment(messages ...any) MarkupBuilder {
	b.A(`<!-- `)
	b.Pr(messages...)
	b.A(` -->`)
	b.Cr()
	return b
}

func (b MarkupBuilder) doIndent() {
	b.WriteString(SPACES[0:b.indent])
	b.indented = true
}

func (b MarkupBuilder) OpenHtml(tag string, comment string) MarkupBuilder {
	CheckState(b.indent < 100, "too many indents")
	comment = b.commentFilter(comment)
	b.A("<")
	b.A(tag)
	b.A(">")
	if comment != "" {
		b.A(" <!--")
		b.A(comment)
		b.A(" -->")
	}
	b.DoIndent()
	return b
}

func (b MarkupBuilder) CloseHtml(tag string, comment string) MarkupBuilder {
	b.DoOutdent()
	b.A("</")
	b.A(tag)
	comment = b.commentFilter(comment)
	if comment != "" {
		b.A("> <!-- ")
		b.A(comment)
		b.A(" -->")
	} else {
		b.A(">")
	}
	return b.Cr()
}

func (b MarkupBuilder) Cr() MarkupBuilder {
	b.crRequest = MaxInt(1, b.crRequest)
	return b
}

func (b MarkupBuilder) Br() MarkupBuilder {
	b.crRequest = 2
	return b
}

func (b MarkupBuilder) Comments(flag bool) {
	b.omitComments = !flag
}

func (b MarkupBuilder) commentFilter(comment string) string {
	if b.omitComments {
		return ""
	}
	return comment
}

func WrapWithinComment(text string) string {
	CheckArg(!strings.HasPrefix(text, "<"))
	return "<!-- " + text + " -->"
}
