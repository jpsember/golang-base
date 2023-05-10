package json

import (
	. "js/base"
	_ "strings"
)

type JSList struct {
	wrappedList []JSEntity
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

func (p *JSONParser) ParseList() JSEntity {
	var result []JSEntity
	p.ReadExpectedByte('[')
	var first = true
	for !p.hasProblem() {
		if p.readIf(']') {
			break
		}
		if !first {
			p.ReadExpectedByte(',')
			if p.readIf(']') {
				break
			}
		} else {
			first = false
		}
		var elem = p.readValue()
		if p.hasProblem() {
			break
		}
		result = append(result, elem)
	}
	p.skipWhitespace()

	var jsList = new(JSList)
	jsList.wrappedList = result

	return jsList
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
