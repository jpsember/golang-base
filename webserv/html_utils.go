package webserv

import (
	. "github.com/jpsember/golang-base/base"
	"html"
	"strings"
)

type HtmlStringStruct struct {
	Source           string
	escaped          string
	escapedGenerated bool
}

type HtmlString = *HtmlStringStruct

func EscapedHtml(markup string) HtmlString {
	h := HtmlStringStruct{
		Source: markup,
	}

	return &h
}

func (h HtmlString) String() string {
	if !h.escapedGenerated {
		h.escaped = html.EscapeString(h.Source)
		Pr("orig:", h.Source)
		Pr("escaped:", h.escaped)
		h.escapedGenerated = true
	}
	return h.escaped
}

type JavascriptStringStruct struct {
	Source           string
	escaped          string
	escapedGenerated bool
}

type JavascriptString = *JavascriptStringStruct

func EscapedJavascript(m string) JavascriptString {
	h := JavascriptStringStruct{
		Source: m,
	}

	return &h
}

func (h JavascriptString) String() string {
	if !h.escapedGenerated {
		h.escaped = EscapeJavascriptChars(h.Source, nil, false).String()
		Pr("orig:", h.Source)
		Pr("escaped:", h.escaped)
		h.escapedGenerated = true
	}
	return h.escaped
}

func EscapeJavascriptChars(sourceSequence string, sb *strings.Builder, withQuotes bool) *strings.Builder {
	if sb == nil {
		sb = &strings.Builder{}
	}
	if withQuotes {
		sb.WriteByte('"')
	}

	for _, c := range sourceSequence {
		const ESCAPE = '\\'
		switch c {
		case ESCAPE:
			sb.WriteByte(ESCAPE)
		case 8:
			sb.WriteByte(ESCAPE)
			c = 'b'
		case 12:
			sb.WriteByte(ESCAPE)
			c = 'f'
		case 10:
			sb.WriteByte(ESCAPE)
			c = 'n'
		case 13:
			sb.WriteByte(ESCAPE)
			c = 'r'
		case 9:
			sb.WriteByte(ESCAPE)
			c = 't'
		default:
			// Remove the '|| c > 126' to leave text as unicode
			if c < ' ' || c > 126 {
				sb.WriteString("\\u")
				ToHex(sb, int(c), 4)
				continue
			}
		}
		sb.WriteByte(byte(c))
	}

	if withQuotes {
		sb.WriteByte('"')
	}
	return sb
}

func ToHex(stringBuilder *strings.Builder, value int, digits int) {
	for digits > 0 {
		digits--
		shift := digits << 2
		v := (value >> shift) & 0xf
		var c byte
		if v < 10 {
			c = byte('0' + v)
		} else {
			c = 'a' + byte(v-10)
		}
		stringBuilder.WriteByte(c)
	}
}
