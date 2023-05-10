package json

import (
	"fmt"
	. "js/base"
	"strconv"
)

// ---------------------------------------------------------------------------------------
// Interface for all json values (JSMap, JSList, JString, JInteger, ...)
// ---------------------------------------------------------------------------------------
type JSEntity interface {

	// Print value to the context (with its pretty printing, indentation parameters).
	PrintTo(context *JSONPrinter)

	// Get value of an integer.
	ToInteger() int64

	// Get value of a number as a float (supported for integers as well).
	ToFloat() float64

	// Get value of a string.
	ToString() string

	// Get value of a bool.
	ToBool() bool
}

// This is very relevant:
// https://blog.riff.org/2014_07_19_golang_fun_adding_methods_to_primitive_types
//
// "Now, Go has the ability to assign types to other types...
//
// ...the two types will be different when compared, but offer the same capabilities...
//
// ...There we have it : we added a method on a type which is a copy of a primitive
// type. No struct boxing, no unboxing in the method."

// -------------------------------------------------------------------------------
// Json type: string
type JString string

// We call this Make... instead of New..., to reflect the fact that we are not returning
// a pointer to a value, rather the value itself.
func MakeJString(value string) JSEntity {
	//Todo("We might just want to do 'JString(value)' instead of a separate function (x number of types)")
	return JString(value)
}

func (v JString) PrintTo(context *JSONPrinter) {
	context.WriteString(EscapedAndQuoted(string(v)))
}

func (v JString) ToInteger() int64 {
	panic("Not supported")
}

func (v JString) ToFloat() float64 {
	panic("Not supported")
}

func (v JString) ToString() string {
	return string(v)
}

func (v JString) ToBool() bool {
	panic("Not supported")
}

// -------------------------------------------------------------------------------
// Json type: number (integral)
type JInteger int64

func MakeJInteger(value int64) JSEntity {
	return JInteger(value)
}

func (v JInteger) PrintTo(context *JSONPrinter) {
	context.WriteString(strconv.FormatInt(int64(v), 10))
}

func (v JInteger) ToInteger() int64 {
	return int64(v)
}

func (v JInteger) ToFloat() float64 {
	return float64(v)
}

func (v JInteger) ToString() string {
	panic("Not supported")
}
func (v JInteger) ToBool() bool {
	panic("Not supported")
}

// -------------------------------------------------------------------------------
// Json type: number (floating point)
type JFloat float64

func MakeJFloat(value float64) JSEntity {
	return JFloat(value)
}

func (v JFloat) PrintTo(context *JSONPrinter) {
	// We could print fewer fractional digits by e.g. %.3f
	var text = fmt.Sprintf("%f", float64(v))
	context.WriteString(text)
}

func (v JFloat) ToInteger() int64 {
	return int64(v)
}

func (v JFloat) ToFloat() float64 {
	return float64(v)
}

func (v JFloat) ToString() string {
	panic("Not supported")
}
func (v JFloat) ToBool() bool {
	panic("Not supported")
}

// -------------------------------------------------------------------------------
// Json type: boolean

type JBool bool

func MakeJBool(value bool) JSEntity {
	return JBool(value)
}

func (v JBool) PrintTo(context *JSONPrinter) {
	var text string
	if v {
		text = "true"
	} else {
		text = "false"
	}
	context.WriteString(text)
}

func (v JBool) ToInteger() int64 {
	panic("Not supported")
}

func (v JBool) ToFloat() float64 {
	panic("Not supported")
}

func (v JBool) ToString() string {
	panic("Not supported")
}
func (v JBool) ToBool() bool {
	return bool(v)
}

// -------------------------------------------------------------------------------
// Json type: null

type JNull int

var JNullValue = JNull(0)

func (v JNull) PrintTo(context *JSONPrinter) {
	context.WriteString("null")
}

func (v JNull) ToInteger() int64 {
	panic("Not supported")
}

func (v JNull) ToFloat() float64 {
	panic("Not supported")
}

func (v JNull) ToString() string {
	panic("Not supported")
}

func (v JNull) ToBool() bool {
	panic("Not supported")
}

// -------------------------------------------------------------------------------

func EscapedAndQuoted(str string) string {
	return Quoted(Escaped(str))
}

func Escaped(str string) string {
	var out []byte
	var ESCAPE = byte('\\')
	for _, c := range str {
		switch c {
		case '"', '\\':
			out = append(out, ESCAPE)
		case 8:
			out, c = append(out, ESCAPE), 'b'
		case 12:
			out, c = append(out, ESCAPE), 'f'
		case 10:
			out, c = append(out, ESCAPE), 'n'
		case 13:
			out, c = append(out, ESCAPE), 'r'
		case 9:
			out, c = append(out, ESCAPE), 't'
		default:
			if c < ' ' || c > 126 {
				out = append(out, ESCAPE)
				out = append(out, 'u')
				out, c = toHex(out, int(c), 4), 0
			}
		}
		if c != 0 {
			out = append(out, byte(c))
		}
	}
	return string(out)
}

// Convert value to hex, store in target, return target.
func toHex(target []byte, value int, digits int) []byte {

	for digits > 0 {
		digits--
		var shift = digits << 2
		var v = (value >> shift) & 0xf
		var c int
		if v < 10 {
			c = ('0' + v)
		} else {
			c = 'a' + (v - 10)
		}
		target = append(target, byte(c))
	}
	return target
}

// Convert a value to an appropriate JSEntity.
func ToJSEntity(value any) JSEntity {
	var val JSEntity
	switch v := value.(type) {
	case int:
		val = MakeJInteger(int64(v))
	case int32:
		val = MakeJInteger(int64(v))
	case int64:
		val = MakeJInteger(v)
	case uint:
		val = MakeJInteger(int64(v))
	case float32:
		val = MakeJFloat(float64(v))
	case float64:
		val = MakeJFloat(v)
	case string:
		val = MakeJString(v)
	case bool:
		val = MakeJBool(v)
	case JSEntity: // Already a JSEntity, i.e., a JSMap or JSList?
		val = v
	default:
		Die("Unsupported:", Info(value))
	}
	return val
}
