package json

import (
	_ "strings"

	. "github.com/jpsember/golang-base/base"
)

type JSList struct {
	wrappedList []JSEntity
}

// Construct a JSList from a slice of any, converting to JSEntities
//
//	func JSListWith(values []any) *JSList {
//		var out = NewJSList()
//		for _, x := range values {
//			out.Add(x)
//		}
//		return out
//	}
//
// // Construct a JSList from a slice of strings
//
//	func JSListWithStrings(values []string) *JSList {
//		var out = NewJSList()
//		for _, x := range values {
//			out.Add(x)
//		}
//		return out
//	}
func JSListWith[T any](values []T) *JSList {
	var out = NewJSList()
	for _, x := range values {
		out.Add(x)
	}
	return out
}

// ---------------------------------------------------------------------------------------
// JSEntity interface
// ---------------------------------------------------------------------------------------

func (m *JSList) ToInteger() int64 {
	panic("not supported")
}

func (m *JSList) ToFloat() float64 {
	panic("not supported")
}
func (m *JSList) ToString() string {
	panic("not supported")
}
func (m *JSList) ToBool() bool {
	panic("not supported")
}

// Implements the fmt.Stringer interface.  By default, we perform
// a pretty print of the JSList.  This simplifies a lot of things.
func (m *JSList) String() string {
	return PrintJSEntity(m, false)
}

func (m *JSList) PrintTo(context *JSONPrinter) {
	var s = context.StringBuilder
	s.WriteByte('[')
	var index = 0
	for _, val := range m.wrappedList {
		if index != 0 {
			s.WriteByte(',')
		}
		index++
		val.PrintTo(context)
	}
	s.WriteByte(']')
}

// ---------------------------------------------------------------------------------------

// Factory constructor.  Do *not* construct via JSList().
func NewJSList() *JSList {
	var m = new(JSList)
	m.wrappedList = make([]JSEntity, 0, 10)
	return m
}

func (p *JSONParser) ParseList() (*JSList, error) {
	p.adjustNest(1)
	var result []JSEntity
	p.ReadExpectedByte('[')
	commaExpected := false
	for !p.hasProblem() {
		if p.readIf(']') {
			break
		}
		if commaExpected {
			p.ReadExpectedByte(',')
			if p.readIf(']') {
				break
			}
			commaExpected = false
		}
		elem := p.readValue()
		if p.hasProblem() {
			break
		}
		result = append(result, elem)
		commaExpected = true
	}
	p.skipWhitespace()
	p.adjustNest(-1)
	var jsList *JSList
	if p.Error == nil {
		jsList = new(JSList)
		jsList.wrappedList = result
	}
	return jsList, p.Error
}

func (this *JSList) Add(value any) *JSList {
	// Is 'this' a good convention to use?  Or self (as in Python)?
	CheckNotNil(value)

	this.wrappedList = append(this.wrappedList, ToJSEntity(value))
	return this
}

func (this *JSList) Get(index int) JSEntity {
	return this.wrappedList[index]
}

func (this *JSList) Length() int {
	return len(this.wrappedList)
}
