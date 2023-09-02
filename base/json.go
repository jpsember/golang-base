package base

import (
	"encoding/base64"
	"fmt"
	"strconv"
)

// ---------------------------------------------------------------------------------------
// Interface for all json values (JSMapStruct, JSListStruct, JString, JInteger, ...)
// ---------------------------------------------------------------------------------------
type JSEntity interface {

	// Print value to the context (with its pretty printing, indentation parameters).
	PrintTo(context *JSONPrinter)

	// Get value of an integer.
	AsInteger() int64

	// Get value of a number as a float (supported for integers as well).
	AsFloat() float64

	// Get value of a string.
	AsString() string

	// Get value of a bool.
	AsBool() bool

	AsJSMap() JSMap

	AsJSList() JSList
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

func (v JString) AsInteger() int64 {
	panic("Not supported")
}

func (v JString) AsFloat() float64 {
	panic("Not supported")
}

func (v JString) AsString() string {
	return string(v)
}

func (v JString) AsBool() bool {
	panic("Not supported")
}

func (v JString) AsJSMap() JSMap {
	panic("Not supported")
}

func (v JString) AsJSList() JSList {
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

func (v JInteger) AsInteger() int64 {
	return int64(v)
}

func (v JInteger) AsFloat() float64 {
	return float64(v)
}

func (v JInteger) AsString() string {
	panic("Not supported")
}
func (v JInteger) AsBool() bool {
	panic("Not supported")
}

func (v JInteger) AsJSMap() JSMap {
	panic("Not supported")
}

func (v JInteger) AsJSList() JSList {
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

func (v JFloat) AsInteger() int64 {
	return int64(v)
}

func (v JFloat) AsFloat() float64 {
	return float64(v)
}

func (v JFloat) AsString() string {
	panic("Not supported")
}
func (v JFloat) AsBool() bool {
	panic("Not supported")
}

func (v JFloat) AsJSMap() JSMap {
	panic("Not supported")
}

func (v JFloat) AsJSList() JSList {
	panic("Not supported")
}

// -------------------------------------------------------------------------------
// Json type: boolean

type JBool bool

const (
	JBoolFalse = JBool(false)
	JBoolTrue  = JBool(true)
)

func MakeJBool(value bool) JSEntity {
	if value {
		return JBoolTrue
	}
	return JBoolFalse
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

func (v JBool) AsInteger() int64 {
	panic("Not supported")
}

func (v JBool) AsFloat() float64 {
	panic("Not supported")
}

func (v JBool) AsString() string {
	panic("Not supported")
}
func (v JBool) AsBool() bool {
	return bool(v)
}

func (v JBool) AsJSMap() JSMap {
	panic("Not supported")
}

func (v JBool) AsJSList() JSList {
	panic("Not supported")
}

// -------------------------------------------------------------------------------
// Json type: null

type JNull int

var JNullValue = JNull(0)

func (v JNull) PrintTo(context *JSONPrinter) {
	context.WriteString("null")
}

func (v JNull) AsInteger() int64 {
	panic("Not supported")
}

func (v JNull) AsFloat() float64 {
	panic("Not supported")
}

func (v JNull) AsString() string {
	panic("Not supported")
}

func (v JNull) AsBool() bool {
	panic("Not supported")
}

func (v JNull) AsJSMap() JSMap {
	panic("Not supported")
}

func (v JNull) AsJSList() JSList {
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
			c = '0' + v
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
	case uint:
		val = MakeJInteger(int64(v))
	case int:
		val = MakeJInteger(int64(v))
	case uint8:
		val = MakeJInteger(int64(v))
	case int8:
		val = MakeJInteger(int64(v))
	case int32:
		val = MakeJInteger(int64(v))
	case uint32:
		val = MakeJInteger(int64(v))
	case int64:
		val = MakeJInteger(v)
	case uint64:
		val = MakeJInteger(int64(v))
	case float32:
		val = MakeJFloat(float64(v))
	case float64:
		val = MakeJFloat(v)
	case string:
		val = MakeJString(v)
	case bool:
		val = MakeJBool(v)
	case JSEntity: // Already a JSEntity, i.e., a JSMapStruct or JSListStruct?
		val = v
	case DataClass:
		val = v.ToJson().(JSEntity)
	default:
		Die("Unsupported:", Info(value))
	}
	return val
}

const DATA_TYPE_DELIMITER = "`"
const DATA_TYPE_SUFFIX_BYTE = DATA_TYPE_DELIMITER + "b"

//const DATA_TYPE_SUFFIX_SHORT = DATA_TYPE_DELIMITER + "s"
//const DATA_TYPE_SUFFIX_INT = DATA_TYPE_DELIMITER + "i"
//const DATA_TYPE_SUFFIX_LONG = DATA_TYPE_DELIMITER + "l"
//const DATA_TYPE_SUFFIX_FLOAT = DATA_TYPE_DELIMITER + "f"
//const DATA_TYPE_SUFFIX_DOUBLE = DATA_TYPE_DELIMITER + "d"

func removeDataTypeSuffix(s string, optionalSuffix string) string {
	if len(s) >= 2 {
		suffixStart := len(s) - 2
		if s[suffixStart] == '`' {
			existingSuffixChar := s[suffixStart+1]
			if existingSuffixChar != optionalSuffix[1] {
				BadArg("string has suffix", existingSuffixChar, "expected", optionalSuffix[1])
			}
			s = s[0:suffixStart]
		}
	}
	return s
}

// Encode a byte array as a Base64 string, with our data type suffix added
func EncodeBase64(byteArray []byte) string {
	Alert("this is using the wrong encoding!")
	return base64.URLEncoding.EncodeToString(byteArray) + DATA_TYPE_SUFFIX_BYTE
}

// Encode a byte array as a Base64 string if it is fairly long
func EncodeBase64Maybe(byteArray []byte) JSEntity {
	if len(byteArray) > 8 {
		return JString(EncodeBase64(byteArray))
	}
	return JSListWith(byteArray)
}

func ParseBase64(s string) []byte {
	s = removeDataTypeSuffix(s, DATA_TYPE_SUFFIX_BYTE)
	return CheckOkWith(base64.StdEncoding.DecodeString(s))
}

/**
 * Parse an array of bytes from a value that is either a JSList, or a base64
 * string. This is so we are prepared to read it whether or not it has been
 * stored in a space-saving base64 form.
 */
func DecodeBase64Maybe(ent JSEntity) []byte {
	if arr, ok := ent.(JString); ok {
		return ParseBase64(arr.AsString())
	}
	if arr, ok := ent.(JSList); ok {
		return arr.AsByteArray()
	}
	BadArg("unexpected type:", Info(ent))
	return nil
}
