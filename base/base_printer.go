// Constructs formatted strings, using a fluent interface
// Adapted from Java
package base

import (
	"fmt"
	"strings"
)

// Print arguments to standard output, using a BasePrinter.
func Pr(messages ...any) {
	fmt.Println(ToString(messages...))
}

// Use a BasePrinter to format a string from an array of objects.
func ToString(message ...any) string {
	var x BasePrinter
	return x.Pr(message...).String()
}

type BasePrinter struct {
	indentColumn  int
	column        int
	maxColumn     int
	pendingBreak  int
	contentBuffer strings.Builder
}

// Get contents as a string.
func (this *BasePrinter) String() string {
	return this.contentBuffer.String()
}

// Clear any existing contents.
func (this *BasePrinter) Clear() *BasePrinter {
	this.contentBuffer.Reset()
	return this
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
func (this *BasePrinter) Cr() *BasePrinter {
	this.pendingBreak = MaxInt(this.pendingBreak, brkLine)
	return this
}

// Append a number of spaces.
func (this *BasePrinter) appendSpaces(count int) {
	var cursor = 0
	for cursor < count {
		var maxSpacesPerRow = 42
		var netSpaces = MinInt(count-cursor, maxSpacesPerRow)
		this.appendCharacters(Spaces(netSpaces))
		cursor += netSpaces
	}
}

// Insert a paragraph break between current content and subsequent content.
//
// A paragraph is marked by two consecutive linefeeds.
func (this *BasePrinter) Br() *BasePrinter {
	this.pendingBreak = brkParagraph
	return this
}

// Request a space before any subsequent non-whitespace.
func (this *BasePrinter) Sp() *BasePrinter {
	if this.column != 0 {
		this.pendingBreak = MaxInt(this.pendingBreak, brkColumn)
	}
	return this
}

// ------------------------------------------------------------------
// Indentation
// ------------------------------------------------------------------

const (
	defaultIndentationColumn = 4
)

// Increase the indent amount to the next tab stop (4 spaces), and generate a linefeed
func (this *BasePrinter) Indent() *BasePrinter {
	this.indentColumn += defaultIndentationColumn
	return this.Cr()
}

// Move the indent amount to the previous tab stop, and generate a linefeed
func (this *BasePrinter) Outdent() *BasePrinter {
	CheckState(this.indentColumn != 0)
	this.indentColumn -= defaultIndentationColumn
	return this.Cr()
}

// Clear variables concerning indentation, pending line breaks
func (this *BasePrinter) ResetIndentation() *BasePrinter {
	this.pendingBreak = 0
	this.indentColumn = 0
	this.column = 0
	return this
}

// ------------------------------------------------------------------
// Appending content
// ------------------------------------------------------------------

// Append string representations of objects, separated by spaces
func (this *BasePrinter) Pr(messages ...any) *BasePrinter {
	for _, obj := range messages {
		this.Sp()
		this.append(obj)
	}
	return this
}

// Append an object's string representation
// (todo: by looking up its converter)
func (this *BasePrinter) append(value any) {
	switch v := value.(type) {
	case nil:
		this.AppendString("<nil>")
	case string:
		this.AppendString(v)
	case int: // We aren't sure if it's 32 or 64, so choose 64
		this.AppendLong(int64(v))
	case int32:
		this.AppendInt(v)
	case uint32:
		this.AppendInt(int32(v))
	case int64:
		this.AppendLong(v)
	case uint64:
		this.AppendLong(int64(v))
	case uint8:
		this.AppendInt(int32(v))
	case int8:
		this.AppendInt(int32(v))
	case uint16:
		this.AppendInt(int32(v))
	case int16:
		this.AppendInt(int32(v))
	case float32:
		this.AppendFloat(float64(v))
	case float64:
		this.AppendFloat(v)
	case bool:
		this.AppendBool(v)
	default:
		{
			// // See if handler exists for this type of value
			// var argType = reflect.TypeOf(value)

			// var handler, exists = handlerMap[argType]
			// if exists {
			// 	var str = handler(value)
			// 	this.AppendString(str)
			// 	return
			// }

			// this should be handled by the fmt.sprintf() call below...
			//
			// // If value implements the HasStringMethod interface, call its String()
			// // method to determine what to append.
			// //
			// stackoverflow.com/questions/27803654/explanation-of-checking-if-value-implements-interface
			// hasStringMethod, ok := value.(fmt.Stringer)
			// if ok {
			// 	this.AppendString(hasStringMethod.String())
			// 	return
			// }
		}

		// Todo("What is it doing here to test the equality?  Is it comparing pointers, or something more elaborate?")
		if value == INDENT {
			this.Indent()
			return
		}
		if value == OUTDENT {
			this.Outdent()
			return
		}
		if value == CR {
			this.Cr()
			return
		}

		if true {
			// Fall back on using fmt.Sprint()
			this.AppendString(fmt.Sprint(v))
		} else {
			this.AppendString("<??? " + Info(v) + " ???>")
		}
	}
}

// Append string; split into separate strings where linefeeds exist, and request linefeeds
// to continue the next string with the appropriate indenting.
func (this *BasePrinter) AppendString(str string) *BasePrinter {

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
			this.flushLineBreak()
			if this.pendingBreak == brkColumn {
				if this.column > 0 {
					if this.contentBuffer.Len() > 0 && !strings.HasSuffix(this.contentBuffer.String(), " ") {
						this.appendCharacters(" ")
					}
				}
				this.pendingBreak = 0
			}
			if this.column == 0 {
				this.appendSpaces(this.indentColumn)
			}
			this.appendCharacters(prefix)
		}

		if nextCrLocation == remainingLength {
			break
		}

		this.Cr()
		str = str[1+nextCrLocation:]
	}
	return this
}

// Append string representation of a bool; T or F.
func (this *BasePrinter) AppendBool(b bool) *BasePrinter {
	var t string
	if b {
		t = "T"
	} else {
		t = "F"
	}
	return this.AppendString(t)
}

// Append a floating point value, fixed width, without scientific notation.
func (this *BasePrinter) AppendFloat(dblVal float64) *BasePrinter {
	var formattedValue = fmt.Sprintf("%v ", dblVal)
	var allZerosSuffix = ".0000 "
	var newVal = strings.TrimSuffix(formattedValue, allZerosSuffix)
	if newVal != formattedValue {
		formattedValue = newVal + "      "
	}
	return this.AppendString(formattedValue)
}

func (this *BasePrinter) AppendInt(intVal int32) *BasePrinter {
	this.formatLong(int64(intVal), 6)
	return this
}

func (this *BasePrinter) AppendLong(longVal int64) *BasePrinter {
	this.formatLong(longVal, 8)
	return this
}

// Adjust tab stop to a particular column.
func (this *BasePrinter) tab(column int) *BasePrinter {
	this.flushLineBreak()
	var spacesToNextTabStop = column - this.column
	if spacesToNextTabStop > 0 {
		this.appendSpaces(spacesToNextTabStop)
	}
	return this
}

func (this *BasePrinter) AppendList(value []any) {
	if value == nil {
		this.AppendString("<nil>")
		return
	}
	this.AppendString("[")
	for i, item := range value {
		if i != 0 {
			this.AppendString(",")
		}
		this.append(item)
	}
	this.AppendString("]")
}

func (this *BasePrinter) appendCharacters(characters string) {
	this.contentBuffer.WriteString(characters)
	this.column += len(characters)
	this.maxColumn = MaxInt(this.maxColumn, this.column)
}

func (this *BasePrinter) flushLineBreak() {
	if this.pendingBreak < brkLine {
		return
	}
	this.contentBuffer.WriteByte('\n')
	if this.pendingBreak >= brkParagraph {
		this.contentBuffer.WriteByte('\n')
	}
	this.pendingBreak = 0
	this.column = 0
}

// Append long integer, padding to particular fixed width
func (this *BasePrinter) formatLong(longVal int64, fixedWidth int) {
	var digits = fmt.Sprintf("%d", AbsLong(longVal))
	var paddingChars = fixedWidth - len(digits)
	var spaces = Spaces(MaxInt(1, paddingChars))
	var signChar string
	if longVal < 0 {
		signChar = "-"
	} else {
		signChar = " "
	}
	this.AppendString(spaces + signChar + digits)
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
var DASH = makeEffect(1)
var BR = makeEffect(2)
var INDENT = makeEffect(3)
var OUTDENT = makeEffect(4)
