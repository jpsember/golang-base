package json

import (
	"fmt"
	//. "github.com/jpsember/golang-base/files"
	"sort"

	. "github.com/jpsember/golang-base/base"
)

type JSMap struct {
	wrappedMap map[string]JSEntity
}

//---------------------------------------------------------------------------
// JSEntity interface
//---------------------------------------------------------------------------

func (v *JSMap) ToInteger() int64 {
	panic("Not supported")
}

func (v *JSMap) ToFloat() float64 {
	panic("Not supported")
}

func (v *JSMap) ToString() string {
	panic("Not supported")
}

func (v *JSMap) ToBool() bool {
	panic("Not supported")
}

func (m *JSMap) PrintTo(context *JSONPrinter) {
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

func (m *JSMap) prettyPrintWithIndent(context *JSONPrinter) {
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

// Factory constructor.  Do *not* construct via JSMap().
func NewJSMap() *JSMap {
	var m = new(JSMap)
	m.wrappedMap = make(map[string]JSEntity)
	return m
}

// Implements the fmt.Stringer interface.  By default, we perform
// a pretty print of the JSMap.  This simplifies a lot of things.
func (m *JSMap) String() string {
	return PrintJSEntity(m, true)
}

func (m *JSMap) Put(key string, value any) *JSMap {
	m.wrappedMap[key] = ToJSEntity(value)
	return m
}

func (m *JSMap) PutNumbered(value any) *JSMap {
	var numKeys = len(m.wrappedMap)
	var key = fmt.Sprintf("%3d", numKeys)
	return m.Put(key, value)
}

func (p *JSONParser) ParseMap() (*JSMap, error) {
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
	var jsMap *JSMap
	if p.Error == nil {
		jsMap = NewJSMap()
		jsMap.wrappedMap = ourMap
	}
	return jsMap, p.Error
}

func (this *JSMap) GetMap(key string) *JSMap {
	var val = this.wrappedMap[key]
	return val.(*JSMap)
}

func (this *JSMap) OptMap(key string) *JSMap {
	var val = this.wrappedMap[key]
	if val == nil {
		return nil
	}
	return val.(*JSMap)
}

func (this *JSMap) OptMapOrEmpty(key string) *JSMap {
	var val = this.wrappedMap[key]
	if val == nil {
		return NewJSMap()
	}
	return val.(*JSMap)
}

func (this *JSMap) GetList(key string) *JSList {
	var val = this.wrappedMap[key]
	return val.(*JSList)
}

func (this *JSMap) OptList(key string) *JSList {
	var val = this.wrappedMap[key]
	if val == nil {
		return nil
	}
	return val.(*JSList)
}

func (this *JSMap) OptListOrEmpty(key string) *JSList {
	var val = this.wrappedMap[key]
	if val == nil {
		return NewJSList()
	}
	return val.(*JSList)
}

func (this *JSMap) GetString(key string) string {
	var val = this.wrappedMap[key]
	return (val.(JSEntity)).ToString()
}

func (this *JSMap) OptString(key string, defaultValue string) string {
	var val = this.wrappedMap[key]
	if val == nil {
		return defaultValue
	}
	return (val.(JSEntity)).ToString()
}

func (this *JSMap) GetInt32(key string) int32 {
	var val = this.wrappedMap[key]
	return int32((val.(JSEntity)).ToInteger())
}

func (this *JSMap) OptInt32(key string, defaultValue int32) int32 {
	var val = this.wrappedMap[key]
	if val == nil {
		return defaultValue
	}
	return int32((val.(JSEntity)).ToInteger())
}

func (this *JSMap) OptFloat32(key string, defaultValue float32) float32 {
	var val = this.wrappedMap[key]
	if val == nil {
		return defaultValue
	}
	return float32((val.(JSEntity)).ToFloat())
}

func (this *JSMap) OptFloat64(key string, defaultValue float64) float64 {
	var val = this.wrappedMap[key]
	if val == nil {
		return defaultValue
	}
	return float64((val.(JSEntity)).ToFloat())
}

func (this *JSMap) GetInt64(key string) int64 {
	var val = this.wrappedMap[key]
	return int64((val.(JSEntity)).ToInteger())
}
func (this *JSMap) OptInt64(key string, defaultValue int) int64 {
	var val = this.wrappedMap[key]
	if val == nil {
		return int64(defaultValue)
	}
	return int64((val.(JSEntity)).ToInteger())
}
func (this *JSMap) GetBool(key string) bool {
	var val = this.wrappedMap[key]
	return (val.(JSEntity)).ToBool()
}

func (this *JSMap) OptBool(key string, defaultValue bool) bool {
	var val = this.wrappedMap[key]
	if val == nil {
		return defaultValue
	}
	return (val.(JSEntity)).ToBool()
}

// If a key/value pair exists, return the value
func (jsmap *JSMap) OptAny(key string) JSEntity {
	return jsmap.wrappedMap[key]
}

func JSMapFromString(content string) (*JSMap, error) {
	var p JSONParser
	p.WithText(content)
	return p.ParseMap()
}

func JSMapFromStringM(content string) *JSMap {
	var result, err = JSMapFromString(content)
	CheckOk(err)
	return result
}
