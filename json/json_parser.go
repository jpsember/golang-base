package json

import (
	"fmt"
	. "github.com/jpsember/golang-base/base"
	"strconv"
	"strings"
)

type JSONParser struct {
	textBytes []byte
	cursor    int
	Error     error
}

type ParseError struct {
	prob    string
	Context string
	Cursor  int
}

func (e *ParseError) Error() string {
	return e.prob
}

type JSKeyword struct {
	text  string
	bytes []byte
	value JSEntity
}

func buildKeyword(keyword string, value JSEntity) JSKeyword {
	return JSKeyword{
		text:  keyword,
		bytes: []byte(keyword),
		value: value,
	}
}

var JSNull = buildKeyword("null", JNullValue)
var JSTrue = buildKeyword("true", MakeJBool(true))
var JSFalse = buildKeyword("false", MakeJBool(false))

func (p *JSONParser) WithText(text string) *JSONParser {
	CheckNotNil(text)
	p.textBytes = []byte(text)
	p.cursor = 0
	// Let's always have the cursor sitting at non-whitespace
	p.skipWhitespace()
	// For fluid interface support, return the JSONParser pointer.
	return p
}

func (p *JSONParser) hasProblem() bool {
	return p.Error != nil
}

func (p *JSONParser) fail(message ...any) {

	if p.Error != nil {
		return
	}

	var txt = "problem parsing json"
	if len(message) != 0 {
		txt += "; " + ToString(message...)
	}

	sb := strings.Builder{}
	sb.WriteString("...")
	{
		i := MaxInt(p.cursor-15, 0)
		if i < p.cursor {
			sb.WriteString(string(p.textBytes[i:p.cursor]))
		}
	}
	sb.WriteString("!")

	{
		i := MinInt(p.cursor+15, len(p.textBytes))
		if i > p.cursor {
		}
		sb.WriteString(string(p.textBytes[p.cursor:i]))
	}
	var context = sb.String()
	var msg = ToString(JoinLists([]any{
		fmt.Sprintf("Problem parsing json, cursor: %v,", p.cursor), "context:",
		context}, message)...)
	p.Error = &ParseError{prob: msg, Context: context, Cursor: p.cursor}
}

func (p *JSONParser) skipWhitespace() bool {

	var mSourceChars, j = p.textBytes, p.cursor
	var length = len(mSourceChars)

	for j < length {
		var c = mSourceChars[j]
		if c == '/' {
			j++
			if j == length || mSourceChars[j] != '/' {
				p.fail()
				return false
			}
			j++
			for j < length && mSourceChars[j] != '\n' {
				j++
			}
			continue
		}
		if c > ' ' {
			p.cursor = j
			return true
		}
		j++
	}
	p.cursor = j
	return false
}

// Read an expected character and any following whitespace.
func (p *JSONParser) ReadExpectedByte(expected byte) {
	if p.read() != expected {
		p.fail("expected '" + string(expected) + "'")
	} else {
		p.skipWhitespace()
	}
}

// If next character matches a value, read it and any following whitespace, and return true.
func (p *JSONParser) readIf(expected byte) bool {
	if p.peek() == expected {
		p.cursor++
		p.skipWhitespace()
		return true
	}
	return false
}

// Read a quoted, escaped string and any following whitespace.
func (p *JSONParser) readString() string {

	// Todo("probably doesn't deal with utf-8 properly?")
	var w strings.Builder

	p.ReadExpectedByte('"')

	for !p.hasProblem() {
		var c = p.read()
		if c == '"' {
			break
		}
		if c != '\\' {
			w.WriteByte(c)
			continue
		}
		c = p.read()
		switch c {
		case '\\', '"', '/':
			w.WriteByte(c)
		case 'b':
			w.WriteByte('\b')
		case 'f':
			w.WriteByte('\f')
		case 'n':
			w.WriteByte('\n')
		case 'r':
			w.WriteByte('\r')
		case 't':
			w.WriteByte('\t')
		case 'u':
			p.fail("Unicode not yet supported")
			//     w.append((char) ((readHex() << 12) | (readHex() << 8) | (readHex() << 4) | readHex()));
		default:
			p.fail()
		}
	}
	p.skipWhitespace()
	return w.String()
}

func (p *JSONParser) assertCompleted() {
	if p.cursor != len(p.textBytes) {
		p.fail("excess characters")
	}
}

func (p *JSONParser) peek() byte {
	if p.cursor >= len(p.textBytes) {
		p.fail("reached end of input")
		return 0
	}
	return p.textBytes[p.cursor]
}

func (p *JSONParser) read() byte {
	var result = p.peek()
	p.cursor++
	return result
}

func (p *JSONParser) readHexint() int {
	result := 0
	var c = int(p.read())
	if c >= 'a' {
		c -= ('a' - 'A')
	}
	if c >= 'A' {
		result = c - ('A') + 10
		if result >= 16 {
			p.fail()
		}
		return result
	} else {
		result = c - '0'
		if result < 0 || result >= 10 {
			p.fail()
		}
	}
	return result
}

func (p *JSONParser) readNumber() JSEntity {

	var start = p.cursor
	var isFloat = false

	var text, length = p.textBytes, len(p.textBytes)
	for p.cursor < length {
		var c = p.peek()
		if c <= ' ' || c == ',' || c == ']' || c == '}' {
			break
		}
		if c == 'e' || c == 'E' || c == '.' {
			isFloat = true
		}
		p.cursor++
	}
	var expr = string(text[start:p.cursor])

	p.skipWhitespace()

	var value JSEntity

	if isFloat {
		v, err := strconv.ParseFloat(expr, 64)
		if err == nil {
			value = MakeJFloat(v)
		}
	} else {
		v, err := strconv.Atoi(expr)
		if err == nil {
			value = MakeJInteger(int64(v))
		}
	}
	if value == nil {
		p.fail("problem parsing number", expr)
		value = JInteger(0)
	}
	return value
}

func (p *JSONParser) ReadExpectedBytes(s []byte) {
	if len(s)+p.cursor > len(p.textBytes) {
		p.fail("end of data reading expected bytes")
	}

	for i, c := range s {
		if c != p.textBytes[p.cursor+i] {
			p.fail()
		}
	}
	p.cursor += len(s)

	p.skipWhitespace()
}

func (p *JSONParser) readValue() JSEntity {
	// Set result to something, in case we get an error before
	// we can assign something
	var result JSEntity
	result = JBoolFalse
	if !p.hasProblem() {
		var ch = p.peek()
		switch ch {
		case '[':
			result = p.ParseList()
		case '{':
			result = p.ParseMap()
		case '"':
			result = MakeJString(p.readString())
		case 't':
			result = MakeJBool(p.readTrue())
		case 'f':
			result = MakeJBool(p.readFalse())
		case 'n':
			p.read()
			p.ReadExpectedBytes(JSNull.bytes)
			result = JNullValue
		default:
			result = p.readNumber()
		}
		p.skipWhitespace()
		if result == nil {
			Die("failed to parse value, ch:", ch)
		}
	}
	return result
}

func (p *JSONParser) readTrue() bool {
	p.ReadExpectedBytes(JSTrue.bytes)
	return true
}

func (p *JSONParser) readFalse() bool {
	p.ReadExpectedBytes(JSFalse.bytes)
	return false
}

func (p *JSONParser) readKeywordValue(kword *JSKeyword) any {
	p.ReadExpectedBytes(kword.bytes)
	return kword.value
}
