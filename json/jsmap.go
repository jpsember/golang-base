package json

import (
	"fmt"
	//. "github.com/jpsember/golang-base/files"
	"sort"

	. "github.com/jpsember/golang-base/base"
)

type JSMapStruct struct {
	wrappedMap map[string]JSEntity
}
type JSMap = *JSMapStruct

//---------------------------------------------------------------------------
// JSEntity interface
//---------------------------------------------------------------------------

func (v JSMap) ToInteger() int64 {
	panic("Not supported")
}

func (v JSMap) ToFloat() float64 {
	panic("Not supported")
}

func (v JSMap) ToString() string {
	panic("The ToString() method is not supported for JSMapStruct")
}

func (v JSMap) ToBool() bool {
	panic("Not supported")
}

func (m JSMap) PrintTo(context *JSONPrinter) {
	var s = context.StringBuilder
	if context.Pretty {
		m.prettyPrintWithIndent(context)
	} else {
		s.WriteByte('{')
		var index = 0
		for key, val := range m.wrappedMap {
			if index != 0 {
				s.WriteByte(',')
			}
			index++
			s.WriteString(EscapedAndQuoted(key))
			s.WriteByte(':')
			val.PrintTo(context)
		}
		s.WriteByte('}')
	}
}

func (m JSMap) prettyPrintWithIndent(context *JSONPrinter) {
	var s = context.StringBuilder
	context.PushIndentAdjust(2)

	w := m.wrappedMap
	sortedKeysList := make([]string, 0, len(w))
	for k := range w {
		sortedKeysList = append(sortedKeysList, k)
	}
	sort.Strings(sortedKeysList)

	// Create a corresponding list of keys in escaped form, suitable for printing

	escapedKeysList := make([]string, 0, len(w))
	for _, k := range sortedKeysList {
		var k2 = Escaped(k)
		escapedKeysList = append(escapedKeysList, k2)
	}

	s.WriteString("{ ")
	var longestKeyLength = 0
	for _, str := range escapedKeysList {
		longestKeyLength = MaxInt(longestKeyLength, len(str))
	}

	for i, keyStr := range sortedKeysList {
		var escKeyStr = escapedKeysList[i]
		var effectiveKeyLength = MinInt(longestKeyLength, len(escKeyStr))
		var keyIndent = longestKeyLength
		if i != 0 {
			s.WriteString(",\n")
			keyIndent += context.Indent()
		}

		s.WriteString(Spaces(keyIndent - effectiveKeyLength))
		s.WriteString("\"")
		s.WriteString(escKeyStr)
		s.WriteString("\" : ")

		context.PushIndentAdjust(longestKeyLength + 5)

		// If the key we just printed was longer than our maximum,
		// print the value on the next line (indented to the appropriate column)
		if len(escKeyStr) > longestKeyLength {
			s.WriteString("\n")
			s.WriteString(Spaces(context.Indent()))
		}
		var value = w[keyStr]

		value.PrintTo(context)
		context.PopIndent()
	}
	context.PopIndent()
	if len(sortedKeysList) >= 2 {
		s.WriteString("\n")
		s.WriteString(Spaces(context.Indent()))
	} else {
		s.WriteByte(' ')
	}
	s.WriteByte('}')
}

//-----------------------------------------------------------------

// Factory constructor.  Do *not* construct via JSMapStruct().
func NewJSMap() *JSMapStruct {
	var m = new(JSMapStruct)
	m.wrappedMap = make(map[string]JSEntity)
	return m
}

// Implements the fmt.Stringer interface.  By default, we perform
// a pretty print of the JSMapStruct.  This simplifies a lot of things.
func (m JSMap) String() string {
	return PrintJSEntity(m, true)
}

// Convert JSListStruct to string, without pretty printing.
func (m JSMap) CompactString() string {
	return PrintJSEntity(m, false)
}

func (m JSMap) Put(key string, value any) *JSMapStruct {
	m.wrappedMap[key] = ToJSEntity(value)
	return m
}

func (m JSMap) PutNumbered(value any) *JSMapStruct {
	var numKeys = len(m.wrappedMap)
	var key = fmt.Sprintf("%3d", numKeys)
	return m.Put(key, value)
}

func (p *JSONParser) ParseMap() (*JSMapStruct, error) {
	p.adjustNest(1)
	var ourMap = make(map[string]JSEntity)
	p.ReadExpectedByte('{')
	commaExpected := false
	for !p.hasProblem() {
		if p.readIf('}') {
			break
		}
		if commaExpected {
			p.ReadExpectedByte(',')
			if p.readIf('}') {
				break
			}
			commaExpected = false
		}
		key := p.readString()
		p.ReadExpectedByte(':')
		ourMap[key] = p.readValue()
		commaExpected = true
	}
	p.adjustNest(-1)
	var jsMap *JSMapStruct
	if p.Error == nil {
		jsMap = NewJSMap()
		jsMap.wrappedMap = ourMap
	}
	return jsMap, p.Error
}

func (this *JSMapStruct) GetMap(key string) *JSMapStruct {
	var val = this.wrappedMap[key]
	return val.(*JSMapStruct)
}

func (this *JSMapStruct) OptMap(key string) *JSMapStruct {
	CheckNotNil(key, "nil key for OptMap")
	var val = this.wrappedMap[key]
	if val == nil {
		return nil
	}
	return val.(*JSMapStruct)
}

func (this *JSMapStruct) OptMapOrEmpty(key string) *JSMapStruct {
	var val = this.wrappedMap[key]
	if val == nil {
		return NewJSMap()
	}
	return val.(*JSMapStruct)
}

func (this *JSMapStruct) GetList(key string) *JSListStruct {
	var val = this.wrappedMap[key]
	return val.(*JSListStruct)
}

func (this *JSMapStruct) OptList(key string) *JSListStruct {
	var val = this.wrappedMap[key]
	if val == nil {
		return nil
	}
	return val.(*JSListStruct)
}

func (this *JSMapStruct) OptListOrEmpty(key string) *JSListStruct {
	var val = this.wrappedMap[key]
	if val == nil {
		return NewJSList()
	}
	return val.(*JSListStruct)
}

func (this *JSMapStruct) GetString(key string) string {
	var val = this.wrappedMap[key]
	return (val.(JSEntity)).ToString()
}

func (this *JSMapStruct) OptString(key string, defaultValue string) string {
	var val = this.wrappedMap[key]
	if val == nil {
		return defaultValue
	}
	return (val.(JSEntity)).ToString()
}

func (this *JSMapStruct) GetInt32(key string) int32 {
	var val = this.wrappedMap[key]
	return int32((val.(JSEntity)).ToInteger())
}

func (this *JSMapStruct) OptInt32(key string, defaultValue int32) int32 {
	var val = this.wrappedMap[key]
	if val == nil {
		return defaultValue
	}
	return int32((val.(JSEntity)).ToInteger())
}

func (this *JSMapStruct) OptFloat32(key string, defaultValue float32) float32 {
	var val = this.wrappedMap[key]
	if val == nil {
		return defaultValue
	}
	return float32((val.(JSEntity)).ToFloat())
}

func (this *JSMapStruct) OptFloat64(key string, defaultValue float64) float64 {
	var val = this.wrappedMap[key]
	if val == nil {
		return defaultValue
	}
	return float64((val.(JSEntity)).ToFloat())
}

func (this *JSMapStruct) GetInt64(key string) int64 {
	var val = this.wrappedMap[key]
	return int64((val.(JSEntity)).ToInteger())
}
func (this *JSMapStruct) OptInt64(key string, defaultValue int) int64 {
	var val = this.wrappedMap[key]
	if val == nil {
		return int64(defaultValue)
	}
	return int64((val.(JSEntity)).ToInteger())
}
func (this *JSMapStruct) GetBool(key string) bool {
	var val = this.wrappedMap[key]
	return (val.(JSEntity)).ToBool()
}

func (this *JSMapStruct) OptBool(key string, defaultValue bool) bool {
	var val = this.wrappedMap[key]
	if val == nil {
		return defaultValue
	}
	return (val.(JSEntity)).ToBool()
}

// If a key/value pair exists, return the value
func (jsmap *JSMapStruct) OptAny(key string) JSEntity {
	return jsmap.wrappedMap[key]
}

func JSMapFromString(content string) (*JSMapStruct, error) {
	var p JSONParser
	p.WithText(content)
	return p.ParseMap()
}

func JSMapFromStringM(content string) *JSMapStruct {
	var result, err = JSMapFromString(content)
	CheckOk(err)
	return result
}

func (jsmap *JSMapStruct) WrappedMap() map[string]JSEntity {
	return jsmap.wrappedMap
}
