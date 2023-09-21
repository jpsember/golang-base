package webserv

import (
	. "github.com/jpsember/golang-base/base"
	"strings"
)

type MarkupBuilderObj struct {
	strings.Builder
	indent           int
	indented         bool
	crRequest        int
	omitComments     bool
	tagStack         []*tagEntry
	pendingComments  []any
	nested           bool
	currentStyleFlag bool
	pendingStyleFlag bool
	pendingQuotes    bool
	pendingEscape    bool
}

type tagEntry struct {
	tag        string // e.g. div, p (no '<' or '>')
	comment    string
	hasContent bool
}

type MarkupBuilder = *MarkupBuilderObj

func NewMarkupBuilder() MarkupBuilder {
	v := MarkupBuilderObj{}
	return &v
}

func (b MarkupBuilder) Bytes() []byte {
	return []byte(b.String())
}

func (b MarkupBuilder) DoIndent() MarkupBuilder {
	b.Cr()
	b.indent += 2
	return b
}

func (b MarkupBuilder) DoOutdent() MarkupBuilder {
	b.indent -= 2
	CheckState(b.indent >= 0, "indent underflow")
	b.Cr()
	return b
}

func (b MarkupBuilder) RenderInvisible(w Widget) MarkupBuilder {
	b.A(`<div id='`, w.Id(), `'></div>`)
	b.Cr()
	return b
}

func (b MarkupBuilder) Escape(arg any) MarkupBuilder {
	Todo("Use print effect to handle ESCAPE")
	if escaper, ok := arg.(Escaper); ok {
		return b.A(escaper.Escaped())
	}
	if str, ok := arg.(string); ok {
		return b.A(NewHtmlString(str).Escaped())
	}
	BadArg("<1Not escapable:", arg)
	return b
}

func (b MarkupBuilder) switchToMode(mode bool) {
	if mode != b.currentStyleFlag {
		if b.currentStyleFlag {
			b.WriteString(`" `)
		} else {
			b.WriteString(` style:"`)
		}
		b.currentStyleFlag = mode
	}
}

// Append markup, generating a linefeed if one is pending.  No escaping is performed.
func (b MarkupBuilder) A(args ...any) MarkupBuilder {
	if b.nested {
		BadState("nested")
	}
	b.nested = true

	b.updateStyleMode()

	for _, arg := range args {
		if b.crRequest != 0 {
			if b.crRequest == 1 {
				b.WriteString("\n")
			} else {
				b.WriteString("\n\n")
			}
			b.crRequest = 0
			b.doIndent()
		}

		switch v := arg.(type) {
		case string:
			b.appendStr(v)
		case int: // We aren't sure if it's 32 or 64, so choose 64
			b.appendStr(IntToString(v))
			break
		case bool:
			b.appendStr(boolToHtmlString(v))
		case PrintEffect:
			b.processPrintEffect(v)
		default:
			Die("<1Unsupported argument type:", Info(arg))
		}
	}
	b.nested = false
	return b
}

func (b MarkupBuilder) appendStr(text string) {
	if b.pendingQuotes {
		b.WriteByte('"')
		b.WriteString(text)
		b.WriteByte('"')
		b.pendingQuotes = false
	} else if b.pendingEscape {
		escaped := NewHtmlString(text).Escaped()
		b.WriteString(escaped)
		b.pendingEscape = false
	} else {
		b.WriteString(text)
	}
}

func (b MarkupBuilder) StyleOff() MarkupBuilder {
	b.pendingStyleFlag = false
	return b
}

func (b MarkupBuilder) Style(args ...any) MarkupBuilder {
	b.pendingStyleFlag = true
	b.A(args...)
	return b
}

func (b MarkupBuilder) updateStyleMode() {
	if b.pendingStyleFlag != b.currentStyleFlag {
		if b.pendingStyleFlag {
			b.WriteString(` style="`)
		} else {
			b.WriteString(`"`)
		}
		b.currentStyleFlag = b.pendingStyleFlag
	}

}
func (b MarkupBuilder) processPrintEffect(v PrintEffect) {
	switch v {
	case CR:
		b.Cr()
	case INDENT:
		b.DoIndent()
	case OUTDENT:
		b.DoOutdent()
	case QUOTED:
		b.pendingQuotes = true
	case ESCAPED:
		b.pendingEscape = true
	default:
		BadArg("Unsupported PrintEffect:", v)
	}
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
	b.A(sb.String(), ` -->`)
	b.Cr()
	return b
}

func (b MarkupBuilder) doIndent() {
	b.WriteString(SPACES[0:b.indent])
	b.indented = true
}

// Set pending comments for next TgOpen (or TgClose) call.
func (b MarkupBuilder) Comments(comments ...any) MarkupBuilder {
	if b.pendingComments != nil {
		Alert("#20<1Previous comments were not used:", b.pendingComments)
	}
	if !b.omitComments {
		b.pendingComments = comments
	}
	return b
}

func (b MarkupBuilder) TgOpen(name string) MarkupBuilder {
	// If there is a space, the user has added some attributes, e.g. `div xxxx="yyyy"...`;
	// treat this as if he did TgOpen(`div`).A(` xxxx....`)
	i := strings.IndexByte(name, ' ')
	tagName := name
	remainder := ""
	if i >= 0 {
		if i == 0 {
			BadArg("leading space in tag name:", Quoted(name))
		}
		tagName = name[0:i]
		remainder = name[i:]
	}

	entry := tagEntry{
		tag: tagName,
	}
	comments := b.pendingComments
	b.pendingComments = nil
	if comments != nil {
		entry.comment = `<!-- ` + ToString(comments...) + " -->"
	}
	if entry.comment != "" {
		b.Br()
		b.A(entry.comment).Cr()
	}
	b.A(`<`, tagName)

	if len(b.tagStack) >= 50 {
		BadState("tags are nested too deeply")
	}
	b.tagStack = append(b.tagStack, &entry)

	if remainder != "" {
		b.A(remainder)
	}
	return b
}

func (b MarkupBuilder) TgContent() MarkupBuilder {
	entry := Last(b.tagStack)
	CheckState(!entry.hasContent)
	entry.hasContent = true
	b.StyleOff()
	b.A(`>`)
	b.DoIndent()
	return b
}

func (b MarkupBuilder) TgClose() MarkupBuilder {
	var entry *tagEntry
	entry, b.tagStack = PopLast(b.tagStack)
	if entry.hasContent {
		b.DoOutdent()
		b.A("</", entry.tag, ">")
	} else {
		b.WriteString(` />`)
	}

	if entry.comment != "" {
		b.A(`  `, entry.comment)
	}

	return b.Br()
}

// Verify that the tag stack size *does not change* before and after some code.  Call this before the code,
// and balance this call with a call to VerifyEnd(), supplying the stack size that VerifyBegin() returned.
func (b MarkupBuilder) VerifyBegin() int {
	return len(b.tagStack)
}

// Verify that the tag stack size *does not change* before and after some code.  Call this before the code,
// and balance this call with a call to VerifyEnd(), supplying the stack size that VerifyBegin() returned.
func (b MarkupBuilder) VerifyEnd(expectedStackSize int, widget Widget) {
	s := len(b.tagStack)
	if s != expectedStackSize {
		BadState("<1tag stack size", s, "!=", expectedStackSize, INDENT,
			"after widget:", widget.Id(), Info(widget))
	}
}

func (b MarkupBuilder) Cr() MarkupBuilder {
	b.crRequest = MaxInt(1, b.crRequest)
	return b
}

func (b MarkupBuilder) Br() MarkupBuilder {
	b.crRequest = 2
	return b
}

func (b MarkupBuilder) SetComments(flag bool) {
	b.omitComments = !flag
}
