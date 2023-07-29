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

func (m MarkupBuilder) RenderInvisible(w Widget) MarkupBuilder {
	b := w.GetBaseWidget()
	m.A(`<div id='`)
	m.A(b.Id)
	m.A(`'></div>`)
	m.Cr()
	return m
}

func (m MarkupBuilder) Quoted(text string) MarkupBuilder {
	return m.A(Quoted(text))
}

func (m MarkupBuilder) Escape(arg any) MarkupBuilder {
	if escaper, ok := arg.(Escaper); ok {
		return m.A(escaper.Escaped())
	}
	if str, ok := arg.(string); ok {
		return m.A(NewHtmlString(str).Escaped())
	}
	BadArg("<1Not escapable:", arg)
	return m
}

// Append markup, generating a linefeed if one is pending.  No escaping is performed.
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

// Append an HTML comment.
func (b MarkupBuilder) Comment(messages ...any) MarkupBuilder {
	b.A(`<!-- `)
	content := ToString(messages...)
	// Look for embedded "-->" substrings within the comment, and escape them so the text doesn't
	// prematurely close the comment.
	const token = `-->`
	const tokenLen = len(token)
	substr := content
	sb := strings.Builder{}
	for {
		i := strings.Index(substr, token)
		if i < 0 {
			break
		}
		sb.WriteString(substr[:i])
		sb.WriteString(`--\>`)
		substr = substr[i+tokenLen:]
	}
	sb.WriteString(substr)
	b.A(sb.String())
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
