// Constructs formatted strings, using a fluent interface
// Adapted from Java
package base

import (
	"fmt"
	"reflect"
	"strings"
)

// Print arguments to standard output, using a BasePrinter.
func Pr(messages ...any) {
	fmt.Println(ToString(messages...))
}

func PrIf(active bool) func(messages ...any) {
	if active {
		Alert("<1Printing is active")
		return Pr
	}
	return PrNull
}

// A printer that generates no output
func PrNull(messages ...any) {

}

// Use a BasePrinter to format a string from an array of objects.
func ToString(message ...any) string {
	var x BasePrinter
	return x.Pr(message...).String()
}

func NewBasePrinter() *BasePrinter {
	var b = new(BasePrinter)
	return b
}

type BasePrinter struct {
	indentColumn  int
	column        int
	maxColumn     int
	pendingBreak  int
	contentBuffer strings.Builder
}

// Get contents as a string.
func (b *BasePrinter) String() string {
	return b.contentBuffer.String()
}

// Clear any existing contents.
func (b *BasePrinter) Clear() *BasePrinter {
	b.contentBuffer.Reset()
	return b
}

// ------------------------------------------------------------------
// Spaces, linefeeds, paragraph breaks
// ------------------------------------------------------------------

const (
	brkColumn = iota + 1
	brkLine
	brkParagraph
)

// Request a linefeed before any subsequent non-whitespace is printed.
func (b *BasePrinter) Cr() *BasePrinter {
	b.pendingBreak = MaxInt(b.pendingBreak, brkLine)
	return b
}

// Append a number of spaces.
func (b *BasePrinter) appendSpaces(count int) {
	var cursor = 0
	for cursor < count {
		var maxSpacesPerRow = 42
		var netSpaces = MinInt(count-cursor, maxSpacesPerRow)
		b.appendCharacters(Spaces(netSpaces))
		cursor += netSpaces
	}
}

// Insert a paragraph break between current content and subsequent content.
//
// A paragraph is marked by two consecutive linefeeds.
func (b *BasePrinter) Br() *BasePrinter {
	b.pendingBreak = brkParagraph
	return b
}

// Request a space before any subsequent non-whitespace.
func (b *BasePrinter) Sp() *BasePrinter {
	if b.column != 0 {
		b.pendingBreak = MaxInt(b.pendingBreak, brkColumn)
	}
	return b
}

// ------------------------------------------------------------------
// Indentation
// ------------------------------------------------------------------

const (
	defaultIndentationColumn = 4
)

// Increase the indent amount to the next tab stop (4 spaces), and generate a linefeed
func (b *BasePrinter) Indent() *BasePrinter {
	b.indentColumn += defaultIndentationColumn
	if b.column > 0 {
		b.Cr()
	}
	return b
}

// Move the indent amount to the previous tab stop, and generate a linefeed
func (b *BasePrinter) Outdent() *BasePrinter {
	CheckState(b.indentColumn != 0)
	b.indentColumn -= defaultIndentationColumn
	return b.Cr()
}

// Clear variables concerning indentation, pending line breaks
func (b *BasePrinter) ResetIndentation() *BasePrinter {
	b.pendingBreak = 0
	b.indentColumn = 0
	b.column = 0
	return b
}

// ------------------------------------------------------------------
// Appending content
// ------------------------------------------------------------------

// Append string representations of objects, separated by spaces
func (b *BasePrinter) Pr(messages ...any) *BasePrinter {
	for _, obj := range messages {
		b.Sp()
		b.Append(obj)
	}
	return b
}

// Append an object's string representation
// (todo: by looking up its converter)
func (b *BasePrinter) Append(value any) {
	switch v := value.(type) {
	case nil:
		b.AppendString("<nil>")
	case string:
		b.AppendString(v)
	case int: // We aren't sure if it's 32 or 64, so choose 64
		b.AppendLong(int64(v))
	case int32:
		b.AppendInt(v)
	case uint32:
		b.AppendInt(int32(v))
	case int64:
		b.AppendLong(v)
	case uint64:
		b.AppendLong(int64(v))
	case uint8:
		b.AppendInt(int32(v))
	case int8:
		b.AppendInt(int32(v))
	case uint16:
		b.AppendInt(int32(v))
	case int16:
		b.AppendInt(int32(v))
	case float32:
		b.AppendFloat(float64(v))
	case float64:
		b.AppendFloat(v)
	case bool:
		b.AppendBool(v)
	case strings.Builder:
		b.AppendString(v.String())
	case PrintEffect:
		processPrintEffect(v, b)
	default:
		q := reflect.TypeOf(v)
		// Fall back on using fmt.Sprint()
		if false {
			b.AppendString("???" + q.String() + "???" + fmt.Sprint(v))
		} else {
			b.AppendString(fmt.Sprint(v))
		}
	}
}

// Append string; split into separate strings where linefeeds exist, and request linefeeds
// to continue the next string with the appropriate indenting.
func (b *BasePrinter) AppendString(str string) *BasePrinter {

	// Scan forward through the string, processing maximal prefixes that don't include linefeeds,
	// requesting linefeeds as we find them, and stopping when we run out of such prefixes.
	for {
		var remainingLength = len(str)
		if remainingLength == 0 {
			break
		}

		// Grab the largest slice that doesn't contain a linefeed
		var nextCrLocation = strings.IndexByte(str, '\n')
		var prefix string
		if nextCrLocation < 0 {
			nextCrLocation = remainingLength
			prefix = str
		} else {
			prefix = str[:nextCrLocation]
		}

		if len(prefix) > 0 {
			b.flushLineBreak()
			if b.pendingBreak == brkColumn {
				if b.column > 0 {
					if b.contentBuffer.Len() > 0 && !strings.HasSuffix(b.contentBuffer.String(), " ") {
						b.appendCharacters(" ")
					}
				}
				b.pendingBreak = 0
			}
			if b.column == 0 {
				b.appendSpaces(b.indentColumn)
			}
			b.appendCharacters(prefix)
		}

		if nextCrLocation == remainingLength {
			break
		}

		b.Cr()
		str = str[1+nextCrLocation:]
	}
	return b
}

// Append string representation of a bool; T or F.
func (b *BasePrinter) AppendBool(v bool) *BasePrinter {
	var t string
	if v {
		t = "T"
	} else {
		t = "F"
	}
	return b.AppendString(t)
}

// Append a floating point value, fixed width, without scientific notation.
func (b *BasePrinter) AppendFloat(dblVal float64) *BasePrinter {
	var formattedValue = fmt.Sprintf("%v ", dblVal)
	var allZerosSuffix = ".0000 "
	var newVal = strings.TrimSuffix(formattedValue, allZerosSuffix)
	if newVal != formattedValue {
		formattedValue = newVal + "      "
	}
	return b.AppendString(formattedValue)
}

func (b *BasePrinter) AppendInt(intVal int32) *BasePrinter {
	b.formatLong(int64(intVal), 6)
	return b
}

func (b *BasePrinter) AppendLong(longVal int64) *BasePrinter {
	b.formatLong(longVal, 8)
	return b
}

// Adjust tab stop to a particular column.
func (b *BasePrinter) tab(column int) *BasePrinter {
	b.flushLineBreak()
	var spacesToNextTabStop = column - b.column
	if spacesToNextTabStop > 0 {
		b.appendSpaces(spacesToNextTabStop)
	}
	return b
}

func (b *BasePrinter) AppendList(value []any) {
	if value == nil {
		b.AppendString("<nil>")
		return
	}
	b.AppendString("[")
	for i, item := range value {
		if i != 0 {
			b.AppendString(",")
		}
		b.Append(item)
	}
	b.AppendString("]")
}

func (b *BasePrinter) appendCharacters(characters string) {
	b.contentBuffer.WriteString(characters)
	b.column += len(characters)
	b.maxColumn = MaxInt(b.maxColumn, b.column)
}

func (b *BasePrinter) flushLineBreak() {
	if b.pendingBreak < brkLine {
		return
	}
	b.contentBuffer.WriteByte('\n')
	if b.pendingBreak >= brkParagraph {
		b.contentBuffer.WriteByte('\n')
	}
	b.pendingBreak = 0
	b.column = 0
}

// Append long integer, padding to particular fixed width
func (b *BasePrinter) formatLong(longVal int64, fixedWidth int) {
	var digits = fmt.Sprintf("%d", AbsLong(longVal))
	var paddingChars = fixedWidth - len(digits)
	var spaces = Spaces(MaxInt(1, paddingChars))
	var signChar string
	if longVal < 0 {
		signChar = "-"
	} else {
		signChar = " "
	}
	b.AppendString(spaces + signChar + digits)
}

// // ---------------------------------------------------------------------------------------
// // Plugging in handlers for client types
// // ---------------------------------------------------------------------------------------

// // A function that clients can register with the BasePrinter class(?) to
// // allow printer to handle types it doesn't know about at compile time.
// //
// // The 'any' argument should have the type that the handler is prepared
// // to convert to a string.
// type BasePrintableFunc func(any) string

// // Register a handler such that BasePrinters will call the handler to
// // convert values of the same type as value to a string to be printed.
// func RegisterBasePrinterType(value any, handler BasePrintableFunc) {
// 	valueType := reflect.TypeOf(value)
// 	_, exists := handlerMap[valueType]
// 	CheckState(!exists, "handler already exists")
// 	handlerMap[valueType] = handler
// }

// var handlerMap = make(map[any]BasePrintableFunc)

// ---------------------------------------------------------------------------------------
// Singleton objects that have effects when included in a BasePrinter.pr() argument list
// ---------------------------------------------------------------------------------------

type PrintEffect struct {
	value int
}

func makeEffect(n int) PrintEffect {
	var x = PrintEffect{value: n}
	return x
}

var CR = makeEffect(0)
var DASHES = makeEffect(1)
var BR = makeEffect(2)
var INDENT = makeEffect(3)
var OUTDENT = makeEffect(4)
var VERT_SP = makeEffect(5)
var RESET = makeEffect(6)
var QUOTED = makeEffect(7)

func processPrintEffect(v PrintEffect, b *BasePrinter) {
	switch v {
	case CR:
		b.Cr()
	case INDENT:
		b.Indent()
	case OUTDENT:
		b.Outdent()
	case DASHES:
		b.Cr()
		b.AppendString(Dashes)
	case VERT_SP:
		b.Cr()
		b.contentBuffer.WriteString("\n\n\n\n")
	case RESET:
		b.ResetIndentation()
	}
}
