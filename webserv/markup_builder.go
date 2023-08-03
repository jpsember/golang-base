package webserv

import (
	. "github.com/jpsember/golang-base/base"
	"strings"
)

// A builder for constructing html markup

type tagEntry struct {
	tag       string // e.g. div, p (no '<' or '>')
	comment   string
	noContent bool // true if tag will have no content; its formatting and commenting will be different
}

type MarkupBuilderObj struct {
	strings.Builder
	indent                     int
	indented                   bool
	crRequest                  int
	omitComments               bool
	tagStack                   *Array[tagEntry]
	suppressClosingCommentFlag bool
	pendingComments            []any
}

type MarkupBuilder = *MarkupBuilderObj

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
	base := w.GetBaseWidget()
	b.A(`<div id='`, base.Id, `'></div>`)
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

// Append markup, generating a linefeed if one is pending.  No escaping is performed.
func (b MarkupBuilder) A(args ...any) MarkupBuilder {
	if b.crRequest != 0 {
		if b.crRequest == 1 {
			b.WriteString("\n")
		} else {
			b.WriteString("\n\n")
		}
		b.crRequest = 0
		b.doIndent()
	}
	for _, arg := range args {

		s := ""
		switch v := arg.(type) {
		case string:
			s = v
		case int: // We aren't sure if it's 32 or 64, so choose 64
			s = IntToString(v)
		//case int32:
		//	b.AppendInt(v)
		//case uint32:
		//	b.AppendInt(int32(v))
		//case int64:
		//	b.AppendLong(v)
		//case uint64:
		//	b.AppendLong(int64(v))
		//case uint8:
		//	b.AppendInt(int32(v))
		//case int8:
		//	b.AppendInt(int32(v))
		//case uint16:
		//	b.AppendInt(int32(v))
		//case int16:
		//	b.AppendInt(int32(v))
		//case float32:
		//	b.AppendFloat(float64(v))
		//case float64:
		//	b.AppendFloat(v)
		case bool:
			s = boolToHtmlString(v)
		default:
			Die("<1Unsupported argument type:", Info(arg))
		}
		b.WriteString(s)
	}
	return b
}

func (b MarkupBuilder) Pr(message ...any) MarkupBuilder {
	b.A(ToString(message...))
	return b
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
func (b MarkupBuilder) OpenTag(tagExpression string) MarkupBuilder {
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
		tag:       tagExpression[0:i],
		noContent: b.suppressClosingCommentFlag,
	}
	comments := b.pendingComments
	b.pendingComments = nil

	b.suppressClosingCommentFlag = false
	if comments != nil {
		entry.comment = `<!-- ` + ToString(comments...) + " -->"
	}
	if entry.comment != "" {
		b.Br()
		b.A(entry.comment).Cr()
	}
	b.tagStack.Add(entry)

	b.A("<", tagExpression, ">")
	if !entry.noContent {
		b.DoIndent()
	}
	return b
}

func (b MarkupBuilder) tagStackInfo() string {
	jl := NewJSList()

	for _, ent := range b.tagStack.Array() {
		jl.Add(NewJSList().Add(ent.tag).Add(ent.comment))
	}
	return jl.String()
}

func (b MarkupBuilder) VerifyBegin() int {
	return b.tagStack.Size()
}
func (b MarkupBuilder) VerifyEnd(expectedStackSize int) {
	s := b.tagStack.Size()
	if s != expectedStackSize {
		BadState("tag stack size", s, "!=", expectedStackSize, "; content:", b.tagStackInfo())
	}
}

func (b MarkupBuilder) CloseTag() MarkupBuilder {
	entry := b.tagStack.Pop()
	if entry.noContent {
		b.A("</", entry.tag, ">")
	} else {
		b.DoOutdent()
		b.A("</", entry.tag, ">")
		if entry.comment != "" {
			b.A(`  `, entry.comment)
		}
	}
	return b.Br()
}

func (b MarkupBuilder) OpenCloseTag(tagExpression string) MarkupBuilder {
	b.suppressClosingCommentFlag = true
	b.OpenTag(tagExpression)
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
