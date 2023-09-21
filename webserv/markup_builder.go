package webserv

import (
	. "github.com/jpsember/golang-base/base"
	"strings"
)

// A builder for constructing html markup

type tagEntry struct {
	openType   tagOpenType
	tag        string // e.g. div, p (no '<' or '>')
	comment    string
	hasContent bool
}

const (
	mode_html = iota
	mode_style
)

type MarkupBuilderObj struct {
	strings.Builder
	indent          int
	indented        bool
	crRequest       int
	omitComments    bool
	tagStack        *Array[tagEntry]
	pendingComments []any
	nested          bool
	currentMode     int
	pendingMode     int
}

type MarkupBuilder = *MarkupBuilderObj

type tagOpenType int

const (
	tagTypeOpen tagOpenType = iota
	tagTypeOpenClose
	tagTypeVoid
)

func NewMarkupBuilder() MarkupBuilder {
	v := MarkupBuilderObj{}
	v.tagStack = NewArray[tagEntry]()
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

func (b MarkupBuilder) Quoted(text string) MarkupBuilder {
	return b.A(Quoted(text))
}

func (b MarkupBuilder) Escape(arg any) MarkupBuilder {
	if escaper, ok := arg.(Escaper); ok {
		return b.A(escaper.Escaped())
	}
	if str, ok := arg.(string); ok {
		return b.A(NewHtmlString(str).Escaped())
	}
	BadArg("<1Not escapable:", arg)
	return b
}

func (b MarkupBuilder) switchToMode(mode int) {
	if mode != b.currentMode {
		if b.currentMode == mode_style {
			b.WriteString(`" `)
		} else {
			b.WriteString(` style:"`)
		}
		b.currentMode = mode
	}
}

// Append markup, generating a linefeed if one is pending.  No escaping is performed.
func (b MarkupBuilder) A(args ...any) MarkupBuilder {
	if b.nested {
		BadState("nested")
	}
	b.nested = true

	b.updateMode()

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
			b.WriteString(v)
		case int: // We aren't sure if it's 32 or 64, so choose 64
			b.WriteString(IntToString(v))
			break
		case bool:
			b.WriteString(boolToHtmlString(v))
		case PrintEffect:
			b.processPrintEffect(v)
		default:
			Die("<1Unsupported argument type:", Info(arg))
		}
	}
	b.nested = false
	return b
}

func (b MarkupBuilder) StyleOn() MarkupBuilder {
	b.pendingMode = mode_style
	return b
}

func (b MarkupBuilder) StyleOff() MarkupBuilder {
	b.pendingMode = mode_html
	return b
}

func (b MarkupBuilder) updateMode() {
	if b.pendingMode != b.currentMode {
		if b.pendingMode == mode_style {
			b.WriteString(` style="`)
		} else {
			b.WriteString(`"`)
		}
		b.currentMode = b.pendingMode
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

// Set pending comments for next OpenTag (or OpenCloseTag) call.
func (b MarkupBuilder) Comments(comments ...any) MarkupBuilder {
	if b.pendingComments != nil {
		Alert("#20<1Previous comments were not used:", b.pendingComments)
	}
	if !b.omitComments {
		b.pendingComments = comments
	}
	return b
}

// Open a tag, e.g.
//
//	<div class="card-body" style="max-height:8em;">
//
// tagExpression in the above case would be:  div class="card-body" style="max-height:8em;"
// Deprecated.
func (b MarkupBuilder) OpenTag(args ...any) MarkupBuilder {
	b.auxOpenTag(tagTypeOpen, args...)
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

	CheckState(b.tagStack.Size() < 50, "tags are nested too deeply")
	b.tagStack.Add(entry)

	if remainder != "" {
		b.A(remainder)
	}
	return b
}

func (b MarkupBuilder) TgContent() MarkupBuilder {
	// We must point to the entry, not copy it, as we are modifying it
	entry := &b.tagStack.Array()[b.tagStack.Size()-1]
	CheckState(!entry.hasContent)
	entry.hasContent = true
	if b.pendingMode != mode_html {
		Alert("#50<1missing StyleOff")
	}
	b.StyleOff()
	b.A(`>`)
	b.DoIndent()
	return b
}

func (b MarkupBuilder) TgClose() MarkupBuilder {
	entry := b.tagStack.Pop()
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

// Deprecated.  Use TgOpen, TgContent, TgClose functions.
func (b MarkupBuilder) auxOpenTag(openType tagOpenType, args ...any) {
	b.updateMode()

	var tagExpression string
	{
		sb := strings.Builder{}

		for _, arg := range args {
			s := ""
			switch v := arg.(type) {
			case string:
				s = v
			case int: // We aren't sure if it's 32 or 64, so choose 64
				s = IntToString(v)
			case bool:
				s = boolToHtmlString(v)
			default:
				Die("<1Unsupported argument type:", Info(arg))
			}
			sb.WriteString(s)
		}
		tagExpression = sb.String()
	}
	Todo("!In debug mode, parse the tag expression to make sure quotes are balanced")

	exprLen := len(tagExpression)
	if tagExpression[0] == '<' || tagExpression[exprLen-1] == '>' {
		BadArg("<1Tag expression contains <,> delimiters:", tagExpression)
	}
	i := strings.IndexByte(tagExpression, ' ')
	if i < 0 {
		i = exprLen
	}

	CheckState(b.tagStack.Size() < 50, "tags are nested too deeply")
	entry := tagEntry{
		tag:      tagExpression[0:i],
		openType: openType,
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

	b.A("<", tagExpression, ">")
	if openType == tagTypeOpen {
		b.DoIndent()
	}
	if openType != tagTypeVoid {
		b.tagStack.Add(entry)
	}
}

func (b MarkupBuilder) tagStackInfo() string {
	jl := NewJSList()

	for _, ent := range b.tagStack.Array() {
		jl.Add(NewJSList().Add(ent.tag).Add(ent.comment))
	}
	return jl.String()
}

// Verify that the tag stack size *does not change* before and after some code.  Call this before the code,
// and balance this call with a call to VerifyEnd(), supplying the stack size that VerifyBegin() returned.
func (b MarkupBuilder) VerifyBegin() int {
	return b.tagStack.Size()
}

// Verify that the tag stack size *does not change* before and after some code.  Call this before the code,
// and balance this call with a call to VerifyEnd(), supplying the stack size that VerifyBegin() returned.
func (b MarkupBuilder) VerifyEnd(expectedStackSize int, widget Widget) {
	s := b.tagStack.Size()
	if s != expectedStackSize {
		BadState("<1tag stack size", s, "!=", expectedStackSize, INDENT,
			"after widget:", widget.Id(), Info(widget))
	}
}

// Deprecated.  Use TgOpen, TgContent, TgClose functions.
func (b MarkupBuilder) CloseTag() MarkupBuilder {
	if b.tagStack.IsEmpty() {
		Die("tag stack is empty:", INDENT, b.String())
	}
	entry := b.tagStack.Pop()
	if entry.openType == tagTypeOpen {
		b.DoOutdent()
		b.A("</", entry.tag, ">")
		if entry.comment != "" {
			b.A(`  `, entry.comment)
		}
	} else {
		b.A("</", entry.tag, ">")
	}
	return b.Br()
}

// Deprecated.  Use TgOpen, TgContent, TgClose functions.
func (b MarkupBuilder) VoidTag(args ...any) MarkupBuilder {
	b.auxOpenTag(tagTypeVoid, args...)
	return b
}

// Deprecated.  Use TgOpen, TgContent, TgClose functions.
func (b MarkupBuilder) OpenCloseTag(args ...any) MarkupBuilder {
	b.auxOpenTag(tagTypeOpenClose, args...)
	return b.CloseTag()
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
