package base

import (
	"fmt"
)

type JSMapStruct struct {
	wrappedMap map[string]JSEntity
}
type JSMap = *JSMapStruct

//---------------------------------------------------------------------------
// JSEntity interface
//---------------------------------------------------------------------------

func (m JSMap) AsInteger() int64 {
	panic("Not supported")
}

func (m JSMap) AsFloat() float64 {
	panic("Not supported")
}

func (m JSMap) AsString() string {
	panic("The AsString() method is not supported for JSMapStruct")
}

func (m JSMap) AsBool() bool {
	panic("Not supported")
}

func (m JSMap) AsJSMap() JSMap {
	return m
}

func (m JSMap) AsJSList() JSList {
	panic("Not supported")
}

func (m JSMap) PrintTo(context *JSONPrinter) {
	var s = context.StringBuilder
	if context.Pretty {
		m.prettyPrintWithIndent(context)
	} else {
		entries := m.Entries()
		s.WriteByte('{')
		for index, entry := range entries {
			if index != 0 {
				s.WriteByte(',')
			}
			index++
			s.WriteString(EscapedAndQuoted(entry.Key))
			s.WriteByte(':')
			entry.Value.PrintTo(context)
		}
		s.WriteByte('}')
	}
}

func (m JSMap) prettyPrintWithIndent(context *JSONPrinter) {
	context.PushIndentAdjust(2)
	w := m.wrappedMap
	entries := m.Entries()

	// Create a corresponding list of keys in escaped form, suitable for printing

	escapedKeysList := make([]string, 0, len(w))
	for _, entry := range entries {
		escapedKeysList = append(escapedKeysList, Escaped(entry.Key))
	}

	var s = context.StringBuilder
	s.WriteString("{ ")

	var longestKeyLength = 0
	for _, str := range escapedKeysList {
		longestKeyLength = MaxInt(longestKeyLength, len(str))
	}

	for i, entry := range entries {
		var value = entry.Value
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

		value.PrintTo(context)
		context.PopIndent()
	}
	context.PopIndent()
	if len(entries) >= 2 {
		s.WriteString("\n")
		s.WriteString(Spaces(context.Indent()))
	} else {
		s.WriteByte(' ')
	}
	s.WriteByte('}')
}

//-----------------------------------------------------------------

// Factory constructor.  Do *not* construct via JSMapStruct().
func NewJSMap() JSMap {
	var m = new(JSMapStruct)
	return m.Clear()
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

func (m JSMap) Delete(key string) JSMap {
	delete(m.wrappedMap, key)
	return m
}

func (m JSMap) Put(key string, value any) JSMap {
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

func (m JSMap) GetMap(key string) *JSMapStruct {
	var val = m.wrappedMap[key]
	return val.(*JSMapStruct)
}

func (m JSMap) OptMap(key string) *JSMapStruct {
	CheckNotNil(key, "nil key for OptMap")
	var val = m.wrappedMap[key]
	if val == nil {
		return nil
	}
	return val.(*JSMapStruct)
}

func (m JSMap) OptMapOrEmpty(key string) *JSMapStruct {
	var val = m.wrappedMap[key]
	if val == nil {
		return NewJSMap()
	}
	return val.(JSMap)
}

func (m JSMap) GetList(key string) *JSListStruct {
	var val = m.wrappedMap[key]
	return val.(*JSListStruct)
}

func (m JSMap) OptList(key string) *JSListStruct {
	var val = m.wrappedMap[key]
	if val == nil {
		return nil
	}
	return val.(*JSListStruct)
}

func (m JSMap) OptListOrEmpty(key string) *JSListStruct {
	var val = m.wrappedMap[key]
	if val == nil {
		return NewJSList()
	}
	return val.(*JSListStruct)
}

func (m JSMap) GetString(key string) string {
	var val = m.wrappedMap[key]
	return (val.(JSEntity)).AsString()
}

func (m JSMap) HasKey(key string) bool {
	return HasKey(m.wrappedMap, key)
}

func (m JSMap) OptString(key string, defaultValue string) string {
	var val = m.wrappedMap[key]
	if val == nil {
		return defaultValue
	}
	return (val.(JSEntity)).AsString()
}

func (m JSMap) GetInt32(key string) int32 {
	var val = m.wrappedMap[key]
	return int32((val.(JSEntity)).AsInteger())
}

func (m JSMap) OptInt32(key string, defaultValue int32) int32 {
	var val = m.wrappedMap[key]
	if val == nil {
		return defaultValue
	}
	return int32((val.(JSEntity)).AsInteger())
}

func (m JSMap) OptInt(key string, defaultValue int) int {
	return int(m.OptInt64(key, defaultValue))
}

// Deprecated.. Use OptByte instead.
func (m JSMap) OptInt8(key string, defaultValue int8) int8 {
	var val = m.wrappedMap[key]
	if val == nil {
		return defaultValue
	}
	return int8((val.(JSEntity)).AsInteger())
}

func (m JSMap) OptByte(key string, defaultValue byte) byte {
	var val = m.wrappedMap[key]
	if val == nil {
		return defaultValue
	}
	return byte((val.(JSEntity)).AsInteger())
}

func (m JSMap) OptBytes(key string, defaultValue []byte) []byte {
	var val = m.wrappedMap[key]
	if val == nil {
		return defaultValue
	}
	return DecodeBase64Maybe(val)
}

func (m JSMap) OptFloat32(key string, defaultValue float32) float32 {
	var val = m.wrappedMap[key]
	if val == nil {
		return defaultValue
	}
	return float32((val.(JSEntity)).AsFloat())
}

func (m JSMap) OptFloat64(key string, defaultValue float64) float64 {
	var val = m.wrappedMap[key]
	if val == nil {
		return defaultValue
	}
	return float64((val.(JSEntity)).AsFloat())
}

func (m JSMap) GetInt64(key string) int64 {
	var val = m.wrappedMap[key]
	return int64((val.(JSEntity)).AsInteger())
}
func (m JSMap) OptInt64(key string, defaultValue int) int64 {
	var val = m.wrappedMap[key]
	if val == nil {
		return int64(defaultValue)
	}
	return int64((val.(JSEntity)).AsInteger())
}
func (m JSMap) GetBool(key string) bool {
	var val = m.wrappedMap[key]
	return (val.(JSEntity)).AsBool()
}

func (m JSMap) OptBool(key string, defaultValue bool) bool {
	var val = m.wrappedMap[key]
	if val == nil {
		return defaultValue
	}
	return (val.(JSEntity)).AsBool()
}

// If a key/value pair exists, return the value
func (m JSMap) OptAny(key string) JSEntity {
	return m.wrappedMap[key]
}

func JSMapFromString(content string) (JSMap, error) {
	var p JSONParser
	p.WithText(content)
	return p.ParseMap()
}

func JSMapFromStringM(content string) JSMap {
	return CheckOkWith(JSMapFromString(content))
}

func (m JSMap) WrappedMap() map[string]JSEntity {
	return m.wrappedMap
}

// Get an ordered list of keys for the JSMap
func (m JSMap) OrderedKeys() []string {
	arr := NewArray[string]()
	for k := range m.wrappedMap {
		arr.Add(k)
	}
	arr.Sort()
	return arr.Array()
}

type JSMapEntry struct {
	Key   string
	Value JSEntity
}

// Get an ordered list of entries for the JSMap, sorted by key
func (m JSMap) Entries() []JSMapEntry {
	arr := NewArray[JSMapEntry]()
	for _, k := range m.OrderedKeys() {
		arr.Add(JSMapEntry{
			Key:   k,
			Value: m.wrappedMap[k],
		})
	}
	return arr.Array()
}

func (m JSMap) OptLong(key string, defaultValue int64) int64 {
	var val = m.wrappedMap[key]
	if val == nil {
		return defaultValue
	}
	return JSEntity(val).AsInteger()
}

func (m JSMap) Clear() JSMap {
	m.wrappedMap = make(map[string]JSEntity)
	return m
}
